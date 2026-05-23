from __future__ import annotations

import ast
import json
import re
from pathlib import Path


SDK_ROOT = Path(__file__).resolve().parents[1]
REPO_ROOT = SDK_ROOT.parents[3]
SDK_RESOURCES = SDK_ROOT / "retab" / "resources"
WORKFLOW_RESOURCES = SDK_RESOURCES / "workflows"
PUBLIC_OPENAPI = REPO_ROOT / "open-source" / "docs" / "api-reference" / "openapi.json"

HTTP_METHODS = {"get", "post", "patch", "delete", "put"}
REMOVED_METHOD_NAMES = {
    "append_version",
    "prepare_append_version",
    "execute",
    "prepare_execute",
}
REMOVED_ROUTE_FRAGMENTS = {
    "/v1/workflows/reviews/{id}",
    "/v1/workflows/runs/{run_id}/steps",
    "/v1/workflows/runs/{run_id}/steps/{step_id}",
    "/v1/workflows/blocks/{block_id}/execute",
    "/v1/workflows/blocks/execute",
}
APPROVED_NON_REFERENCE_ROUTES: set[tuple[str, str]] = set()


def _stringish_ast_value(node: ast.AST) -> str | None:
    if isinstance(node, ast.Constant) and isinstance(node.value, str):
        return node.value
    if isinstance(node, ast.JoinedStr):
        parts: list[str] = []
        for value in node.values:
            if isinstance(value, ast.Constant) and isinstance(value.value, str):
                parts.append(value.value)
            elif isinstance(value, ast.FormattedValue):
                parts.append("{}")
            else:
                return None
        return "".join(parts)
    return None


def _normalized_path(path: str) -> str:
    path_without_query = path.split("?", 1)[0]
    return re.sub(r"{[^}]+}", "{}", path_without_query)


def _workflow_openapi_operations() -> set[tuple[str, str]]:
    return {operation for operation in _openapi_operations() if operation[1].startswith("/v1/workflows")}


def _openapi_operations() -> set[tuple[str, str]]:
    spec = json.loads(PUBLIC_OPENAPI.read_text())
    operations: set[tuple[str, str]] = set()
    for path, path_item in spec["paths"].items():
        sdk_path = path.removeprefix("/v1")
        for method in path_item:
            if method in HTTP_METHODS:
                operations.add((method.upper(), _normalized_path(f"/v1{sdk_path}")))
    return operations


def _python_sdk_workflow_operations() -> set[tuple[str, str]]:
    return {operation for operation in _python_sdk_operations(WORKFLOW_RESOURCES) if operation[1].startswith("/v1/workflows")}


def _python_sdk_operations(root: Path) -> set[tuple[str, str]]:
    operations: set[tuple[str, str]] = set()
    for source_file in root.rglob("*.py"):
        module = ast.parse(source_file.read_text())
        for node in ast.walk(module):
            if not isinstance(node, ast.Call):
                continue
            if getattr(node.func, "id", None) != "PreparedRequest":
                continue
            method: str | None = None
            url: str | None = None
            for keyword in node.keywords:
                if keyword.arg == "method":
                    method = _stringish_ast_value(keyword.value)
                elif keyword.arg == "url":
                    url = _stringish_ast_value(keyword.value)
            if method is not None and url is not None:
                normalized_url = url if url.startswith("/v1") else f"/v1{url}"
                operations.add((method, _normalized_path(normalized_url)))
    return operations


def test_python_sdk_routes_are_public_openapi_routes_or_explicit_exceptions() -> None:
    openapi_operations = _openapi_operations()
    sdk_operations = _python_sdk_operations(SDK_RESOURCES)

    assert sdk_operations - openapi_operations == APPROVED_NON_REFERENCE_ROUTES


def test_python_workflow_sdk_routes_exist_in_public_openapi() -> None:
    openapi_operations = _workflow_openapi_operations()
    sdk_operations = _python_sdk_workflow_operations()

    assert sdk_operations - openapi_operations == set()


def test_python_workflow_sdk_covers_public_openapi_workflow_routes() -> None:
    openapi_operations = _workflow_openapi_operations()
    sdk_operations = _python_sdk_workflow_operations()

    assert openapi_operations - sdk_operations == set()


def test_python_workflow_sdk_does_not_reintroduce_removed_method_names() -> None:
    offenders: dict[str, list[str]] = {}

    for source_file in WORKFLOW_RESOURCES.rglob("*.py"):
        module = ast.parse(source_file.read_text())
        removed_names = sorted({node.name for node in ast.walk(module) if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)) and node.name in REMOVED_METHOD_NAMES})
        if removed_names:
            offenders[source_file.relative_to(SDK_ROOT).as_posix()] = removed_names

    assert offenders == {}


def test_python_workflow_sdk_does_not_call_removed_routes() -> None:
    offenders: dict[str, list[str]] = {}

    for source_file in WORKFLOW_RESOURCES.rglob("*.py"):
        text = source_file.read_text()
        matched_fragments = sorted(fragment for fragment in REMOVED_ROUTE_FRAGMENTS if fragment in text)
        if matched_fragments:
            offenders[source_file.relative_to(SDK_ROOT).as_posix()] = matched_fragments

    assert offenders == {}


def test_python_workflow_review_versions_require_parent_id() -> None:
    source = (SDK_ROOT / "retab" / "resources" / "workflows" / "reviews.py").read_text()

    assert "parent_id: str," in source
    assert 'raise ValueError("parent_id is required' in source
    assert "append_version" not in source
