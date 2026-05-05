import ast
import shutil
import subprocess
from pathlib import Path


def test_sdk_sources_parse_with_python311_grammar():
    sdk_root = Path(__file__).resolve().parents[1] / "retab"

    syntax_errors: list[str] = []
    for source_path in sorted(sdk_root.rglob("*.py")):
        relative_path = source_path.relative_to(sdk_root.parent)
        try:
            ast.parse(source_path.read_text(), filename=str(relative_path), feature_version=(3, 11))
        except SyntaxError as exc:
            syntax_errors.append(f"{relative_path}:{exc.lineno}: {exc.msg}")

    assert syntax_errors == []


def test_sdk_sources_compile_with_python311():
    sdk_root = Path(__file__).resolve().parents[1] / "retab"

    if shutil.which("python3.11"):
        command = ["python3.11", "-m", "compileall", "-q", str(sdk_root)]
    elif shutil.which("uv"):
        command = ["uv", "run", "--isolated", "--no-project", "--python", "3.11", "python", "-m", "compileall", "-q", str(sdk_root)]
    else:
        raise AssertionError("Python 3.11 compatibility test requires python3.11 or uv")

    result = subprocess.run(command, cwd=sdk_root.parent, capture_output=True, text=True)

    assert result.returncode == 0, result.stdout + result.stderr


def test_sdk_imports_with_python311():
    sdk_root = Path(__file__).resolve().parents[1]

    if shutil.which("uv") is None:
        raise AssertionError("Python 3.11 import test requires uv")

    command = [
        "uv",
        "run",
        "--isolated",
        "--no-project",
        "--python",
        "3.11",
        "--with-editable",
        str(sdk_root),
        "python",
        "-c",
        "from retab import AsyncRetab, Retab; print(Retab.__name__, AsyncRetab.__name__)",
    ]
    result = subprocess.run(command, cwd=sdk_root, capture_output=True, text=True)

    assert result.returncode == 0, result.stdout + result.stderr
