import inspect
from io import BytesIO

import pytest
import httpx

from retab.resources.extractions.client import AsyncExtractions, Extractions
from retab import AsyncRetab, Retab
from retab.types.extractions import ExtractionRequest
from retab.types.mime import MIMEData
from retab.types.partitions import Partition
from retab.types.splits import Split


def _sample_document() -> MIMEData:
    return MIMEData(
        filename="invoice.txt",
        url="data:text/plain;base64,aW52b2ljZQ==",
    )


def test_extractions_delete_uses_new_resource_route(monkeypatch: pytest.MonkeyPatch) -> None:
    captured: dict[str, object] = {}

    def fake_prepared_request(request: object) -> dict[str, object]:
        captured["request"] = request
        return {}

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        client.extractions.delete("extr_123")

    assert getattr(captured["request"], "url") == "/extractions/extr_123"


def test_extractions_create_has_no_stream_argument() -> None:
    assert "stream" not in inspect.signature(Extractions.create).parameters
    assert "stream" not in inspect.signature(AsyncExtractions.create).parameters
    assert "stream" not in inspect.signature(Extractions.prepare_create).parameters
    assert "stream" not in ExtractionRequest.model_fields

    with pytest.raises(TypeError, match="unexpected keyword argument 'stream'"):
        Extractions(object()).create(
            document=_sample_document(),
            json_schema={"type": "object"},
            stream=True,
        )

    with pytest.raises(TypeError, match="unexpected keyword argument 'stream'"):
        AsyncExtractions(object()).create(
            document=_sample_document(),
            json_schema={"type": "object"},
            stream=True,
        )


def test_files_upload_rejects_signed_bucket_url() -> None:
    signed_url = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc"

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        with pytest.raises(ValueError, match="local file paths"):
            client.files.upload(signed_url)


def test_files_upload_rejects_http_url() -> None:
    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        with pytest.raises(ValueError, match="local file paths"):
            client.files.upload("http://example.com/invoice.pdf")


def test_files_upload_uses_direct_storage_upload_for_local_paths(
    monkeypatch: pytest.MonkeyPatch,
    tmp_path,
) -> None:
    captured: dict[str, object] = {}
    document_path = tmp_path / "invoice.pdf"
    document_path.write_bytes(b"%PDF-1.4")

    def fake_prepared_request(request: object) -> dict[str, object]:
        captured.setdefault("requests", []).append(request)  # type: ignore[union-attr]
        if getattr(request, "url") == "/files/upload":
            return {
                "fileId": "file_123",
                "uploadUrl": "https://storage.googleapis.com/signed-upload",
                "uploadMethod": "PUT",
                "uploadHeaders": {"Content-Type": "application/pdf"},
                "mimeData": {"filename": "invoice.pdf", "url": "https://storage.retab.com/org_1/file_123.pdf"},
                "expiresAt": "2026-04-24T12:00:00Z",
            }
        return {
            "filename": "invoice.pdf",
            "url": "https://storage.retab.com/org_1/file_123.pdf",
        }

    class FakeUploadResponse:
        def raise_for_status(self) -> None:
            return None

    def fake_put(*args: object, **kwargs: object) -> FakeUploadResponse:
        captured["put_args"] = args
        captured["put_kwargs"] = kwargs
        return FakeUploadResponse()

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        monkeypatch.setattr(client.client, "put", fake_put)
        result = client.files.upload(document_path)

    assert result.filename == "invoice.pdf"
    assert result.url == "https://storage.retab.com/org_1/file_123.pdf"
    assert result.id == "file_123"
    requests = captured["requests"]  # type: ignore[assignment]
    assert [getattr(request, "url") for request in requests] == [
        "/files/upload",
        "/files/upload/file_123/complete",
    ]
    assert getattr(requests[0], "data") == {
        "filename": "invoice.pdf",
        "content_type": "application/pdf",
        "size_bytes": 8,
        "sha256": "e16fa5d9b51928755db85b917f0297babaf22c7a47e97d9212adab56e61ba04e",
    }
    assert getattr(requests[1], "data") == {
        "sha256": "e16fa5d9b51928755db85b917f0297babaf22c7a47e97d9212adab56e61ba04e",
    }
    assert captured["put_args"] == ("https://storage.googleapis.com/signed-upload",)
    assert captured["put_kwargs"]["headers"] == {"Content-Type": "application/pdf"}  # type: ignore[index]


