import asyncio
import json
import time
from typing import Any, Literal, get_args

import httpx
import nanoid  # type: ignore
import pytest
from pydantic import BaseModel

from retab import AsyncRetab, Retab
from retab.types.documents.extract import RetabParsedChatCompletion

# Global modality setting for all tests
MODALITY = "native_fast"

# List of AI Providers to test
AI_MODELS = Literal[
    "gpt-4.1-nano",
    "gemini-2.5-flash-lite",
    # "grok-2-vision-1212",
    # "claude-3-5-sonnet-latest",
    # "gemini-1.5-flash-8b"
]
ClientType = Literal[
    "sync",
    "async",
]
ResponseModeType = Literal[
    "stream",
    "parse",
]


def validate_extraction_response(response: RetabParsedChatCompletion | None) -> None:
    # Assert the instance
    assert isinstance(response, RetabParsedChatCompletion), f"Response should be of type RetabParsedChatCompletion, received {type(response)}"

    # Assert the response content is not None
    assert response.choices[0].message.content is not None, "Response content should not be None"
    # Assert that content is a valid JSON object
    try:
        json.loads(response.choices[0].message.content)
    except json.JSONDecodeError:
        assert False, "Response content should be a valid JSON object"
    # Assert that the response.choices[0].message.parsed is either a valid pydantic BaseModel instance or None
    # (None can happen when an invalid schema is provided and parsing fails)
    assert (
        response.choices[0].message.parsed is None or isinstance(response.choices[0].message.parsed, BaseModel)
    ), f"Response parsed should be a valid pydantic BaseModel instance or None, received {type(response.choices[0].message.parsed)}"


# Test the extraction endpoint
async def base_test_extract(
    model: AI_MODELS,
    client_type: ClientType,
    response_mode: ResponseModeType,
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: dict[str, Any],
) -> None:
    json_schema = booking_confirmation_json_schema
    document = booking_confirmation_file_path_1
    modality = MODALITY
    response: RetabParsedChatCompletion | None = None

    # First create a processor

    # Wait a random amount of time between 0 and 2 seconds (to avoid sending multiple requests to the same provider at the same time)
    if client_type == "sync":
        with sync_client as client:
            if response_mode == "stream":
                with client.documents.extract_stream(
                    json_schema=json_schema,
                    document=document,
                    model=model,
                    modality=modality,
                ) as stream_iterator:
                    response = stream_iterator.__next__()
                    for response in stream_iterator:
                        pass
            else:
                response = client.documents.extract(
                    json_schema=json_schema,
                    document=document,
                    model=model,
                    modality=modality,
                )
    if client_type == "async":
        async with async_client:
            if response_mode == "stream":
                async with async_client.documents.extract_stream(
                    json_schema=json_schema,
                    document=document,
                    model=model,
                    modality=modality,
                ) as stream_async_iterator:
                    response = await stream_async_iterator.__anext__()
                    async for response in stream_async_iterator:
                        pass
            else:
                response = await async_client.documents.extract(
                    json_schema=json_schema,
                    document=document,
                    model=model,
                    modality=modality,
                )
    validate_extraction_response(response)


@pytest.mark.asyncio
@pytest.mark.xdist_group(name="openai")
@pytest.mark.parametrize("client_type", get_args(ClientType))
@pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
async def test_extract_openai(
    client_type: ClientType,
    response_mode: ResponseModeType,
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: dict[str, Any],
) -> None:
    await base_test_extract(
        model="gpt-4.1-nano",
        client_type=client_type,
        response_mode=response_mode,
        sync_client=sync_client,
        async_client=async_client,
        booking_confirmation_file_path_1=booking_confirmation_file_path_1,
        booking_confirmation_json_schema=booking_confirmation_json_schema,
    )


@pytest.mark.asyncio
@pytest.mark.parametrize("request_number", range(10))
async def test_extract_overload(
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: dict[str, Any],
    request_number: int,
) -> None:
    await asyncio.sleep(request_number * 0.1)
    await base_test_extract(
        model="gpt-4.1-nano",
        client_type="async",
        response_mode="parse",
        sync_client=sync_client,
        async_client=async_client,
        booking_confirmation_file_path_1=booking_confirmation_file_path_1,
        booking_confirmation_json_schema=booking_confirmation_json_schema,
    )


# @pytest.mark.asyncio
# @pytest.mark.xdist_group(name="anthropic")
# @pytest.mark.parametrize("client_type", get_args(ClientType))
# @pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
# async def test_extract_anthropic(client_type: ClientType, response_mode: ResponseModeType, sync_client: Retab, async_client: AsyncRetab, booking_confirmation_file_path_1: str, booking_confirmation_json_schema: dict[str, Any]) -> None:
#     await base_test_extract(
#         model="claude-3-5-sonnet-latest",
#         client_type=client_type,
#         response_mode=response_mode,
#         sync_client=sync_client,
#         async_client=async_client,
#         booking_confirmation_file_path_1=booking_confirmation_file_path_1,
#         booking_confirmation_json_schema=booking_confirmation_json_schema
#     )

