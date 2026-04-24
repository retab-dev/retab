import pytest

from retab import Retab
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


def test_files_upload_accepts_signed_bucket_url(monkeypatch: pytest.MonkeyPatch) -> None:
    captured: dict[str, object] = {}
    signed_url = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc"

    def fake_prepared_request(request: object) -> dict[str, object]:
        captured["request"] = request
        return {
            "fileId": "file_123",
            "filename": "invoice.pdf",
        }

    with Retab(api_key="test", base_url="http://example.com/v1") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        result = client.files.upload(signed_url)

    assert result.file_id == "file_123"
    assert getattr(captured["request"], "url") == "/files/upload"
    assert getattr(captured["request"], "data") == {
        "mimeData": {
            "filename": "invoice.pdf",
            "url": signed_url,
        },
    }


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
