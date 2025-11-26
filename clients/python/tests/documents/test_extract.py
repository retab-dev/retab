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
    response: RetabParsedChatCompletion | None = None

    if client_type == "sync":
        with sync_client as client:
            if response_mode == "stream":
                with client.documents.extract_stream(
                    json_schema=json_schema,
                    document=document,
                    model=model,
                ) as stream_iterator:
                    response = stream_iterator.__next__()
                    for response in stream_iterator:
                        pass
            else:
                response = client.documents.extract(
                    json_schema=json_schema,
                    document=document,
                    model=model,
                )
    if client_type == "async":
        async with async_client:
            if response_mode == "stream":
                async with async_client.documents.extract_stream(
                    json_schema=json_schema,
                    document=document,
                    model=model,
                ) as stream_async_iterator:
                    response = await stream_async_iterator.__anext__()
                    async for response in stream_async_iterator:
                        pass
            else:
                response = await async_client.documents.extract(
                    json_schema=json_schema,
                    document=document,
                    model=model,
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

