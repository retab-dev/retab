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


def validate_create_messages_response(response: DocumentMessage | None) -> None:
    """Validate the response from create_messages endpoint."""
    # Assert the instance
    assert isinstance(response, DocumentMessage), f"Response should be of type DocumentMessage, received {type(response)}"

    # Assert the response has required attributes
    assert response.messages is not None, "Response messages should not be None"
    assert isinstance(response.messages, list), "Response messages should be a list"
    assert len(response.messages) > 0, "Response messages should not be empty"

    # Assert each message has "role" and "content"
    for message in response.messages:
        assert "role" in message, "Message should have a 'role' key"
        assert "content" in message, "Message should have a 'content' key"
        assert message["role"] in ["user", "developer"], f"Role should be 'user' or 'developer', got '{message['role']}'"
        assert isinstance(message["content"], (str, list)), "Message content should be str or list"


# Test the create_messages endpoint
async def base_test_create_messages(
    client_type: ClientType,
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
) -> None:
    """Base test function for create_messages endpoint."""
    document = booking_confirmation_file_path_1
    modality: Literal["text", "native"] = "native"
    response: DocumentMessage | None = None

    if client_type == "sync":
        with sync_client as client:
            response = client.documents.create_messages(
                document=document,
                modality=modality,
            )
    elif client_type == "async":
        async with async_client as client:
            response = await client.documents.create_messages(
                document=document,
                modality=modality,
            )
    
    validate_create_messages_response(response)


@pytest.mark.asyncio
@pytest.mark.xdist_group(name="create_messages")
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_create_messages_basic(
    client_type: ClientType,
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
) -> None:
    """Test basic create_messages functionality."""
    await base_test_create_messages(
        client_type=client_type,
        sync_client=sync_client,
        async_client=async_client,
        booking_confirmation_file_path_1=booking_confirmation_file_path_1,
    )


@pytest.mark.asyncio
@pytest.mark.parametrize("request_number", range(5))
async def test_create_messages_overload(
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    request_number: int,
) -> None:
    """Test create_messages with multiple concurrent requests."""
    await asyncio.sleep(request_number * 0.1)
    await base_test_create_messages(
        client_type="async",
        sync_client=sync_client,
        async_client=async_client,
        booking_confirmation_file_path_1=booking_confirmation_file_path_1,
    )


@pytest.mark.asyncio
async def test_create_messages_with_optional_params(
    sync_client: Retab,
    booking_confirmation_file_path_1: str,
) -> None:
    """Test create_messages with optional parameters."""
    document = booking_confirmation_file_path_1
    
    with sync_client as client:
        response = client.documents.create_messages(
            document=document,
            modality="native",
            image_resolution_dpi=96,
        )
    
    validate_create_messages_response(response)



@pytest.mark.asyncio
async def test_create_messages_different_modalities(
    sync_client: Retab,
    booking_confirmation_file_path_1: str,
) -> None:
    """Test create_messages with different modalities."""
    document = booking_confirmation_file_path_1
    
    with sync_client as client:
        # Test with 'native' modality
        response_native = client.documents.create_messages(
            document=document,
            modality="native",
        )
        validate_create_messages_response(response_native)
        
        # Test with 'text' modality
        response_text = client.documents.create_messages(
            document=document,
            modality="text",
        )
        validate_create_messages_response(response_text)
        
        # Both should be valid responses with messages
        assert len(response_native.messages) > 0
        assert len(response_text.messages) > 0
        assert all("content" in msg for msg in response_native.messages)
        assert all("content" in msg for msg in response_text.messages)


@pytest.mark.asyncio
async def test_create_messages_async_vs_sync(
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
) -> None:
    """Test that sync and async clients produce consistent results."""
    document = booking_confirmation_file_path_1
    modality = "native"
    
    # Sync request
    with sync_client as client:
        sync_response = client.documents.create_messages(
            document=document,
            modality=modality,
        )
    
    # Async request
    async with async_client as client:
        async_response = await client.documents.create_messages(
            document=document,
            modality=modality,
        )
    
    # Validate both responses
    validate_create_messages_response(sync_response)
    validate_create_messages_response(async_response)
    
    # Both should be valid DocumentMessage instances
    assert type(sync_response) is type(async_response)
    assert len(sync_response.messages) > 0
    assert len(async_response.messages) > 0
    assert all("content" in msg for msg in sync_response.messages)
    assert all("content" in msg for msg in async_response.messages)
