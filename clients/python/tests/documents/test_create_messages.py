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
            browser_canvas="A4",
        )
    
    validate_create_messages_response(response)


@pytest.mark.asyncio
async def test_create_messages_with_idempotency(
    sync_client: Retab, 
    booking_confirmation_file_path_1: str,
) -> None:
    """Test create_messages with idempotency key."""
    idempotency_key = nanoid.generate()
    document = booking_confirmation_file_path_1
    modality = "native"

    with sync_client as client:
        # First request
        response_initial = client.documents.create_messages(
            document=document,
            modality=modality,
            idempotency_key=idempotency_key,
        )
        
        await asyncio.sleep(2)
        t0 = time.time()
        
        # Second request with same idempotency key
        response_second = client.documents.create_messages(
            document=document,
            modality=modality,
            idempotency_key=idempotency_key,
        )
        t1 = time.time()
        
        # Verify both responses
        validate_create_messages_response(response_initial)
        validate_create_messages_response(response_second)
        
        # Should be fast due to idempotency
        assert t1 - t0 < 10, "Request should take less than 10 seconds due to idempotency"
        
        # Messages should be identical
        assert response_initial.messages == response_second.messages, "Response messages should be identical for idempotent requests"


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "error_scenario",
    ["missing_document"],
)
async def test_create_messages_error_scenarios(
    sync_client: Retab, 
    booking_confirmation_file_path_1: str, 
    error_scenario: str,
) -> None:
    """Test create_messages with various error conditions."""
    idempotency_key = str(nanoid.generate())
    
    # Set up error scenarios
    if error_scenario == "missing_document":
        document = "/nonexistent/file.pdf"
        modality = "native"
    else:
        pytest.fail(f"Unknown error scenario: {error_scenario}")
    
    response1: Any = None
    response2: Any = None
    raised_exception_1: Exception | None = None
    raised_exception_2: Exception | None = None
    
    with sync_client as client:
        # First request attempt
        try:
            response1 = client.documents.create_messages(
                document=document,
                modality=modality,
                idempotency_key=idempotency_key,
            )
        except Exception as e:
            raised_exception_1 = e

        await asyncio.sleep(2)
        t0 = time.time()
        
        # Second request attempt with same idempotency key
        try:
            response2 = client.documents.create_messages(
                document=document,
                modality=modality,
                idempotency_key=idempotency_key,
            )
        except Exception as e:
            raised_exception_2 = e
            
        t1 = time.time()
        assert t1 - t0 < 10, "Request should take less than 10 seconds"
        
        # Verify that both requests behaved consistently (idempotent behavior)
        if raised_exception_1 is not None:
            assert raised_exception_2 is not None, "Both requests should raise exceptions consistently"
            assert response1 is None, "Response should be None when exception is raised"
            assert response2 is None, "Response should be None when exception is raised"
            # Verify that both exceptions are of the same type
            assert type(raised_exception_1) is type(raised_exception_2), "Exception types should be the same"
            # For HTTP errors, verify the status codes are the same
            if isinstance(raised_exception_1, httpx.HTTPStatusError) and isinstance(raised_exception_2, httpx.HTTPStatusError):
                assert raised_exception_1.response.status_code == raised_exception_2.response.status_code, "HTTP status codes should be the same"
        else:
            # If no exception was raised, both responses should be successful and identical
            assert raised_exception_2 is None, "Both requests should succeed consistently"
            assert response1 is not None, "Response should not be None when successful"
            assert response2 is not None, "Response should not be None when successful"
            validate_create_messages_response(response1)
            validate_create_messages_response(response2)
            assert response1.messages == response2.messages, "Response messages should be identical"


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
