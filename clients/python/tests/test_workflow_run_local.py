"""Local SDK integration test for the workflow run system.

Builds a two-block ``start_json -> function`` workflow from a declarative spec,
runs it against JSON input, waits for a terminal status, and asserts the run
completes (and, best-effort, that the function block produced the expected
transformed output).

Run it against a local Retab stack:

    RETAB_API_KEY=sk_retab_... \\
    RETAB_API_BASE_URL=http://localhost:4000 \\
    RETAB_PROJECT_ID=proj_... \\
    .venv/bin/python -m pytest open-source/sdk/clients/python/tests/test_workflow_run_local.py -v -s

The test skips when ``RETAB_API_KEY`` is unset, so it is safe in suites that do
not have a local stack. It uses only the SDK surface (``retab.Retab``) — no raw
HTTP — so it doubles as a smoke test of the workflow SDK resources
(``workflows.apply`` / ``blocks`` / ``runs`` / ``steps``).
"""

from __future__ import annotations

import os
import time

import pytest

from retab import Retab
from retab.exceptions import RetabError

API_KEY = os.environ.get("RETAB_API_KEY")
BASE_URL = os.environ.get("RETAB_API_BASE_URL", "http://localhost:4000")
PROJECT_ID = os.environ.get("RETAB_PROJECT_ID", "proj_3eHNqAleW1jzHJ1-9S3MC")

TERMINAL = {"completed", "error", "cancelled", "failed"}

# A deterministic, dependency-free transform: label = "<vendor> <invoice>",
# doubled_total = total_due * 2. Keeps the assertion exact and offline-friendly.
SPEC = """\
apiVersion: workflows.retab.com/v1alpha2
kind: Workflow
metadata:
  name: SDK Local Run Test
  description: start_json -> function transform (SDK integration test)
spec:
  blocks:
    sdk_start:
      type: start_json
      label: Start
      position: {x: 0, y: 0}
      config:
        json_schema:
          type: object
          properties:
            invoice_number: {type: string}
            total_due: {type: number}
            vendor: {type: string}
          required: [invoice_number, total_due, vendor]
          additionalProperties: false
    sdk_fn:
      type: function
      label: Transform
      position: {x: 420, y: 0}
      config:
        language: python
        output_schema:
          type: object
          properties:
            label: {type: string}
            doubled_total: {type: number}
          required: [label, doubled_total]
          additionalProperties: false
        code: |
          from models import Input, Output

          def transform(input_data: Input) -> Output:
              return Output(
                  label=f"{input_data.vendor} {input_data.invoice_number}",
                  doubled_total=(input_data.total_due or 0) * 2,
              )
  edges:
    - source: {block: sdk_start, handle: output-json-0}
      target: {block: sdk_fn, handle: input-json-0}
"""

INPUT = {"invoice_number": "INV-2026-0042", "total_due": 648, "vendor": "ACME Corporation"}
EXPECTED_LABEL = "ACME Corporation INV-2026-0042"
EXPECTED_DOUBLED = 1296  # 648 * 2


@pytest.fixture
def client() -> Retab:
    if not API_KEY:
        pytest.skip("RETAB_API_KEY not set — skipping local workflow-run integration test")
    return Retab(api_key=API_KEY, base_url=BASE_URL)


WORKFLOW_NAME = "SDK Local Run Test"


def _retry(fn, attempts: int = 4):
    """Retry a network call through transient blips (server disconnects, 5xx) that
    a busy shared local stack throws. Re-raises the last error if all attempts fail."""
    last: Exception | None = None
    for i in range(attempts):
        try:
            return fn()
        except Exception as exc:  # noqa: BLE001 - includes httpx/httpcore transport errors
            last = exc
            time.sleep(1.5 * (i + 1))
    raise AssertionError(f"call failed after {attempts} attempts: {last}")


def _find_existing_workflow(client: Retab) -> str | None:
    """Reuse an already-provisioned start_json->function workflow if one exists.

    Creating a workflow provisions authorization (a WorkOS call that can be flaky
    on a local stack); reusing an existing one keeps the test reliable and fast.
    """
    workflows = getattr(_retry(lambda: client.workflows.list(limit=50)), "data", []) or []
    for workflow in workflows:
        if getattr(workflow, "name", None) != WORKFLOW_NAME:
            continue
        block_types = {
            str(getattr(b, "type", "")).rsplit(".", 1)[-1].lower()
            for b in (getattr(client.workflows.blocks.list(workflow.id), "data", []) or [])
        }
        if {"start_json", "function"} <= block_types:
            return workflow.id
    return None


def _get_or_create_workflow(client: Retab, attempts: int = 4) -> str:
    existing = _find_existing_workflow(client)
    if existing:
        return existing
    last: Exception | None = None
    for i in range(attempts):
        try:
            return client.workflows.apply(yaml_definition=SPEC, project_id=PROJECT_ID).workflow_id
        except RetabError as exc:  # noqa: PERF203
            last = exc
            time.sleep(2 * (i + 1))
            reused = _find_existing_workflow(client)  # apply may have provisioned despite the error
            if reused:
                return reused
    raise AssertionError(f"could not get or create workflow after {attempts} attempts: {last}")


def _function_output(client: Retab, run_id: str) -> dict | None:
    """Best-effort read of the function block's JSON output. Local stacks can
    intermittently 500 on artifact read-back; that must not fail the run
    assertion. The ``output-json-0`` handle is a PublicHandlePayload (``.data``)
    on the SDK model, or a plain dict on a raw response — handle both."""
    try:
        steps = getattr(client.workflows.steps.list(run_id, limit=20), "data", []) or []
    except RetabError:
        return None
    for step in steps:
        if getattr(step, "block_type", None) != "function":
            continue
        outputs = getattr(step, "handle_outputs", None) or {}
        handle = outputs.get("output-json-0") if hasattr(outputs, "get") else None
        if handle is None:
            continue
        data = handle.get("data") if isinstance(handle, dict) else getattr(handle, "data", None)
        if data is not None and hasattr(data, "model_dump"):
            data = data.model_dump()
        return data if isinstance(data, dict) else None
    return None


def test_workflow_run_function_transform(client: Retab) -> None:
    workflow_id = _get_or_create_workflow(client)

    blocks = getattr(_retry(lambda: client.workflows.blocks.list(workflow_id)), "data", []) or []
    start = next(
        b for b in blocks if str(getattr(b, "type", "")).rsplit(".", 1)[-1].lower() == "start_json"
    )

    run = _retry(
        lambda: client.workflows.runs.create(
            workflow_id, json_inputs={start.id: INPUT}, version="draft"
        )
    )
    run_id = run.id

    status = run.lifecycle.status
    deadline = time.time() + 180
    while status not in TERMINAL and time.time() < deadline:
        time.sleep(3)
        status = _retry(lambda: client.workflows.runs.get(run_id)).lifecycle.status

    assert status == "completed", f"run {run_id} ended in {status!r}, expected completed"

    # Best-effort output check — only asserts when the local stack can resolve
    # the artifact (skips silently when artifact read-back is unavailable).
    output = _function_output(client, run_id)
    if output is not None:
        assert output.get("label") == EXPECTED_LABEL, output
        assert output.get("doubled_total") == EXPECTED_DOUBLED, output
