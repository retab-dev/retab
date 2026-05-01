import asyncio
import json
import time
from typing import Any, Literal, get_args

import httpx
import nanoid  # type: ignore
import pytest
from pydantic import BaseModel

from retab import AsyncRetab, Retab
from retab.resources.documents.client import AsyncDocuments, Documents
from retab.types.documents.extract import RetabParsedChatCompletion, RetabParsedChatCompletionChunk, maybe_parse_to_pydantic
from retab.types.standards import PreparedRequest
from retab.types.chat import ChatCompletionRetabMessage
# List of AI Providers to test
AI_MODELS = Literal[
    "retab-micro",
    "retab-micro",
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

    assert response.data == response.choices[0].message.parsed, "Response data should mirror choices[0].message.parsed"
    assert response.text == response.choices[0].message.content, "Response text should mirror choices[0].message.content"

    response_dump = response.model_dump()
    assert "data" in response_dump, "Serialized response should include top-level data"
    assert "text" in response_dump, "Serialized response should include top-level text"


def test_extract_result_serializes_computed_fields() -> None:
    response = RetabParsedChatCompletion.model_validate(
        {
            "id": "chatcmpl_test",
            "object": "chat.completion",
            "created": 1,
            "model": "retab-micro",
            "choices": [
                {
                    "index": 0,
                    "finish_reason": "stop",
                    "message": {
                        "role": "assistant",
                        "content": "{\"status\":\"ok\"}",
                        "parsed": {"status": "ok"},
                    },
                }
            ],
        }
    )

    assert response.data == {"status": "ok"}
    assert response.text == "{\"status\":\"ok\"}"
    assert response.model_dump()["data"] == {"status": "ok"}
    assert response.model_dump()["text"] == "{\"status\":\"ok\"}"
    assert response.model_dump(mode="json")["data"] == {"status": "ok"}
    assert response.model_dump(mode="json")["text"] == "{\"status\":\"ok\"}"


def test_stream_chunk_to_completion_preserves_finish_reason() -> None:
    chunk = RetabParsedChatCompletionChunk.model_validate(
        {
            "id": "chatcmpl_test",
            "object": "chat.completion.chunk",
            "created": 1,
            "model": "retab-large",
            "choices": [
                {
                    "index": 0,
                    "finish_reason": "length",
                    "delta": {
                        "content": "{\"status\":\"partial\"}",
                        "flat_parsed": {"status": "partial"},
                        "flat_likelihoods": {},
                        "flat_deleted_keys": [],
                        "is_valid_json": True,
                        "full_parsed": {"status": "partial"},
                    },
                }
            ],
        }
    )

    completion = chunk.chunk_accumulator().to_completion()

    assert completion.choices[0].finish_reason == "length"


def test_documents_extract_stream_preserves_finish_reason(monkeypatch: pytest.MonkeyPatch) -> None:
    class FakeClient:
        def _prepared_request_stream(self, request: PreparedRequest):
            assert request.url == "/documents/extractions"
            yield {
                "id": "chatcmpl_test",
                "object": "chat.completion.chunk",
                "created": 1,
                "model": "retab-large",
                "choices": [
                    {
                        "index": 0,
                        "finish_reason": "length",
                        "delta": {
                            "content": "{\"status\":\"partial\"}",
                            "flat_parsed": {"status": "partial"},
                            "flat_likelihoods": {},
                            "flat_deleted_keys": [],
                            "is_valid_json": True,
                            "full_parsed": {"status": "partial"},
                        },
                    }
                ],
            }

    def fake_prepare_extract(self: Documents, **_: Any) -> PreparedRequest:
        return PreparedRequest(method="POST", url="/documents/extractions", data={})

    monkeypatch.setattr(Documents, "_prepare_extract", fake_prepare_extract)

    documents = Documents(FakeClient())
    with pytest.warns(DeprecationWarning):
        with documents.extract_stream(SIMPLE_SCHEMA, "retab-large", document="unused.pdf") as stream:
            responses = list(stream)

    assert responses[-1].choices[0].finish_reason == "length"


def test_documents_extract_stream_preserves_token_content_deltas(monkeypatch: pytest.MonkeyPatch) -> None:
    class FakeClient:
        def _prepared_request_stream(self, request: PreparedRequest):
            assert request.url == "/documents/extractions"
            for content, full_parsed, finish_reason in [
                ('{"name":"Ac', None, None),
                ('me","amount":42', None, None),
                ('.5}', {"name": "Acme", "amount": 42.5}, "stop"),
            ]:
                delta: dict[str, Any] = {
                    "content": content,
                    "flat_parsed": {},
                    "flat_likelihoods": {},
                    "flat_deleted_keys": [],
                    "is_valid_json": full_parsed is not None,
                }
                if full_parsed is not None:
                    delta["full_parsed"] = full_parsed
                yield {
                    "id": "chatcmpl_stream",
                    "object": "chat.completion.chunk",
                    "created": 1,
                    "model": "retab-large",
                    "choices": [
                        {
                            "index": 0,
                            "finish_reason": finish_reason,
                            "delta": delta,
                        }
                    ],
                }

    def fake_prepare_extract(self: Documents, **_: Any) -> PreparedRequest:
        return PreparedRequest(method="POST", url="/documents/extractions", data={})

    monkeypatch.setattr(Documents, "_prepare_extract", fake_prepare_extract)

    snapshots: list[tuple[str | None, Any]] = []
    documents = Documents(FakeClient())
    with pytest.warns(DeprecationWarning):
        with documents.extract_stream(SIMPLE_SCHEMA, "retab-large", document="unused.pdf") as stream:
            for response in stream:
                snapshots.append((response.text, response.data))

    assert [text for text, _data in snapshots[:3]] == [
        '{"name":"Ac',
        '{"name":"Acme","amount":42',
        '{"name": "Acme", "amount": 42.5}',
    ]
    assert json.loads(snapshots[-1][0] or "{}") == {"name": "Acme", "amount": 42.5}
    assert snapshots[-1][1].name == "Acme"
    assert snapshots[-1][1].amount == 42.5


@pytest.mark.asyncio
async def test_async_documents_extract_stream_preserves_token_content_deltas(monkeypatch: pytest.MonkeyPatch) -> None:
    class FakeClient:
        async def _prepared_request_stream(self, request: PreparedRequest):
            assert request.url == "/documents/extractions"
            for content, full_parsed, finish_reason in [
                ('{"name":"Ac', None, None),
                ('me","amount":42', None, None),
                ('.5}', {"name": "Acme", "amount": 42.5}, "stop"),
            ]:
                delta: dict[str, Any] = {
                    "content": content,
                    "flat_parsed": {},
                    "flat_likelihoods": {},
                    "flat_deleted_keys": [],
                    "is_valid_json": full_parsed is not None,
                }
                if full_parsed is not None:
                    delta["full_parsed"] = full_parsed
                yield {
                    "id": "chatcmpl_stream",
                    "object": "chat.completion.chunk",
                    "created": 1,
                    "model": "retab-large",
                    "choices": [
                        {
                            "index": 0,
                            "finish_reason": finish_reason,
                            "delta": delta,
                        }
                    ],
                }

    def fake_prepare_extract(self: AsyncDocuments, **_: Any) -> PreparedRequest:
        return PreparedRequest(method="POST", url="/documents/extractions", data={})

    monkeypatch.setattr(AsyncDocuments, "_prepare_extract", fake_prepare_extract)

    snapshots: list[tuple[str | None, Any]] = []
    documents = AsyncDocuments(FakeClient())
    with pytest.warns(DeprecationWarning):
        async with documents.extract_stream(SIMPLE_SCHEMA, "retab-large", document="unused.pdf") as stream:
            async for response in stream:
                snapshots.append((response.text, response.data))

    assert [text for text, _data in snapshots[:3]] == [
        '{"name":"Ac',
        '{"name":"Acme","amount":42',
        '{"name": "Acme", "amount": 42.5}',
    ]
    assert json.loads(snapshots[-1][0] or "{}") == {"name": "Acme", "amount": 42.5}
    assert snapshots[-1][1].name == "Acme"
    assert snapshots[-1][1].amount == 42.5
# ---------------------------------------------------------------------------
# Unit tests for maybe_parse_to_pydantic
# ---------------------------------------------------------------------------

SIMPLE_SCHEMA = {
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "amount": {"type": "number"},
    },
    "required": ["name", "amount"],
}


