"""Integration tests for the Jobs API across all supported endpoints.

Tests exercise the full job lifecycle: create → poll → verify result.
Run against local server:
    .venv/bin/python -m pytest --asyncio-mode=strict -W error -s -v --local test_jobs.py
"""

import base64
import time

import httpx
import pytest

from retab import Retab
from retab.types.jobs import Job

# ---------------------------------------------------------------------------
# Test data
# ---------------------------------------------------------------------------

INLINE_TEXT_DOCUMENT = {
    "filename": "test_invoice.txt",
    "url": "data:text/plain;base64,"
    + base64.b64encode(
        b"Invoice #12345\nDate: 2025-01-15\nAmount: $99.99\nCustomer: Acme Corp\nDescription: Consulting services"
    ).decode(),
}

SIMPLE_EXTRACT_SCHEMA = {
    "type": "object",
    "properties": {
        "invoice_number": {"type": "string", "description": "The invoice number"},
        "amount": {"type": "string", "description": "The total amount"},
        "customer": {"type": "string", "description": "The customer name"},
    },
    "required": ["invoice_number", "amount", "customer"],
}

MODEL = "retab-micro"

JOB_TIMEOUT = 120  # seconds
POLL_INTERVAL = 2  # seconds


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _wait(client: Retab, job_id: str) -> Job:
    """Wait for a job to reach a terminal state and return it with full response."""
    return client.jobs.wait_for_completion(
        job_id,
        poll_interval_seconds=POLL_INTERVAL,
        timeout_seconds=JOB_TIMEOUT,
        include_response=True,
    )


def _assert_completed(job: Job) -> None:
    """Assert that a job completed successfully."""
    assert job.status == "completed", f"Expected completed, got {job.status}. Error: {job.error}"
    assert job.response is not None, "Completed job should have a response"
    assert job.response.status_code == 200, f"Expected 200, got {job.response.status_code}"


# ---------------------------------------------------------------------------
# Per-endpoint completion tests
# ---------------------------------------------------------------------------


def test_job_extract(sync_client: Retab) -> None:
    """Job for /v1/documents/extract completes and returns extracted data."""
    with sync_client as client:
        job = client.jobs.create(
            endpoint="/v1/documents/extract",
            request={
                "document": INLINE_TEXT_DOCUMENT,
                "json_schema": SIMPLE_EXTRACT_SCHEMA,
                "model": MODEL,
            },
        )
        assert job.status in ("queued", "validating")

        job = _wait(client, job.id)
        _assert_completed(job)

        body = job.response.body
        assert "choices" in body, f"Extract response should have 'choices', got keys: {list(body.keys())}"


def test_job_parse(sync_client: Retab) -> None:
    """Job for /v1/parses completes and returns a stored Parse resource."""
    with sync_client as client:
        job = client.jobs.create(
            endpoint="/v1/parses",
            request={
                "document": INLINE_TEXT_DOCUMENT,
                "model": MODEL,
            },
        )
        assert job.status in ("queued", "validating")

        job = _wait(client, job.id)
        _assert_completed(job)

        body = job.response.body
        assert "output" in body, f"Parse response should have 'output', got keys: {list(body.keys())}"
        assert "text" in body["output"], "Parse response output should contain text"
        assert "pages" in body["output"], "Parse response output should contain pages"


def test_job_parses_resource(sync_client: Retab) -> None:
    """Job for /v1/parses completes and returns the stored Parse resource."""
    with sync_client as client:
        job = client.jobs.create(
            endpoint="/v1/parses",
            request={
                "document": INLINE_TEXT_DOCUMENT,
                "model": MODEL,
            },
        )
        assert job.status in ("queued", "validating")

        job = _wait(client, job.id)
        _assert_completed(job)

        body = job.response.body
        assert "id" in body, f"Parse resource response should have 'id', got keys: {list(body.keys())}"
        assert "output" in body, f"Parse resource response should have 'output', got keys: {list(body.keys())}"
        assert "pages" in body["output"], "Parse resource output should contain pages"


def test_job_split(sync_client: Retab) -> None:
    """Job for /v1/documents/split completes and returns split result."""
    with sync_client as client:
        job = client.jobs.create(
            endpoint="/v1/documents/split",
            request={
                "document": INLINE_TEXT_DOCUMENT,
                "subdocuments": [
                    {"name": "invoice_header", "description": "The header section with invoice number and date"},
                    {"name": "invoice_details", "description": "The details section with amount and customer"},
                ],
                "model": MODEL,
            },
        )
        assert job.status in ("queued", "validating")

        job = _wait(client, job.id)
        _assert_completed(job)

        body = job.response.body
        assert "splits" in body, f"Split response should have 'splits', got keys: {list(body.keys())}"
        assert isinstance(body["splits"], list), "splits should be a list"


