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


@pytest.mark.asyncio
async def test_create_inputs_with_idempotency(
    sync_client: Retab, booking_confirmation_file_path_1: str, booking_confirmation_json_schema: dict[str, Any]
) -> None:
    idempotency_key = nanoid.generate()

    response_initial = sync_client.documents.create_inputs(
        json_schema=booking_confirmation_json_schema,
        document=booking_confirmation_file_path_1,
        idempotency_key=idempotency_key,
    )
    await asyncio.sleep(2)
    t0 = time.time()
    response_second = sync_client.documents.create_inputs(
        json_schema=booking_confirmation_json_schema,
        document=booking_confirmation_file_path_1,
        idempotency_key=idempotency_key,
    )
    t1 = time.time()
    assert t1 - t0 < 10, "Request should take less than 10 seconds"
    assert response_initial.messages == response_second.messages, "Response should be the same"


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "error_scenario",
    ["invalid_json_schema", "missing_document"],
)
async def test_create_inputs_with_idempotency_error_scenarios(
    sync_client: Retab, booking_confirmation_file_path_1: str, booking_confirmation_json_schema: dict[str, Any], error_scenario: str
) -> None:
    """Test idempotency behavior when various error conditions occur during extraction."""
    idempotency_key = str(nanoid.generate())

    # Set up error scenarios
    if error_scenario == "invalid_json_schema":
        json_schema = {"invalid": "schema without required fields"}
        document = booking_confirmation_file_path_1
    elif error_scenario == "missing_document":
        json_schema = booking_confirmation_json_schema
        document = "/nonexistent/file.pdf"
    else:
        pytest.fail(f"Unknown error scenario: {error_scenario}")

    response1: Any = None
    response2: Any = None
    raised_exception_1: Exception | None = None
    raised_exception_2: Exception | None = None

    # First request attempt
    try:
        response1 = sync_client.documents.create_inputs(
            json_schema=json_schema,
            document=document,
            idempotency_key=idempotency_key,
        )
    except Exception as e:
        raised_exception_1 = e

    await asyncio.sleep(2)
    t0 = time.time()

    # Second request attempt with same idempotency key
    try:
        response2 = sync_client.documents.create_inputs(
            json_schema=json_schema,
            document=document,
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
        assert isinstance(raised_exception_1, type(raised_exception_2)), "Exception types should be the same"
        # For HTTP errors, verify the status codes are the same
        if isinstance(raised_exception_1, httpx.HTTPStatusError) and isinstance(raised_exception_2, httpx.HTTPStatusError):
            assert (
                raised_exception_1.response.status_code == raised_exception_2.response.status_code
            ), "HTTP status codes should be the same"
    else:
        # If no exception was raised, both responses should be successful and identical
        assert raised_exception_2 is None, "Both requests should succeed consistently"
        assert response1 is not None, "Response should not be None when successful"
        assert response2 is not None, "Response should not be None when successful"
        validate_response(response1)
        validate_response(response2)
        assert response1.messages == response2.messages, "Response content should be identical"