def _build_completion(content: str, parsed: Any = None) -> RetabParsedChatCompletion:
    """Helper to build a minimal RetabParsedChatCompletion for unit tests."""
    return RetabParsedChatCompletion.model_validate(
        {
            "id": "chatcmpl_test",
            "object": "chat.completion",
            "created": 1,
            "model": "test-model",
            "choices": [
                {
                    "index": 0,
                    "finish_reason": "stop",
                    "message": {
                        "role": "assistant",
                        "content": content,
                        "parsed": parsed,
                    },
                }
            ],
        }
    )


def test_maybe_parse_to_pydantic_success() -> None:
    """When content matches the schema, parsed should be a Pydantic model."""
    content = '{"name": "Alice", "amount": 42.0}'
    response = _build_completion(content)
    result = maybe_parse_to_pydantic(SIMPLE_SCHEMA, response)

    assert result.choices[0].message.parsed is not None
    assert isinstance(result.choices[0].message.parsed, BaseModel)
    assert result.data is not None
    assert result.data.name == "Alice"  # type: ignore[union-attr]
    assert result.data.amount == 42.0  # type: ignore[union-attr]


def test_maybe_parse_to_pydantic_preserves_existing_parsed_on_failure() -> None:
    """Consensus bug regression: when Pydantic validation fails but parsed was
    already set from the API response, the existing dict must be preserved."""
    # Content has null for a required string field → Pydantic validation will fail
    content = '{"name": null, "amount": 42.0}'
    existing_parsed = {"name": None, "amount": 42.0}
    response = _build_completion(content, parsed=existing_parsed)

    # Confirm parsed is set before calling maybe_parse_to_pydantic
    assert response.choices[0].message.parsed == existing_parsed

    result = maybe_parse_to_pydantic(SIMPLE_SCHEMA, response)

    # parsed must NOT be overwritten with None — the existing dict is preserved
    assert result.choices[0].message.parsed is not None
    assert result.data is not None
    assert result.data["amount"] == 42.0  # type: ignore[index]


