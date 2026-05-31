import pytest

from retab import Retab
from retab.types.mime import MIMEData
from retab.types.parses import Parse, ParseRequest


def _sample_document() -> MIMEData:
    return MIMEData(
        filename="invoice.txt",
        url="data:text/plain;base64,aW52b2ljZQ==",
    )


def test_parses_create_uses_new_resource_route(monkeypatch: pytest.MonkeyPatch) -> None:
    captured: dict[str, object] = {}

    def fake_prepared_request(request: object) -> dict[str, object]:
        captured["request"] = request
        return {
            "id": "parse_123",
            "file": {
                "id": "file_123",
                "filename": "invoice.txt",
                "mime_type": "text/plain",
            },
            "model": "retab-small",
            "table_parsing_format": "html",
            "image_resolution_dpi": 192,
            "output": {
                "pages": ["invoice"],
                "text": "invoice",
            },
            "usage": {
                "page_count": 1,
                "credits": 1.0,
            },
        }

    with Retab(api_key="test", base_url="http://example.com") as client:
        monkeypatch.setattr(client, "_prepared_request", fake_prepared_request)
        result = client.parses.create(
            document=_sample_document(),
            model="retab-small",
        )

    assert isinstance(result, Parse)
    assert result.id == "parse_123"
    assert result.output is not None
    assert result.output.text == "invoice"
    assert getattr(captured["request"], "url") == "/v1/parses"


def test_parse_request_ignores_benchmark_model_policy_fields() -> None:
    # ``ParseRequest`` mirrors the backend's ``extra="ignore"`` config, so
    # internal benchmark/policy fields are dropped rather than rejected:
    # they must never surface on the validated public model.
    request = ParseRequest.model_validate(
        {
            "document": _sample_document().model_dump(mode="json"),
            "candidate_scope": "exact_model",
            "capacity_retry_owner": "caller",
        }
    )
    dumped = request.model_dump()
    assert "candidate_scope" not in dumped
    assert "capacity_retry_owner" not in dumped
