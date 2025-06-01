import asyncio
import json
import time
from typing import Any, Literal, get_args

import httpx
import nanoid  # type: ignore
import pytest
from pydantic import BaseModel

from uiform import AsyncUiForm, UiForm
from uiform.types.documents.extractions import UiParsedChatCompletion

# List of AI Providers to test
AI_MODELS = Literal[
    "gpt-4o-mini",
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


def validate_extraction_response(response: UiParsedChatCompletion | None) -> None:
    # Assert the instance
    assert isinstance(response, UiParsedChatCompletion), f"Response should be of type UiParsedChatCompletion, received {type(response)}"

    # Assert the response content is not None
    assert response.choices[0].message.content is not None, "Response content should not be None"
    # Assert that content is a valid JSON object
    try:
        json.loads(response.choices[0].message.content)
    except json.JSONDecodeError:
        assert False, "Response content should be a valid JSON object"
    # Assert that the response.choices[0].message.parsed is a valid pydantic BaseModel instance
    assert isinstance(response.choices[0].message.parsed, BaseModel), "Response parsed should be a valid pydantic BaseModel instance"


# Test the extraction endpoint
async def base_test_extract(
    model: AI_MODELS,
    client_type: ClientType,
    response_mode: ResponseModeType,
    sync_client: UiForm,
    async_client: AsyncUiForm,
    booking_confirmation_file_path: str,
    booking_confirmation_json_schema: dict[str, Any],
) -> None:
    json_schema = booking_confirmation_json_schema
    document = booking_confirmation_file_path
    modality: Literal["text"] = "text"
    response: UiParsedChatCompletion | None = None
    # Wait a random amount of time between 0 and 2 seconds (to avoid sending multiple requests to the same provider at the same time)
    if client_type == "sync":
        with sync_client as client:
            if response_mode == "stream":
                with client.documents.extractions.stream(json_schema=json_schema, document=document, model=model, modality=modality) as stream_iterator:
                    response = stream_iterator.__next__()
                    for response in stream_iterator:
                        pass
            else:
                response = client.documents.extractions.parse(json_schema=json_schema, document=document, model=model, modality=modality)
    if client_type == "async":
        async with async_client:
            if response_mode == "stream":
                async with async_client.documents.extractions.stream(json_schema=json_schema, document=document, model=model, modality=modality) as stream_async_iterator:
                    response = await stream_async_iterator.__anext__()
                    async for response in stream_async_iterator:
                        pass
            else:
                response = await async_client.documents.extractions.parse(json_schema=json_schema, document=document, model=model, modality=modality)
    validate_extraction_response(response)


@pytest.mark.asyncio
@pytest.mark.xdist_group(name="openai")
@pytest.mark.parametrize("client_type", get_args(ClientType))
@pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
async def test_extract_openai(
    client_type: ClientType,
    response_mode: ResponseModeType,
    sync_client: UiForm,
    async_client: AsyncUiForm,
    booking_confirmation_file_path: str,
    booking_confirmation_json_schema: dict[str, Any],
) -> None:
    await base_test_extract(
        model="gpt-4o-mini",
        client_type=client_type,
        response_mode=response_mode,
        sync_client=sync_client,
        async_client=async_client,
        booking_confirmation_file_path=booking_confirmation_file_path,
        booking_confirmation_json_schema=booking_confirmation_json_schema,
    )


@pytest.mark.asyncio
@pytest.mark.parametrize("request_number", range(10))
async def test_extract_overload(
    sync_client: UiForm,
    async_client: AsyncUiForm,
    booking_confirmation_file_path: str,
    booking_confirmation_json_schema: dict[str, Any],
    request_number: int,
) -> None:
    await asyncio.sleep(request_number * 0.1)
    await base_test_extract(
        model="gpt-4o-mini",
        client_type="async",
        response_mode="parse",
        sync_client=sync_client,
        async_client=async_client,
        booking_confirmation_file_path=booking_confirmation_file_path,
        booking_confirmation_json_schema=booking_confirmation_json_schema,
    )


# @pytest.mark.asyncio
# @pytest.mark.xdist_group(name="anthropic")
# @pytest.mark.parametrize("client_type", get_args(ClientType))
# @pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
# async def test_extract_anthropic(client_type: ClientType, response_mode: ResponseModeType, sync_client: UiForm, async_client: AsyncUiForm, booking_confirmation_file_path: str, booking_confirmation_json_schema: dict[str, Any]) -> None:
#     await base_test_extract(
#         model="claude-3-5-sonnet-latest",
#         client_type=client_type,
#         response_mode=response_mode,
#         sync_client=sync_client,
#         async_client=async_client,
#         booking_confirmation_file_path=booking_confirmation_file_path,
#         booking_confirmation_json_schema=booking_confirmation_json_schema
#     )

# @pytest.mark.asyncio
# @pytest.mark.xdist_group(name="xai")
# @pytest.mark.parametrize("client_type", get_args(ClientType))
# @pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
# async def test_extract_xai(client_type: ClientType, response_mode: ResponseModeType, sync_client: UiForm, async_client: AsyncUiForm, booking_confirmation_file_path: str, booking_confirmation_json_schema: dict[str, Any]) -> None:
#     await base_test_extract(
#         model="grok-2-vision-1212",
#         client_type=client_type,
#         response_mode=response_mode,
#         sync_client=sync_client,
#         async_client=async_client,
#         booking_confirmation_file_path=booking_confirmation_file_path,
#         booking_confirmation_json_schema=booking_confirmation_json_schema
#     )

# @pytest.mark.asyncio
# @pytest.mark.xdist_group(name="google")
# @pytest.mark.parametrize("client_type", get_args(ClientType))
# @pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
# async def test_extract_google(client_type: ClientType, response_mode: ResponseModeType, sync_client: UiForm, async_client: AsyncUiForm, booking_confirmation_file_path: str, booking_confirmation_json_schema: dict[str, Any]) -> None:
#     await base_test_extract(
#         model="gemini-1.5-flash-8b",
#         client_type=client_type,
#         response_mode=response_mode,
#         sync_client=sync_client,
#         async_client=async_client,
#         booking_confirmation_file_path=booking_confirmation_file_path,
#         booking_confirmation_json_schema=booking_confirmation_json_schema
#     )


@pytest.mark.asyncio
async def test_extraction_with_idempotency(sync_client: UiForm, booking_confirmation_file_path: str, booking_confirmation_json_schema: dict[str, Any]) -> None:
    client = sync_client
    idempotency_key = nanoid.generate()
    response_initial = client.documents.extractions.parse(
        json_schema=booking_confirmation_json_schema, document=booking_confirmation_file_path, model="gpt-4o-mini", modality="native", idempotency_key=idempotency_key
    )
    await asyncio.sleep(2)
    t0 = time.time()
    response_second = client.documents.extractions.parse(
        json_schema=booking_confirmation_json_schema, document=booking_confirmation_file_path, model="gpt-4o-mini", modality="native", idempotency_key=idempotency_key
    )
    t1 = time.time()
    assert t1 - t0 < 10, "Request should take less than 10 seconds"
    assert response_initial.choices[0].message.content == response_second.choices[0].message.content, "Response should be the same"
    assert response_initial.choices[0].message.parsed is not None, "Parsed response should not be None for the first request"
    assert response_second.choices[0].message.parsed is not None, "Parsed response should not be None for the second request"
    assert response_initial.choices[0].message.parsed.model_dump() == response_second.choices[0].message.parsed.model_dump(), "Parsed response should be the same"


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "test_exception",
    ["before_handle_extraction", "within_extraction_parse_or_stream", "after_handle_extraction", "within_process_document_stream_generator"],
)
async def test_extraction_with_idempotency_exceptions(
    sync_client: UiForm, booking_confirmation_file_path: str, booking_confirmation_json_schema: dict[str, Any], test_exception: str
) -> None:
    client = sync_client
    # Now we will validate some exception scenarios
    idempotency_key = str(nanoid.generate())
    loading_request = client.documents.extractions.prepare_extraction(
        json_schema=booking_confirmation_json_schema,
        document=booking_confirmation_file_path,
        image_resolution_dpi=96,
        browser_canvas="A4",
        model="gpt-4o-mini",
        temperature=0,
        modality="native",
        reasoning_effort="medium",
        stream=False,
        n_consensus=1,
        store=False,
        idempotency_key=idempotency_key,
    )
    assert loading_request.data is not None, "Loading request should not be None"

    # Validate idempotency behavior for each of the following scenarios:
    loading_request.data["test_exception"] = test_exception
    response1: Any = None
    response2: Any = None
    raised_exception_1: Exception | None = None
    raised_exception_2: Exception | None = None
    try:
        loading_request.raise_for_status = True
        if test_exception == "within_process_document_stream_generator":
            loading_request.data["stream"] = True
            stream_iterator = client._prepared_request_stream(loading_request)
            response1 = stream_iterator.__next__()
            for response1 in stream_iterator:
                pass
        else:
            response1 = client._prepared_request(loading_request)
    except Exception as e:
        raised_exception_1 = e

    await asyncio.sleep(2)
    t0 = time.time()
    try:
        loading_request.raise_for_status = True
        if test_exception == "within_process_document_stream_generator":
            loading_request.data["stream"] = True
            stream_iterator = client._prepared_request_stream(loading_request)
            response2 = stream_iterator.__next__()
            for response2 in stream_iterator:
                pass
        else:
            response2 = client._prepared_request(loading_request)
    except Exception as e:
        raised_exception_2 = e
    t1 = time.time()
    assert t1 - t0 < 10, "Request should take less than 10 seconds"
    assert raised_exception_1 is not None, "Exception should be raised"
    assert raised_exception_2 is not None, "Exception should be raised"
    assert response1 is None, "Response should be None"
    assert response2 is None, "Response should be None"
    # Assert that the both exceptions is a HTTPStatusError
    assert isinstance(raised_exception_1, httpx.HTTPStatusError), "Exception should be a HTTPStatusError"
    assert isinstance(raised_exception_2, httpx.HTTPStatusError), "Exception should be a HTTPStatusError"
    # Assert that the message of both exceptions is the same
    assert raised_exception_1.args[0] == raised_exception_2.args[0], "Exception message should be the same"
