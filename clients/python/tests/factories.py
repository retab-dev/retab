"""Reusable e2e setup factories — the create/upload/discovery paths the tests
take over and over, turned into one place.

Everything here is CREDITLESS by construction: it only uploads bytes to storage
(no model is ever invoked), creates/deletes config *definitions* (workflows,
secrets, tables — running them would be billable and is never done here), or
discovers existing resources via ``list``. None of it triggers extraction,
parsing, OCR, classification, or any LLM call.

Two ways to use it:
  * import the functions directly (``from factories import upload_file``), or
  * use the conftest fixtures that wrap them with setup/teardown
    (``project_id``, ``uploaded_file``, ``temp_workflow``).

The context managers (``temporary_*``) create a resource and always delete it on
exit, so tests stay self-cleaning. Cleanup is best-effort (suppressed) so a
teardown failure never masks the test's own assertion.
"""

from __future__ import annotations

import contextlib
import hashlib
import uuid
from typing import Any, AsyncIterator, Iterator

import httpx
from pydantic import ValidationError as PydanticValidationError

from retab import AsyncRetab, Retab
from retab.types.files import CreateUploadResponse, File
from retab.types.secrets import Secret
from retab.types.tables import WorkflowTable
from retab.types.workflows import Workflow

# Tiny, obviously-test payloads. Uploading bytes is pure storage and never
# triggers processing, so these stay creditless.
TINY_FILE_CONTENT = b"creditless storage test fixture\n"
TINY_CSV = b"name,amount\nalice,10\nbob,20\n"


def unique_name(prefix: str = "creditless", suffix: str = "") -> str:
    """A collision-resistant, clearly-tagged name for test-created resources."""
    base = f"{prefix}-{uuid.uuid4().hex[:12]}"
    return f"{base}.{suffix}" if suffix else base


# --------------------------------------------------------------------------- #
# Files — the create_upload -> signed PUT -> complete_upload dance
# --------------------------------------------------------------------------- #


def upload_file(
    client: Retab,
    content: bytes = TINY_FILE_CONTENT,
    *,
    filename: str | None = None,
    content_type: str = "text/plain",
) -> File:
    """Run the full storage upload flow and return the stored ``File`` metadata.

    Storage-only and creditless: create an upload session, PUT the bytes to the
    signed URL with the server-provided headers, complete the upload, and read
    the file back. No ``files.delete`` exists in the SDK, so the bytes persist —
    keep ``content`` tiny.
    """
    filename = filename or unique_name("creditless_storage", "txt")
    sha = hashlib.sha256(content).hexdigest()

    session = client.files.create_upload(
        filename=filename,
        size_bytes=len(content),
        content_type=content_type,
        sha256=sha,
    )
    if not isinstance(session, CreateUploadResponse) or not session.file_id or not session.upload_url:
        raise AssertionError(f"create_upload returned an unusable session: {session!r}")

    _put_bytes(session, content)
    client.files.complete_upload(session.file_id, sha256=sha)
    return client.files.get(session.file_id)


async def upload_file_async(
    client: AsyncRetab,
    content: bytes = TINY_FILE_CONTENT,
    *,
    filename: str | None = None,
    content_type: str = "text/plain",
) -> File:
    """Async counterpart of :func:`upload_file`."""
    filename = filename or unique_name("creditless_storage", "txt")
    sha = hashlib.sha256(content).hexdigest()

    session = await client.files.create_upload(
        filename=filename,
        size_bytes=len(content),
        content_type=content_type,
        sha256=sha,
    )
    if not isinstance(session, CreateUploadResponse) or not session.file_id or not session.upload_url:
        raise AssertionError(f"create_upload returned an unusable session: {session!r}")

    _put_bytes(session, content)
    await client.files.complete_upload(session.file_id, sha256=sha)
    return await client.files.get(session.file_id)


def _put_bytes(session: CreateUploadResponse, content: bytes) -> None:
    """PUT bytes to the signed URL, echoing the server's headers exactly.

    The signed PUT is signature-sensitive: the request must send back the
    server-provided ``upload_headers`` verbatim (adding/overriding headers breaks
    the signature -> 403). Kept here so every caller gets it right.
    """
    put = httpx.request(
        session.upload_method or "PUT",
        session.upload_url,
        content=content,
        headers=session.upload_headers or {},
    )
    if put.status_code not in (200, 201):
        raise AssertionError(f"signed PUT failed: {put.status_code} {put.text}")


# --------------------------------------------------------------------------- #
# Projects — the SDK has no projects resource, so discover via existing data
# --------------------------------------------------------------------------- #


def discover_project_id(client: Retab) -> str | None:
    """Return an existing ``project_id`` (every workflow carries its owner).

    Reusing an existing project keeps resource creation creditless and avoids
    org-level data we cannot clean up. Returns ``None`` if the org has none.
    """
    page = client.workflows.list(limit=25)
    for wf in page.data:
        pid = getattr(wf, "project_id", None)
        if pid:
            return pid
    return None


async def discover_project_id_async(client: AsyncRetab) -> str | None:
    """Async counterpart of :func:`discover_project_id`."""
    page = await client.workflows.list(limit=25)
    for wf in page.data:
        pid = getattr(wf, "project_id", None)
        if pid:
            return pid
    return None


# --------------------------------------------------------------------------- #
# Workflow definitions (creditless config — creating a definition does NOT run)
# --------------------------------------------------------------------------- #


