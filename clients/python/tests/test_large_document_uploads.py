import base64
import json

import httpx
import pytest

from retab import AsyncRetab, Retab
from retab.exceptions import APIError
from retab.utils import uploads


pytestmark = pytest.mark.unit


def document(size: int) -> dict[str, str]:
    return {"filename": "large.pdf", "url": "data:application/pdf;base64," + base64.b64encode(b"A" * size).decode()}


def storage_transport(calls: list[httpx.Request], *, fail_upload: bool = False, retry_submission: bool = False) -> httpx.MockTransport:
    submissions = 0

    def handle(request: httpx.Request) -> httpx.Response:
        nonlocal submissions
        calls.append(request)
        if request.url.host == "upload.example":
            assert "Authorization" not in request.headers
            return httpx.Response(503 if fail_upload else 200)
        assert request.headers["Authorization"] == "Bearer test_key"
        if request.url.path == "/v1/files/upload":
            body = json.loads(request.content)
            assert body["content_type"] == "application/pdf"
            return httpx.Response(
                200,
                json={
                    "fileId": "file_large",
                    "uploadUrl": "https://upload.example/file",
                    "uploadMethod": "PUT",
                    "expiresAt": "2099-01-01T00:00:00Z",
                },
            )
        if request.url.path == "/v1/files/upload/file_large/complete":
            return httpx.Response(200, json={"filename": "large.pdf", "url": "https://storage.retab.com/file_large"})
        assert len(request.content) < 32 * 1024 * 1024
        submissions += 1
        if retry_submission and submissions == 1:
            return httpx.Response(500, json={"detail": "retry"})
        return httpx.Response(200, json={"id": "run_1"})

    return httpx.MockTransport(handle)


def test_large_document_upload_precedes_submission_and_survives_retry() -> None:
    calls: list[httpx.Request] = []
    body = {"workflow_id": "wf_1", "documents": {"start": document(25 * 1024 * 1024)}}
    original_url = body["documents"]["start"]["url"]
    with Retab(api_key="test_key", max_retries=1) as client:
        client.client.close()
        client.client = httpx.Client(transport=storage_transport(calls, retry_submission=True))
        assert client._request("POST", "/v1/workflows/runs", data=body) == {"id": "run_1"}
    assert [request.url.path for request in calls] == [
        "/v1/files/upload",
        "/file",
        "/v1/files/upload/file_large/complete",
        "/v1/workflows/runs",
        "/v1/workflows/runs",
    ]
    assert len(calls[1].content) == 25 * 1024 * 1024
    assert json.loads(calls[-1].content)["documents"]["start"]["url"] == "https://storage.retab.com/file_large"
    assert body["documents"]["start"]["url"] == original_url


@pytest.mark.asyncio
@pytest.mark.parametrize("as_array", [False, True])
async def test_async_upload_uses_combined_request_budget(monkeypatch: pytest.MonkeyPatch, as_array: bool) -> None:
    monkeypatch.setattr(uploads, "MAX_DOCUMENT_REQUEST_BYTES", 1000)
    calls: list[httpx.Request] = []
    first, last = document(500), document(500)
    remote = {"filename": "remote.pdf", "url": "https://example.com/remote.pdf"}
    body = {"documents": [first, remote, last] if as_array else {"one": first, "two": last}}
    original = json.dumps(body)
    endpoint = "/v1/schemas/generate" if as_array else "/v1/workflows/runs"
    async with AsyncRetab(api_key="test_key") as client:
        await client.client.aclose()
        client.client = httpx.AsyncClient(transport=storage_transport(calls))
        assert await client._request("POST", endpoint, data=body) == {"id": "run_1"}
    assert len([request for request in calls if request.url.host == "upload.example"]) == 2
    submitted = json.loads(calls[-1].content)["documents"]
    if as_array:
        assert isinstance(submitted, list)
        assert submitted[1] == remote
        assert submitted[0]["url"] == submitted[2]["url"] == "https://storage.retab.com/file_large"
    else:
        assert isinstance(submitted, dict)
    assert json.dumps(body) == original


def test_small_documents_and_unrelated_json_are_preserved() -> None:
    body = {"document": document(64), "json_inputs": {"example": document(64)}}
    calls: list[httpx.Request] = []
    with Retab(api_key="test_key") as client:
        client.client.close()
        client.client = httpx.Client(transport=storage_transport(calls))
        client._request("POST", "/v1/parses", data=body)
    assert len(calls) == 1
    assert json.loads(calls[0].content) == body


@pytest.mark.asyncio
async def test_failed_upload_never_finalizes_or_submits(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr(uploads, "MAX_INLINE_DOCUMENT_BYTES", 32)
    calls: list[httpx.Request] = []
    async with AsyncRetab(api_key="test_key") as client:
        await client.client.aclose()
        client.client = httpx.AsyncClient(transport=storage_transport(calls, fail_upload=True))
        with pytest.raises(APIError, match="Direct document upload failed"):
            await client._request("POST", "/v1/workflows/runs", data={"documents": {"start": document(64)}})
    assert len(calls) == 2