def test_maybe_parse_to_pydantic_falls_back_to_raw_dict() -> None:
    """When Pydantic validation fails and parsed is None, fall back to raw dict."""
    content = '{"name": null, "amount": 42.0}'
    response = _build_completion(content, parsed=None)

    result = maybe_parse_to_pydantic(SIMPLE_SCHEMA, response)

    # Should fall back to the raw filtered dict, not None
    assert result.choices[0].message.parsed is not None
    assert result.data is not None
    assert result.data["amount"] == 42.0  # type: ignore[index]


def test_maybe_parse_to_pydantic_filters_reasoning_fields() -> None:
    """Auxiliary fields like reasoning___ should be filtered from parsed."""
    content = '{"name": "Bob", "amount": 10.0, "reasoning___amount": "Because..."}'
    response = _build_completion(content)
    result = maybe_parse_to_pydantic(SIMPLE_SCHEMA, response)

    assert result.data is not None
    dumped = result.data.model_dump() if isinstance(result.data, BaseModel) else result.data
    assert "reasoning___amount" not in dumped


def test_maybe_parse_to_pydantic_consensus_multi_choice() -> None:
    """Full consensus scenario: choices[0] is consensus with nulls, choices[1..N] are votes."""
    consensus_content = json.dumps({
        "name": None,
        "amount": 100.0,
    })
    consensus_parsed = {"name": None, "amount": 100.0}

    vote_content = json.dumps({
        "name": "Alice",
        "amount": 100.0,
        "reasoning___amount": "Stated in the doc",
    })
    vote_parsed = {"name": "Alice", "amount": 100.0, "reasoning___amount": "Stated in the doc"}

    response = RetabParsedChatCompletion.model_validate(
        {
            "id": "chatcmpl_consensus",
            "object": "chat.completion",
            "created": 1,
            "model": "test-model",
            "likelihoods": {"name": 0.5, "amount": 1.0},
            "choices": [
                {
                    "index": 0,
                    "finish_reason": "stop",
                    "message": {
                        "role": "assistant",
                        "content": consensus_content,
                        "parsed": consensus_parsed,
                        "likelihoods": {"name": 0.5, "amount": 1.0},
                    },
                },
                {
                    "index": 1,
                    "finish_reason": "stop",
                    "message": {
                        "role": "assistant",
                        "content": vote_content,
                        "parsed": vote_parsed,
                    },
                },
            ],
        }
    )

    result = maybe_parse_to_pydantic(SIMPLE_SCHEMA, response)

    # data (choices[0].message.parsed) must not be None
    assert result.data is not None
    assert result.data["amount"] == 100.0  # type: ignore[index]
    # choices[1] should be untouched
    assert result.choices[1].message.parsed == vote_parsed


