import pytest

from retab import Retab
from retab.types.mime import MIMEData
from retab.types.partitions import PartitionResponse
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

    assert isinstance(result, PartitionResponse)
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