def test_files_upload_file_object_streams_from_start_and_uses_object_filename(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    captured: dict[str, object] = {}
    file_obj = BytesIO(b"hello world")
    file_obj.name = "/tmp/report.txt"  # type: ignore[attr-defined]
    file_obj.seek(6)

    def fake_prepared_request(request: object) -> dict[str, object]:
        captured.setdefault("requests", []).append(request)  # type: ignore[union-attr]
        if getattr(request, "url") == "/files/upload":
            return {
                "fileId": "file_123",
                "uploadUrl": "https://storage.googleapis.com/signed-upload",
                "uploadMethod": "PUT",
                "uploadHeaders": {"Content-Type": "text/plain"},
                "mimeData": {"filename": "report.txt", "url": "https://storage.retab.com/org_1/file_123.txt"},
                "expiresAt": "2026-04-24T12:00:00Z",
            }
        return {
            "filename": "report.txt",
            "url": "https://storage.retab.com/org_1/file_123.txt",
        }

    class FakeUploadResponse:
        def raise_for_status(self) -> None:
            return None

    def fake_put(*_args: object, **kwargs: object) -> FakeUploadResponse:
        content = kwargs["content"]
        captured["uploaded_bytes"] = content.read()  # type: ignore[attr-defined]
        return FakeUploadResponse()

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        monkeypatch.setattr(client.client, "put", fake_put)
        result = client.files.upload(file_obj)

    assert result.filename == "report.txt"
    requests = captured["requests"]  # type: ignore[assignment]
    assert getattr(requests[0], "data") == {
        "filename": "report.txt",
        "content_type": "text/plain",
        "size_bytes": 11,
        "sha256": "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
    }
    assert getattr(requests[1], "data") == {
        "sha256": "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
    }
    assert captured["uploaded_bytes"] == b"hello world"


@pytest.mark.asyncio
async def test_async_files_upload_uses_direct_storage_upload_for_local_paths(
    monkeypatch: pytest.MonkeyPatch,
    tmp_path,
) -> None:
    captured: dict[str, object] = {}
    document_path = tmp_path / "invoice.pdf"
    document_path.write_bytes(b"%PDF-1.4")

    async def fake_prepared_request(request: object) -> dict[str, object]:
        captured.setdefault("requests", []).append(request)  # type: ignore[union-attr]
        if getattr(request, "url") == "/files/upload":
            return {
                "fileId": "file_123",
                "uploadUrl": "https://storage.googleapis.com/signed-upload",
                "uploadMethod": "PUT",
                "uploadHeaders": {"Content-Type": "application/pdf"},
                "mimeData": {"filename": "invoice.pdf", "url": "https://storage.retab.com/org_1/file_123.pdf"},
                "expiresAt": "2026-04-24T12:00:00Z",
            }
        return {
            "filename": "invoice.pdf",
            "url": "https://storage.retab.com/org_1/file_123.pdf",
        }

    class FakeUploadResponse:
        def raise_for_status(self) -> None:
            return None

    async def fake_put(*args: object, **kwargs: object) -> FakeUploadResponse:
        chunks: list[bytes] = []
        async for chunk in kwargs["content"]:  # type: ignore[index, union-attr]
            chunks.append(chunk)
        captured["put_args"] = args
        captured["uploaded_bytes"] = b"".join(chunks)
        captured["put_headers"] = kwargs["headers"]  # type: ignore[index]
        return FakeUploadResponse()

    async with AsyncRetab(api_key="test", base_url="http://example.com/v1") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        monkeypatch.setattr(client.client, "put", fake_put)
        result = await client.files.upload(document_path)

    requests = captured["requests"]  # type: ignore[assignment]
    assert result.filename == "invoice.pdf"
    assert result.url == "https://storage.retab.com/org_1/file_123.pdf"
    assert [getattr(request, "url") for request in requests] == [
        "/files/upload",
        "/files/upload/file_123/complete",
    ]
    assert getattr(requests[0], "data") == {
        "filename": "invoice.pdf",
        "content_type": "application/pdf",
        "size_bytes": 8,
        "sha256": "e16fa5d9b51928755db85b917f0297babaf22c7a47e97d9212adab56e61ba04e",
    }
    assert getattr(requests[1], "data") == {
        "sha256": "e16fa5d9b51928755db85b917f0297babaf22c7a47e97d9212adab56e61ba04e",
    }
    assert captured["put_args"] == ("https://storage.googleapis.com/signed-upload",)
    assert captured["uploaded_bytes"] == b"%PDF-1.4"
    assert captured["put_headers"] == {"Content-Type": "application/pdf"}


def test_files_upload_does_not_complete_when_direct_upload_fails(
    monkeypatch: pytest.MonkeyPatch,
    tmp_path,
) -> None:
    captured: dict[str, object] = {}
    document_path = tmp_path / "invoice.pdf"
    document_path.write_bytes(b"%PDF-1.4")

    def fake_prepared_request(request: object) -> dict[str, object]:
        captured.setdefault("requests", []).append(request)  # type: ignore[union-attr]
        return {
            "fileId": "file_123",
            "uploadUrl": "https://storage.googleapis.com/signed-upload",
            "uploadMethod": "PUT",
            "uploadHeaders": {"Content-Type": "application/pdf"},
            "mimeData": {"filename": "invoice.pdf", "url": "https://storage.retab.com/org_1/file_123.pdf"},
            "expiresAt": "2026-04-24T12:00:00Z",
        }

    class FakeUploadResponse:
        def raise_for_status(self) -> None:
            request = httpx.Request("PUT", "https://storage.googleapis.com/signed-upload")
            response = httpx.Response(403, request=request, text="signature mismatch")
            raise httpx.HTTPStatusError("upload failed", request=request, response=response)

    def fake_put(*_args: object, **_kwargs: object) -> FakeUploadResponse:
        return FakeUploadResponse()

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        monkeypatch.setattr(client.client, "put", fake_put)
        with pytest.raises(httpx.HTTPStatusError):
            client.files.upload(document_path)

    requests = captured["requests"]  # type: ignore[assignment]
    assert [getattr(request, "url") for request in requests] == ["/files/upload"]


def test_files_upload_rejects_non_seekable_file_objects() -> None:
    class NonSeekable(BytesIO):
        def seekable(self) -> bool:
            return False

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        with pytest.raises(ValueError, match="seekable"):
            client.files.upload(NonSeekable(b"%PDF-1.4"))


def test_extractions_create_accepts_signed_bucket_url(monkeypatch: pytest.MonkeyPatch) -> None:
    captured: dict[str, object] = {}
    signed_url = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc"

    def fake_prepared_request(request: object) -> dict[str, object]:
        captured["request"] = request
        return {
            "id": "extr_123",
            "file": {
                "id": "file_123",
                "filename": "invoice.pdf",
                "mime_type": "application/pdf",
            },
            "model": "retab-small",
            "json_schema": {"type": "object"},
            "n_consensus": 1,
            "image_resolution_dpi": 192,
            "output": {},
            "consensus": {"choices": []},
            "metadata": {},
        }

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        result = client.extractions.create(
            document=signed_url,
            json_schema={"type": "object"},
            model="retab-small",
        )

    assert result.id == "extr_123"
    assert getattr(captured["request"], "url") == "/extractions"
    assert getattr(captured["request"], "data")["document"] == {
        "filename": "invoice.pdf",
        "url": signed_url,
    }


def test_mime_data_id_extracts_file_id_from_canonical_retab_storage_url() -> None:
    mime = MIMEData(
        filename="invoice.pdf",
        url="https://storage.retab.com/org_1/file_123.pdf",
    )

    assert mime.id == "file_123"


@pytest.mark.parametrize(
    ("resource_name", "prepare_name", "kwargs"),
    [
        (
            "classifications",
            "_prepare_create",
            {
                "categories": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "parses",
            "_prepare_create",
            {
                "model": "retab-small",
            },
        ),
        (
            "splits",
            "_prepare_create",
            {
                "subdocuments": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "partitions",
            "_prepare_create",
            {
                "key": "invoice_number",
                "instructions": "Split the document into one chunk per invoice number.",
                "model": "retab-small",
            },
        ),
        (
            "extractions",
            "prepare_create",
            {
                "json_schema": {"type": "object"},
                "model": "retab-small",
            },
        ),
    ],
)
def test_resource_create_builders_preserve_signed_bucket_urls(
    resource_name: str,
    prepare_name: str,
    kwargs: dict[str, object],
) -> None:
    signed_url = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc"

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        resource = getattr(client, resource_name)
        request = getattr(resource, prepare_name)(document=signed_url, **kwargs)

    assert request.data["document"] == {
        "filename": "invoice.pdf",
        "url": signed_url,
    }


@pytest.mark.parametrize(
    ("resource_name", "prepare_name", "kwargs"),
    [
        (
            "classifications",
            "_prepare_create",
            {
                "categories": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "parses",
            "_prepare_create",
            {
                "model": "retab-small",
            },
        ),
        (
            "splits",
            "_prepare_create",
            {
                "subdocuments": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "partitions",
            "_prepare_create",
            {
                "key": "invoice_number",
                "instructions": "Split the document into one chunk per invoice number.",
                "model": "retab-small",
            },
        ),
        (
            "extractions",
            "prepare_create",
            {
                "json_schema": {"type": "object"},
                "model": "retab-small",
            },
        ),
    ],
)
def test_resource_create_builders_accept_retab_storage_url_string(
    resource_name: str,
    prepare_name: str,
    kwargs: dict[str, object],
) -> None:
    retab_url = "https://storage.retab.com/org_1/file_123.pdf"

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        resource = getattr(client, resource_name)
        request = getattr(resource, prepare_name)(document=retab_url, **kwargs)

    assert request.data["document"] == {
        "filename": "file_123.pdf",
        "url": retab_url,
    }


@pytest.mark.parametrize(
    ("resource_name", "prepare_name", "kwargs"),
    [
        (
            "classifications",
            "_prepare_create",
            {
                "categories": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "parses",
            "_prepare_create",
            {
                "model": "retab-small",
            },
        ),
        (
            "splits",
            "_prepare_create",
            {
                "subdocuments": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "partitions",
            "_prepare_create",
            {
                "key": "invoice_number",
                "instructions": "Split the document into one chunk per invoice number.",
                "model": "retab-small",
            },
        ),
        (
            "extractions",
            "prepare_create",
            {
                "json_schema": {"type": "object"},
                "model": "retab-small",
            },
        ),
    ],
)
def test_resource_create_builders_accept_uploaded_mime_ref(
    resource_name: str,
    prepare_name: str,
    kwargs: dict[str, object],
) -> None:
    mime_ref = MIMEData(
        filename="invoice.pdf",
        url="https://storage.retab.com/org_1/file_123.pdf",
    )

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        resource = getattr(client, resource_name)
        request = getattr(resource, prepare_name)(document=mime_ref, **kwargs)

    assert request.data["document"] == {
        "filename": "invoice.pdf",
        "url": "https://storage.retab.com/org_1/file_123.pdf",
    }


def test_resource_create_builders_include_supported_route_fields() -> None:
    signed_url = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc"

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        classification = client.classifications._prepare_create(
            document=signed_url,
            categories=[{"name": "invoice", "description": "Invoice documents"}],
            model="retab-small",
            first_n_pages=1,
            instructions="Use the cover page.",
        )
        parse = client.parses._prepare_create(
            document=signed_url,
            model="retab-small",
            instructions="Keep invoice sections separate.",
        )
        split = client.splits._prepare_create(
            document=signed_url,
            subdocuments=[{"name": "invoice", "description": "Invoice documents"}],
            model="retab-small",
            instructions="Split attachments from invoices.",
        )
        extraction = client.extractions.prepare_create(
            document=signed_url,
            json_schema={"type": "object"},
            model="retab-small",
            additional_messages=[{"role": "user", "content": "Use totals only."}],
        )

    assert classification.data["first_n_pages"] == 1
    assert classification.data["instructions"] == "Use the cover page."
    assert parse.data["instructions"] == "Keep invoice sections separate."
    assert split.data["instructions"] == "Split attachments from invoices."
    assert extraction.data["additional_messages"] == [{"role": "user", "content": "Use totals only."}]


def test_resource_list_builders_include_filename_filters() -> None:
    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        assert client.classifications._prepare_list(filename="invoice.pdf").params["filename"] == "invoice.pdf"
        assert client.parses._prepare_list(filename="invoice.pdf").params["filename"] == "invoice.pdf"
        assert client.splits._prepare_list(filename="invoice.pdf").params["filename"] == "invoice.pdf"
        assert client.partitions._prepare_list(filename="invoice.pdf").params["filename"] == "invoice.pdf"
        assert client.extractions.prepare_list(filename="invoice.pdf").params["filename"] == "invoice.pdf"


def test_partitions_list_and_delete_use_resource_routes(monkeypatch: pytest.MonkeyPatch) -> None:
    captured: list[object] = []

    def fake_prepared_request(request: object) -> dict[str, object]:
        captured.append(request)
        if getattr(request, "method") == "GET":
            return {"data": [], "list_metadata": {"before": None, "after": None}}
        return {}

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        page = client.partitions.list(limit=5, filename="invoice.pdf")
        client.partitions.delete("prtn_123")

    assert len(page.data) == 0
    assert getattr(captured[0], "url") == "/partitions"
    assert getattr(captured[0], "params") == {
        "limit": 5,
        "order": "desc",
        "filename": "invoice.pdf",
    }
    assert getattr(captured[1], "url") == "/partitions/prtn_123"
    assert getattr(captured[1], "method") == "DELETE"


def test_splits_create_uses_new_resource_route(monkeypatch: pytest.MonkeyPatch) -> None:
    captured: dict[str, object] = {}

    def fake_prepared_request(request: object) -> dict[str, object]:
        captured["request"] = request
        return {
            "id": "split_123",
            "file": {
                "id": "file_123",
                "filename": "invoice.txt",
                "mime_type": "text/plain",
            },
            "model": "retab-small",
            "subdocuments": [
                {"name": "invoice", "description": "Invoice documents"},
            ],
            "output": [
                {"name": "invoice", "pages": [1]},
            ],
            "consensus": {
                "choices": [],
                "likelihoods": None,
            },
            "usage": {
                "page_count": 1,
                "credits": 1.0,
            },
        }

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        result = client.splits.create(
            document=_sample_document(),
            model="retab-small",
            subdocuments=[{"name": "invoice", "description": "Invoice documents"}],
        )

    assert isinstance(result, Split)
    assert result.id == "split_123"
    assert getattr(captured["request"], "url") == "/splits"


def test_partitions_create_uses_new_resource_route(monkeypatch: pytest.MonkeyPatch) -> None:
    captured: dict[str, object] = {}

    def fake_prepared_request(request: object) -> dict[str, object]:
        captured["request"] = request
        return {
            "id": "prtn_123",
            "file": {
                "id": "file_123",
                "filename": "invoice.txt",
                "mime_type": "text/plain",
            },
            "model": "retab-small",
            "key": "invoice_number",
            "instructions": "Split the document into one chunk per invoice number.",
            "n_consensus": 3,
            "output": [
                {
                    "key": "INV-001",
                    "pages": [1, 2],
                }
            ],
            "consensus": {
                "choices": [],
                "likelihoods": None,
            },
            "usage": {
                "credits": 1.0,
            },
        }

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        result = client.partitions.create(
            document=_sample_document(),
            key="invoice_number",
            instructions="Split the document into one chunk per invoice number.",
            model="retab-small",
            n_consensus=3,
        )

    assert isinstance(result, Partition)
    assert result.id == "prtn_123"
    assert result.output[0].key == "INV-001"
    assert result.consensus.choices == []
    assert result.usage is not None
    assert result.usage.credits == 1.0
    assert getattr(captured["request"], "url") == "/partitions"
    assert getattr(captured["request"], "data") == {
        "document": {
            "filename": "invoice.txt",
            "url": "data:text/plain;base64,aW52b2ljZQ==",
        },
        "key": "invoice_number",
        "instructions": "Split the document into one chunk per invoice number.",
        "model": "retab-small",
        "n_consensus": 3,
    }


@pytest.mark.parametrize(
    ("method_name", "kwargs", "response_payload", "expected_url"),
    [
        (
            "extract",
            {
                "document": _sample_document(),
                "model": "retab-small",
                "json_schema": {"type": "object", "properties": {"invoice_number": {"type": "string"}}},
            },
            {
                "id": "chatcmpl_123",
                "object": "chat.completion",
                "created": 1710000000,
                "model": "retab-small",
                "choices": [
                    {
                        "index": 0,
                        "message": {
                            "role": "assistant",
                            "content": "{\"invoice_number\":\"INV-001\"}",
                            "parsed": {"invoice_number": "INV-001"},
                        },
                        "finish_reason": "stop",
                    }
                ],
            },
            "/documents/extract",
        ),
        (
            "split",
            {
                "document": _sample_document(),
                "model": "retab-small",
                "subdocuments": [{"name": "invoice", "description": "Invoice documents"}],
            },
            {
                "splits": [{"name": "invoice", "pages": [1], "partitions": []}],
                "consensus": {"choices": [], "likelihoods": None},
                "usage": {"page_count": 1, "credits": 1.0},
            },
            "/documents/split",
        ),
        (
            "classify",
            {
                "document": _sample_document(),
                "model": "retab-small",
                "categories": [{"name": "invoice", "description": "Invoice documents"}],
            },
            {
                "classification": {
                    "category": "invoice",
                    "reasoning": "Detected invoice markers.",
                },
                "consensus": {"choices": [], "likelihood": 1.0},
                "usage": {"page_count": 1, "credits": 1.0},
            },
            "/documents/classify",
        ),
    ],
)
def test_legacy_document_methods_warn_and_keep_compat_routes(
    monkeypatch: pytest.MonkeyPatch,
    method_name: str,
    kwargs: dict[str, object],
    response_payload: dict[str, object],
    expected_url: str,
) -> None:
    captured: dict[str, object] = {}

    def fake_prepared_request(request: object) -> dict[str, object]:
        captured["request"] = request
        return response_payload

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        method = getattr(client.documents, method_name)
        with pytest.warns(DeprecationWarning):
            method(**kwargs)

    assert getattr(captured["request"], "url") == expected_url