def test_maybe_parse_to_pydantic_empty_content() -> None:
    """When content is empty, parsed should remain unchanged."""
    response = _build_completion("", parsed=None)
    result = maybe_parse_to_pydantic(SIMPLE_SCHEMA, response)
    assert result.choices[0].message.parsed is None
    assert result.data is None


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
                with pytest.warns(DeprecationWarning):
                    with client.documents.extract_stream(
                        json_schema=json_schema,
                        document=document,
                        model=model,
                    ) as stream_iterator:
                        response = stream_iterator.__next__()
                        for response in stream_iterator:
                            pass
            else:
                with pytest.warns(DeprecationWarning):
                    response = client.documents.extract(
                        json_schema=json_schema,
                        document=document,
                        model=model,
                    )
    if client_type == "async":
        async with async_client:
            if response_mode == "stream":
                with pytest.warns(DeprecationWarning):
                    async with async_client.documents.extract_stream(
                        json_schema=json_schema,
                        document=document,
                        model=model,
                    ) as stream_async_iterator:
                        response = await stream_async_iterator.__anext__()
                        async for response in stream_async_iterator:
                            pass
            else:
                with pytest.warns(DeprecationWarning):
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
        model="retab-micro",
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
        model="retab-micro",
        client_type="async",
        response_mode="parse",
        sync_client=sync_client,
        async_client=async_client,
        booking_confirmation_file_path_1=booking_confirmation_file_path_1,
        booking_confirmation_json_schema=booking_confirmation_json_schema,
    )


# Tests for additional_messages feature
@pytest.mark.asyncio
async def test_extract_with_text_additional_message(
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: dict[str, Any],
) -> None:
    """Test extraction with a text-only additional message providing context."""
    additional_messages: list[ChatCompletionRetabMessage] = [
        {
            "role": "user",
            "content": "Important context: Please extract all booking details carefully, paying attention to dates and confirmation numbers."
        }
    ]

    async with async_client:
        with pytest.warns(DeprecationWarning):
            response = await async_client.documents.extract(
                json_schema=booking_confirmation_json_schema,
                document=booking_confirmation_file_path_1,
                model="retab-micro",
                additional_messages=additional_messages,
            )

    validate_extraction_response(response)


@pytest.mark.asyncio
async def test_extract_with_multipart_additional_message(
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: dict[str, Any],
) -> None:
    """Test extraction with additional messages containing both text and image URL."""
    additional_messages = [
        {
            "role": "user",
            "content": "Context for extraction: This is a booking confirmation document. Please extract all relevant fields."
        },
        {
            "role": "user",
            "content": [
                {
                    "type": "text",
                    "text": "Reference image for brand identification:"
                },
                {
                    "type": "image_url",
                    "image_url": {
                        "url": "https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png",
                        "detail": "auto"
                    }
                }
            ]
        }
    ]

    async with async_client:
        with pytest.warns(DeprecationWarning):
            response = await async_client.documents.extract(
                json_schema=booking_confirmation_json_schema,
                document=booking_confirmation_file_path_1,
                model="retab-micro",
                additional_messages=additional_messages,
            )

    validate_extraction_response(response)


@pytest.mark.asyncio
async def test_extract_with_system_additional_message(
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: dict[str, Any],
) -> None:
    """Test extraction with a system message to set behavior."""
    additional_messages: list[ChatCompletionRetabMessage] = [
        {
            "role": "system",
            "content": "You are a precise document extraction assistant. Extract information exactly as it appears in the document without making assumptions."
        },
        {
            "role": "user",
            "content": "Please be thorough in extracting all booking confirmation details."
        }
    ]

    async with async_client:
        with pytest.warns(DeprecationWarning):
            response = await async_client.documents.extract(
                json_schema=booking_confirmation_json_schema,
                document=booking_confirmation_file_path_1,
                model="retab-micro",
                additional_messages=additional_messages,
            )

    validate_extraction_response(response)


@pytest.mark.asyncio
async def test_extract_with_multiple_additional_messages(
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: dict[str, Any],
) -> None:
    """Test extraction with multiple additional messages of different types."""
    additional_messages: list[ChatCompletionRetabMessage] = [
        {
            "role": "developer",
            "content": "Extract data with high precision. When uncertain, prefer leaving fields empty rather than guessing."
        },
        {
            "role": "user",
            "content": "This booking confirmation is from a travel agency. Look for: confirmation number, dates, guest names, and total amount."
        },
        {
            "role": "user",
            "content": "If any field is unclear or ambiguous, please indicate that in your extraction."
        }
    ]

    async with async_client:
        with pytest.warns(DeprecationWarning):
            response = await async_client.documents.extract(
                json_schema=booking_confirmation_json_schema,
                document=booking_confirmation_file_path_1,
                model="retab-micro",
                additional_messages=additional_messages,
            )

    validate_extraction_response(response)
