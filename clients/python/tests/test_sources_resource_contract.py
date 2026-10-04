import pytest

from mocks import mock_async_retab, mock_retab
from retab.types.classifications import ExtractionSourcesMode
from retab.types.extraction_sources import SourceJobMode


pytestmark = pytest.mark.unit


def test_sources_response_remains_available_from_package_root() -> None:
    from retab import SourcesResponse
    from retab.types.extraction_sources import SourcesResponse as ResourceSourcesResponse

    assert SourcesResponse is ResourceSourcesResponse


def sources_response(mode: str = "located") -> dict[str, object]:
    return {
        "object": "extraction.sources",
        "extraction_id": "extr_1",
        "extraction": {},
        "document_type": "pdf",
        "file": {"id": "file_1", "filename": "invoice.pdf", "mime_type": "application/pdf"},
        "sources": {},
        "job": {
            "object": "extraction.sources.job",
            "id": "src_1",
            "mode": mode,
            "status": "completed",
            "is_partial": False,
            "revision": 1,
            "updated_at": "2026-10-04T18:00:00Z",
        },
    }


def test_sources_create_and_get_are_nested_resources() -> None:
    client, recorder = mock_retab(sources_response())
    with client:
        created = client.extractions.sources.create("extr_1", mode=SourceJobMode.LOCATED, background=True)
        assert created.job is not None
        fetched = client.extractions.sources.get("extr_1", mode=ExtractionSourcesMode.LOCATED, job_id=created.job.id)
        assert fetched.job is not None and fetched.job.id == created.job.id

    create, get = recorder.requests
    assert create.method == "POST" and get.method == "GET"
    assert create.url == get.url == "/v1/extractions/extr_1/sources"
    assert create.data is not None and create.data["mode"] == "located" and create.data["background"] is True
    assert get.params is not None and get.params["mode"] == "located" and get.params["job_id"] == "src_1"


@pytest.mark.asyncio
async def test_async_sources_create_and_get_are_nested_resources() -> None:
    client, recorder = mock_async_retab(sources_response("cited"))
    async with client:
        created = await client.extractions.sources.create("extr_1", mode=SourceJobMode.CITED, background=True)
        assert created.job is not None
        await client.extractions.sources.get("extr_1", mode=ExtractionSourcesMode.CITED, job_id=created.job.id)
    create, get = recorder.requests
    assert create.method == "POST" and get.method == "GET"
    assert create.url == get.url == "/v1/extractions/extr_1/sources"
    assert create.data is not None and create.data["mode"] == "cited"
    assert get.params is not None and get.params["mode"] == "cited" and get.params["job_id"] == "src_1"
