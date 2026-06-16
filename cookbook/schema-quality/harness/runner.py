"""Run a single extraction via the authenticated Retab CLI, with local caching.

We shell out to the `retab` CLI instead of the Python SDK because the CLI uses
the user's existing OAuth session (`~/.retab/config.json`) — no API key needs to
be created or handled. The CLI returns the full extraction record as JSON,
including `output` and `consensus.likelihoods`.

Results are cached on disk keyed by (document bytes, canonical schema, model,
n_consensus). Re-running the report therefore costs no credits unless the
inputs change or `--no-cache` is passed.
"""

from __future__ import annotations

import hashlib
import json
import os
import shutil
import subprocess
import tempfile
from typing import Optional

HERE = os.path.dirname(os.path.abspath(__file__))
RESULTS_DIR = os.path.join(os.path.dirname(HERE), "results")

_DEFAULT_CLI_CANDIDATES = [
    os.path.expanduser(os.path.join("~", ".retab", "bin", "retab.exe")),
    os.path.expanduser(os.path.join("~", ".retab", "bin", "retab")),
]


def resolve_cli() -> str:
    """Find the retab CLI: $RETAB_CLI, then PATH, then the default install."""
    env = os.environ.get("RETAB_CLI")
    if env and os.path.exists(env):
        return env
    on_path = shutil.which("retab")
    if on_path:
        return on_path
    for cand in _DEFAULT_CLI_CANDIDATES:
        if os.path.exists(cand):
            return cand
    raise FileNotFoundError(
        "Could not find the retab CLI. Set RETAB_CLI to its full path."
    )


def _canonical(obj) -> str:
    return json.dumps(obj, sort_keys=True, separators=(",", ":"))


def _cache_key(doc_path: str, schema: dict, model: str, n_consensus: int) -> str:
    h = hashlib.sha1()
    with open(doc_path, "rb") as f:
        h.update(f.read())
    h.update(_canonical(schema).encode())
    h.update(f"{model}|{n_consensus}".encode())
    return h.hexdigest()[:16]


def run_extraction(
    doc_path: str,
    schema: dict,
    model: str = "retab-micro",
    n_consensus: int = 5,
    use_cache: bool = True,
    cli: Optional[str] = None,
) -> dict:
    """Extract `doc_path` against `schema`; return the parsed extraction record."""
    os.makedirs(RESULTS_DIR, exist_ok=True)
    key = _cache_key(doc_path, schema, model, n_consensus)
    cache_path = os.path.join(RESULTS_DIR, f"{key}.json")

    if use_cache and os.path.exists(cache_path):
        with open(cache_path, encoding="utf-8") as f:
            record = json.load(f)
        record["_cached"] = True
        return record

    cli = cli or resolve_cli()
    with tempfile.NamedTemporaryFile(
        "w", suffix=".json", delete=False, encoding="utf-8"
    ) as tf:
        json.dump(schema, tf)
        schema_file = tf.name

    try:
        cmd = [
            cli, "extractions", "create",
            "--file", doc_path,
            "--json-schema-file", schema_file,
            "--model", model,
            "--n-consensus", str(n_consensus),
            "--output", "json",
        ]
        proc = subprocess.run(cmd, capture_output=True, text=True)
        if proc.returncode != 0:
            raise RuntimeError(
                f"retab CLI failed (exit {proc.returncode}):\n{proc.stderr.strip()}"
            )
        try:
            record = json.loads(proc.stdout)
        except json.JSONDecodeError as exc:
            raise RuntimeError(
                f"Could not parse CLI output as JSON: {exc}\n--- stdout ---\n"
                f"{proc.stdout[:2000]}"
            ) from exc
    finally:
        os.unlink(schema_file)

    with open(cache_path, "w", encoding="utf-8") as f:
        json.dump(record, f, indent=2)
    record["_cached"] = False
    return record