# @pytest.mark.asyncio
# @pytest.mark.xdist_group(name="xai")
# @pytest.mark.parametrize("client_type", get_args(ClientType))
# @pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
# async def test_extract_xai(client_type: ClientType, response_mode: ResponseModeType, sync_client: Retab, async_client: AsyncRetab, booking_confirmation_file_path_1: str, booking_confirmation_json_schema: dict[str, Any]) -> None:
#     await base_test_extract(
#         model="grok-2-vision-1212",
#         client_type=client_type,
#         response_mode=response_mode,
#         sync_client=sync_client,
#         async_client=async_client,
#         booking_confirmation_file_path_1=booking_confirmation_file_path_1,
#         booking_confirmation_json_schema=booking_confirmation_json_schema
#     )

# @pytest.mark.asyncio
# @pytest.mark.xdist_group(name="google")
# @pytest.mark.parametrize("client_type", get_args(ClientType))
# @pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
# async def test_extract_google(client_type: ClientType, response_mode: ResponseModeType, sync_client: Retab, async_client: AsyncRetab, booking_confirmation_file_path_1: str, booking_confirmation_json_schema: dict[str, Any]) -> None:
#     await base_test_extract(
#         model="gemini-1.5-flash-8b",
#         client_type=client_type,
#         response_mode=response_mode,
#         sync_client=sync_client,
#         async_client=async_client,
#         booking_confirmation_file_path_1=booking_confirmation_file_path_1,
#         booking_confirmation_json_schema=booking_confirmation_json_schema
#     )


@pytest.mark.asyncio
async def test_extraction_with_idempotency(sync_client: Retab, booking_confirmation_file_path_1: str, booking_confirmation_json_schema: dict[str, Any]) -> None:
    idempotency_key = nanoid.generate()
    model = "gpt-4o-mini"
    modality = MODALITY



    response_initial = sync_client.documents.extract(
        json_schema=booking_confirmation_json_schema, document=booking_confirmation_file_path_1, model=model, modality=modality, idempotency_key=idempotency_key
    )
    await asyncio.sleep(2)
    t0 = time.time()
    response_second = sync_client.documents.extract(
        json_schema=booking_confirmation_json_schema, document=booking_confirmation_file_path_1, model=model, modality=modality, idempotency_key=idempotency_key
    )
    t1 = time.time()
    assert t1 - t0 < 10, "Request should take less than 10 seconds"
    assert response_initial.choices[0].message.content == response_second.choices[0].message.content, "Response should be the same"
    assert response_initial.choices[0].message.parsed is not None, "Parsed response should not be None for the first request"
    assert response_second.choices[0].message.parsed is not None, "Parsed response should not be None for the second request"
    assert response_initial.choices[0].message.parsed.model_dump() == response_second.choices[0].message.parsed.model_dump(), "Parsed response should be the same"



@pytest.mark.asyncio
@pytest.mark.parametrize(
    "error_scenario",
    ["invalid_model", "invalid_json_schema", "missing_document"],
)
async def test_extraction_with_idempotency_error_scenarios(
    sync_client: Retab, booking_confirmation_file_path_1: str, booking_confirmation_json_schema: dict[str, Any], error_scenario: str
) -> None:
    """Test idempotency behavior when various error conditions occur during extraction."""
    idempotency_key = str(nanoid.generate())
    
    # Set up error scenarios
    if error_scenario == "invalid_model":
        model = "invalid-model-name"
        json_schema = booking_confirmation_json_schema
        document = booking_confirmation_file_path_1
    elif error_scenario == "invalid_json_schema":
        model = "gpt-4o-mini"
        json_schema = {"invalid": "schema without required fields"}
        document = booking_confirmation_file_path_1
    elif error_scenario == "missing_document":
        model = "gpt-4o-mini"
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
        response1 = sync_client.documents.extract(
            json_schema=json_schema,
            document=document,
            image_resolution_dpi=96,
            browser_canvas="A4",
            model=model,
            temperature=0,
            modality=MODALITY,
            reasoning_effort="medium",
            n_consensus=1,
            idempotency_key=idempotency_key,
        )
    except Exception as e:
        raised_exception_1 = e

    await asyncio.sleep(2)
    t0 = time.time()
    
    # Second request attempt with same idempotency key
    try:
        response2 = sync_client.documents.extract(
            json_schema=json_schema,
            document=document,
            image_resolution_dpi=96,
            browser_canvas="A4",
            model=model,
            temperature=0,
            modality=MODALITY,
            reasoning_effort="medium",
            n_consensus=1,
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
        assert type(raised_exception_1) == type(raised_exception_2), "Exception types should be the same"
        # For HTTP errors, verify the status codes are the same
        if isinstance(raised_exception_1, httpx.HTTPStatusError) and isinstance(raised_exception_2, httpx.HTTPStatusError):
            assert raised_exception_1.response.status_code == raised_exception_2.response.status_code, "HTTP status codes should be the same"
    else:
        # If no exception was raised, both responses should be successful and identical
        assert raised_exception_2 is None, "Both requests should succeed consistently"
        assert response1 is not None, "Response should not be None when successful"
        assert response2 is not None, "Response should not be None when successful"
        validate_extraction_response(response1)
        validate_extraction_response(response2)
        assert response1.choices[0].message.content == response2.choices[0].message.content, "Response content should be identical"
