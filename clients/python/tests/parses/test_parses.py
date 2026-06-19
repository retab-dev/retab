import pytest

from retab.types.parses import Parse, ParseRequest

from mocks import mock_retab
from samples import sample_document

# Whole module is unit (pure offline; no server/credentials needed).
pytestmark = pytest.mark.unit


def test_parses_create_uses_new_resource_route() -> None:
    client, rec = mock_retab(
        {
            "id": "parse_123",
            "file": {
                "id": "file_123",
                "filename": "invoice.txt",
                "mime_type": "text/plain",
            },
            "model": "retab-small",
            "table_parsing_format": "html",
            "output": {
                "pages": ["invoice"],
                "text": "invoice",
            },
            "usage": {
                "page_count": 1,
                "credits": 1.0,
            },
        }
    )
    with client:
        result = client.parses.create(
            document=sample_document(),
            model="retab-small",
        )

    assert isinstance(result, Parse)
    assert result.id == "parse_123"
    assert result.output is not None
    assert result.output.text == "invoice"
    assert rec.request.url == "/v1/parses"


def test_parse_request_ignores_benchmark_model_policy_fields() -> None:
    # ``ParseRequest`` mirrors the backend's ``extra="ignore"`` config, so
    # internal benchmark/policy fields are dropped rather than rejected:
    # they must never surface on the validated public model.
    request = ParseRequest.model_validate(
        {
            "document": sample_document().model_dump(mode="json"),
            "candidate_scope": "exact_model",
            "capacity_retry_owner": "caller",
        }
    )
    dumped = request.model_dump()
    assert "candidate_scope" not in dumped
    assert "capacity_retry_owner" not in dumped
