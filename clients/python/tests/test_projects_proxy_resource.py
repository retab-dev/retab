from typing import Any

import pytest

from retab.resources.projects.client import AsyncProjects, Projects
from retab.types.documents.extract import RetabParsedChatCompletion
from retab.types.standards import PreparedRequest


def _completion_payload() -> dict[str, Any]:
    return {
        "id": "chatcmpl_123",
        "object": "chat.completion",
        "created": 1710000000,
        "model": "retab-micro",
        "choices": [
            {
                "index": 0,
                "finish_reason": "stop",
                "message": {
                    "role": "assistant",
                    "content": "{}",
                    "parsed": {},
                },
            }
        ],
        "extraction_id": "extract_123",
    }


class _SyncRecordingClient:
    def __init__(self) -> None:
        self.requests: list[PreparedRequest] = []

    def _prepared_request(self, request: PreparedRequest) -> dict[str, Any]:
        self.requests.append(request)
        return _completion_payload()


class _AsyncRecordingClient:
    def __init__(self) -> None:
        self.requests: list[PreparedRequest] = []

    async def _prepared_request(self, request: PreparedRequest) -> dict[str, Any]:
        self.requests.append(request)
        return _completion_payload()


def test_projects_extract_uses_legacy_proxy_path() -> None:
    client = _SyncRecordingClient()
    resource = Projects(client=client)

    response = resource.extract(project_id="proj_123", document=b"fake pdf bytes")

    assert isinstance(response, RetabParsedChatCompletion)
    assert client.requests[-1].url == "/projects/extract/proj_123"


@pytest.mark.asyncio
async def test_async_projects_extract_uses_legacy_proxy_path() -> None:
    client = _AsyncRecordingClient()
    resource = AsyncProjects(client=client)

    response = await resource.extract(project_id="proj_123", document=b"fake pdf bytes")

    assert isinstance(response, RetabParsedChatCompletion)
    assert client.requests[-1].url == "/projects/extract/proj_123"