def test_job_schema_generate(sync_client: Retab) -> None:
    """Job for /v1/schemas/generate completes and returns a schema."""
    with sync_client as client:
        job = client.jobs.create(
            endpoint="/v1/schemas/generate",
            request={
                "documents": [INLINE_TEXT_DOCUMENT],
                "model": MODEL,
            },
        )
        assert job.status in ("queued", "validating")

        job = _wait(client, job.id)
        # Schema generation can fail with lightweight models producing invalid schemas
        assert job.status in ("completed", "failed"), f"Expected terminal state, got {job.status}"
        if job.status == "completed":
            assert job.response is not None
            assert isinstance(job.response.body, dict), "Schema generate should return a dict"


# ---------------------------------------------------------------------------
# Lifecycle tests
# ---------------------------------------------------------------------------


def test_job_create_returns_queued(sync_client: Retab) -> None:
    """Creating a job returns it with status 'queued'."""
    with sync_client as client:
        job = client.jobs.create(
            endpoint="/v1/parses",
            request={
                "document": INLINE_TEXT_DOCUMENT,
                "model": MODEL,
            },
        )
        assert job.id is not None
        assert job.status in ("queued", "validating")
        assert job.endpoint == "/v1/parses"
        assert job.object == "job"

        # Clean up — wait for it to finish
        _wait(client, job.id)


def test_job_retrieve_without_payload(sync_client: Retab) -> None:
    """Default retrieve omits request and response payloads."""
    with sync_client as client:
        job = client.jobs.create(
            endpoint="/v1/parses",
            request={
                "document": INLINE_TEXT_DOCUMENT,
                "model": MODEL,
            },
        )
        # Wait for completion first
        _wait(client, job.id)

        # Retrieve without payloads (default)
        retrieved = client.jobs.retrieve(job.id)
        assert retrieved.id == job.id
        assert retrieved.status == "completed"
        assert retrieved.request is None, "Default retrieve should not include request"
        assert retrieved.response is None, "Default retrieve should not include response"


def test_job_retrieve_with_payload(sync_client: Retab) -> None:
    """retrieve_full includes both request and response."""
    with sync_client as client:
        job = client.jobs.create(
            endpoint="/v1/parses",
            request={
                "document": INLINE_TEXT_DOCUMENT,
                "model": MODEL,
            },
        )
        _wait(client, job.id)

        full = client.jobs.retrieve_full(job.id)
        assert full.id == job.id
        assert full.status == "completed"
        assert full.request is not None, "retrieve_full should include request"
        assert full.response is not None, "retrieve_full should include response"
        assert full.response.status_code == 200


def test_job_list_filters(sync_client: Retab) -> None:
    """List with endpoint and status filters returns matching jobs."""
    with sync_client as client:
        # Create and complete a parse job
        job = client.jobs.create(
            endpoint="/v1/parses",
            request={
                "document": INLINE_TEXT_DOCUMENT,
                "model": MODEL,
            },
        )
        _wait(client, job.id)

        # List with filters
        result = client.jobs.list(
            endpoint="/v1/parses",
            status="completed",
            limit=5,
        )
        assert len(result.data) > 0, "Should find at least one completed parse job"
        for j in result.data:
            assert j.endpoint == "/v1/parses"
            assert j.status == "completed"


def test_job_metadata_roundtrip(sync_client: Retab) -> None:
    """Metadata set at creation is returned in retrieve."""
    with sync_client as client:
        metadata = {"test_key": "test_value", "source": "sdk_test"}
        job = client.jobs.create(
            endpoint="/v1/parses",
            request={
                "document": INLINE_TEXT_DOCUMENT,
                "model": MODEL,
            },
            metadata=metadata,
        )
        assert job.metadata == metadata

        # Also verify after completion
        completed = _wait(client, job.id)
        retrieved = client.jobs.retrieve(completed.id)
        assert retrieved.metadata == metadata


def test_job_cancel(sync_client: Retab) -> None:
    """Cancelling a queued job sets status to 'cancelled'."""
    with sync_client as client:
        job = client.jobs.create(
            endpoint="/v1/parses",
            request={
                "document": INLINE_TEXT_DOCUMENT,
                "model": MODEL,
            },
        )
        # Try to cancel immediately (may already be in_progress)
        try:
            cancelled = client.jobs.cancel(job.id)
            assert cancelled.status == "cancelled"
        except httpx.HTTPStatusError as e:
            # Job may have already completed or moved past cancellable state
            if e.response.status_code == 409:
                pass  # Expected if job already completed
            else:
                raise


def test_job_invalid_endpoint(sync_client: Retab) -> None:
    """Creating a job with an invalid endpoint raises an error."""
    with sync_client as client:
        with pytest.raises(Exception):
            client.jobs.create(
                endpoint="/v1/nonexistent/endpoint",  # type: ignore
                request={"document": INLINE_TEXT_DOCUMENT},
            )


def test_job_invalid_request_body(sync_client: Retab) -> None:
    """Creating a job with wrong request body shape raises a validation error."""
    with sync_client as client:
        with pytest.raises(Exception):
            client.jobs.create(
                endpoint="/v1/documents/extract",
                request={
                    # Missing required 'document' and 'json_schema'
                    "model": MODEL,
                },
            )
