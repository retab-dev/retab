import asyncio
import time
from typing import Any, Literal, get_args

import httpx
import nanoid  # type: ignore
import pytest

from retab import AsyncRetab, Retab
from retab.types.documents.create_messages import DocumentMessage

ClientType = Literal[
    "sync",
    "async",
]


def validate_response(response: DocumentMessage | None) -> None:
    # Assert the instance
    assert isinstance(response, DocumentMessage), f"Response should be of type DocumentMessage, received {type(response)}"

    # Assert the response content is not None
    assert response.messages is not None, "Response content should not be None"
    assert isinstance(response.messages, list), "Response content should be a list"

    # Assert each message has "role" and "content"
    for message in response.messages:
        assert "role" in message, "Message should have a 'role' key"
        assert "content" in message, "Message should have a 'content' key"
        assert message["role"] in ["user", "developer"], f"Role should be 'user' or 'developer', got '{message['role']}'"


# Test the extraction endpoint
async def base_test_create_inputs(
    client_type: ClientType,
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: dict[str, Any],
) -> None:
    json_schema = booking_confirmation_json_schema
    document = booking_confirmation_file_path_1
    response: Any | None = None

    if client_type == "sync":
        with sync_client as client:
            response = client.documents.create_inputs(
                json_schema=json_schema,
                document=document,
            )
    if client_type == "async":
        async with async_client:
            response = await async_client.documents.create_inputs(
                json_schema=json_schema,
                document=document,
            )

    validate_response(response)


@pytest.mark.asyncio
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_create_inputs(
    client_type: ClientType,
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: dict[str, Any],
) -> None:
    await base_test_create_inputs(
        client_type=client_type,
        sync_client=sync_client,
        async_client=async_client,
        booking_confirmation_file_path_1=booking_confirmation_file_path_1,
        booking_confirmation_json_schema=booking_confirmation_json_schema,
    )