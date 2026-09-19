#!/usr/bin/env python3
"""Keep manifest-owned PHP output present and eligible for ordinary Git staging."""

import json
import os
from pathlib import Path, PurePosixPath
import subprocess
import sys
import tempfile


def main() -> None:
    sdk_ignore = Path(sys.argv[1]).absolute()
    manifest = Path(sys.argv[2]).absolute()
    sdk_root = sdk_ignore.parent
    php_root = manifest.parent
    paths = json.loads(manifest.read_text())["files"]
    if not isinstance(paths, list) or not paths:
        raise SystemExit("PHP generated manifest must contain files")
    if len(paths) != len(set(paths)):
        raise SystemExit("PHP generated manifest contains duplicate files")
    for path in paths:
        if (
            not isinstance(path, str)
            or not PurePosixPath(path).parts
            or PurePosixPath(path).is_absolute()
            or ".." in PurePosixPath(path).parts
            or any(character in path for character in "\r\n\0")
        ):
            raise SystemExit(f"Unsafe PHP generated manifest path: {path!r}")
    missing = [path for path in paths if not (php_root / path).is_file()]
    candidates = [str((php_root / path).relative_to(sdk_root)) for path in paths]
    # A synthetic empty Git metadata directory checks committed ignore rules
    # even for already-tracked local files. Copy only ignore rules as ordinary
    # files: Git deliberately does not follow Bazel runfile .gitignore symlinks.
    # No real repository index, config, refs, or source content is changed.
    environment = {
        "PATH": os.defpath,
        "GIT_CONFIG_NOSYSTEM": "1",
        "GIT_CONFIG_GLOBAL": os.devnull,
    }
    with tempfile.TemporaryDirectory(prefix="php-generated-tracking-") as temporary:
        metadata = Path(temporary) / "git"
        rules_root = Path(temporary) / "rules"
        rules_root.mkdir()
        ignore_paths = {sdk_ignore, php_root / ".gitignore"}
        for path in paths:
            parent = (php_root / path).parent
            while parent != php_root:
                if (parent / ".gitignore").is_file():
                    ignore_paths.add(parent / ".gitignore")
                parent = parent.parent
        for ignore in ignore_paths:
            destination = rules_root / ignore.relative_to(sdk_root)
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_bytes(ignore.read_bytes())
        subprocess.run(
            ["git", "init", "--bare", "--quiet", "--template=", str(metadata)],
            check=True,
            env=environment,
        )
        ignored = subprocess.run(
            [
                "git",
                "-c",
                f"core.excludesFile={os.devnull}",
                f"--git-dir={metadata}",
                f"--work-tree={rules_root}",
                "check-ignore",
                "--no-index",
                "--stdin",
            ],
            cwd=rules_root,
            input="\n".join(candidates) + "\n",
            text=True,
            capture_output=True,
            env=environment,
        )
        if ignored.returncode not in (0, 1):
            raise SystemExit(ignored.stderr)
    failures = []
    if missing:
        failures.append("Missing manifest-owned PHP files:\n" + "\n".join(missing))
    if ignored.stdout:
        failures.append("Ignored manifest-owned PHP files:\n" + ignored.stdout.rstrip())
    if failures:
        raise SystemExit("\n".join(failures))
    print(f"All {len(paths)} PHP manifest files exist and are not ignored")


if __name__ == "__main__":
    main()