def create_workflow(
    client: Retab,
    project_id: str,
    *,
    name: str | None = None,
    description: str = "",
) -> Workflow:
    """Create a workflow definition under ``project_id``."""
    return client.workflows.create(
        project_id=project_id,
        name=name or unique_name("creditless-wf"),
        description=description,
    )


@contextlib.contextmanager
def temporary_workflow(
    client: Retab,
    project_id: str,
    *,
    name: str | None = None,
    description: str = "",
) -> Iterator[Workflow]:
    """Create a workflow definition and delete it on exit."""
    workflow = create_workflow(client, project_id, name=name, description=description)
    try:
        yield workflow
    finally:
        with contextlib.suppress(Exception):
            client.workflows.delete(workflow.id)


@contextlib.asynccontextmanager
async def temporary_workflow_async(
    client: AsyncRetab,
    project_id: str,
    *,
    name: str | None = None,
    description: str = "",
) -> AsyncIterator[Workflow]:
    """Async counterpart of :func:`temporary_workflow`."""
    workflow = await client.workflows.create(
        project_id=project_id,
        name=name or unique_name("creditless-wf"),
        description=description,
    )
    try:
        yield workflow
    finally:
        with contextlib.suppress(Exception):
            await client.workflows.delete(workflow.id)


# --------------------------------------------------------------------------- #
# Secrets (creditless config)
# --------------------------------------------------------------------------- #


@contextlib.contextmanager
def temporary_secret(
    client: Retab,
    *,
    name: str | None = None,
    value: str = "initial-value",
) -> Iterator[Secret]:
    """Create a secret and delete it on exit; yields the created ``Secret``.

    Secret names are restricted to ``^[A-Za-z_][A-Za-z0-9_]*$`` (no dashes), so
    the default name is underscore-only rather than the dashed ``unique_name``.
    """
    name = name or f"creditless_secret_{uuid.uuid4().hex[:12]}"
    created = client.secrets.create_secret(name=name, value=value)
    try:
        yield created.secret
    finally:
        with contextlib.suppress(Exception):
            client.secrets.delete_secret(name)


# --------------------------------------------------------------------------- #
# Tables (creditless config — a CSV upload that becomes a table definition)
# --------------------------------------------------------------------------- #


@contextlib.contextmanager
def temporary_table(
    client: Retab,
    project_id: str,
    *,
    name: str | None = None,
    file: bytes = TINY_CSV,
) -> Iterator[WorkflowTable]:
    """Create a table from ``file`` and delete it on exit; yields the table."""
    name = name or unique_name("creditless-table")
    created = client.tables.create(name=name, file=file, project_id=project_id)
    if not created.tables:
        raise AssertionError("tables.create returned no table")
    table = created.tables[0]
    try:
        yield table
    finally:
        with contextlib.suppress(Exception):
            client.tables.delete(table.id)


# --------------------------------------------------------------------------- #
# List inspection — read the raw envelope / validate tolerantly
# --------------------------------------------------------------------------- #


def raw_list(client: Retab, resource_name: str, **list_params: Any) -> dict[str, Any]:
    """Return a list resource's raw envelope, bypassing per-item model validation.

    Lets pagination / filtering be asserted even when legacy staging rows would
    fail typed validation. Use the model-validating SDK ``list`` when you want
    typed items; use this when you want the wire shape.
    """
    resource = getattr(client, resource_name)
    prepared = resource.prepare_list(**list_params)
    return resource._client._prepared_request(prepared)


async def raw_list_async(client: AsyncRetab, resource_name: str, **list_params: Any) -> dict[str, Any]:
    """Async counterpart of :func:`raw_list`."""
    resource = getattr(client, resource_name)
    prepared = resource.prepare_list(**list_params)
    return await resource._client._prepared_request(prepared)


def raw_ids(envelope: dict[str, Any]) -> list[str]:
    """Extract the string ``id`` of every row in a raw list envelope."""
    out: list[str] = []
    for item in envelope.get("data") or []:
        if isinstance(item, dict) and isinstance(item.get("id"), str):
            out.append(item["id"])
    return out


def validate_tolerant(envelope: dict[str, Any], model: type) -> tuple[list[Any], int]:
    """Validate page items one at a time; return ``(validated, skipped_legacy)``.

    Staging holds legacy rows that violate the current public contract; validating
    per-item lets a test assert the happy path while skipping (and counting) the
    rows that fail typed validation instead of failing the whole page.
    """
    validated: list[Any] = []
    skipped = 0
    for item in envelope.get("data") or []:
        if not isinstance(item, dict):
            continue
        try:
            validated.append(model.model_validate(item))
        except PydanticValidationError:
            skipped += 1
    return validated, skipped


# --------------------------------------------------------------------------- #
# Auth-failure clients — a junk API key for 401/permission assertions
# --------------------------------------------------------------------------- #

# A clearly-invalid key shared by every auth-failure test.
JUNK_API_KEY = "sk_junk_invalid_creditless"


def junk_key_client(base_url: str, *, key: str = JUNK_API_KEY) -> Retab:
    """A sync client with an invalid key (no retries) for 401/permission tests.

    The caller owns closing it; prefer the ``bad_key_client`` conftest fixture,
    which handles teardown.
    """
    return Retab(api_key=key, base_url=base_url, max_retries=0)


def junk_key_async_client(base_url: str, *, key: str = JUNK_API_KEY) -> AsyncRetab:
    """Async counterpart of :func:`junk_key_client`."""
    return AsyncRetab(api_key=key, base_url=base_url, max_retries=0)
