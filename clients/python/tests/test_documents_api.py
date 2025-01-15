import pytest
import json
from typing import Literal, get_args, Any
from pydantic import BaseModel
from uiform import UiForm, AsyncUiForm
from uiform.types.documents.parse import DocumentExtractResponse
# List of AI Providers to test
AI_MODELS = Literal[
    "gpt-4o-mini",
    "grok-2-vision-1212",
    "claude-3-5-sonnet-latest",
    "gemini-1.5-flash-8b"
]
ClientType = Literal[
    "sync",
    "async",
]
ResponseModeType = Literal[
    "stream", 
    "parse",
]

def validate_extraction_response(response: DocumentExtractResponse) -> None:
    # Assert the instance
    assert isinstance(response, DocumentExtractResponse), f"Response should be of type DocumentExtractResponse, received {type(response)}"

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
@pytest.mark.asyncio
@pytest.mark.parametrize("model", get_args(AI_MODELS))
@pytest.mark.parametrize("client_type", get_args(ClientType))
@pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
# @pytest.mark.parametrize("stream", [False])
async def test_extract_success(model: AI_MODELS, client_type: ClientType, response_mode: ResponseModeType, sync_client: UiForm, async_client: AsyncUiForm, booking_confirmation_file_path: str, booking_confirmation_json_schema: dict[str, Any]) -> None:
    json_schema = booking_confirmation_json_schema
    document=booking_confirmation_file_path
    modality: Literal["text"] = "text"
    
    # Wait a random amount of time between 0 and 2 seconds (to avoid sending multiple requests to the same provider at the same time)
    if client_type == "sync":
        with sync_client as client:
            if response_mode == "stream":
                with client.documents.extractions.stream(
                    json_schema=json_schema,
                    document=document,
                    model=model,
                    modality=modality
                ) as stream_iterator:
                    response = stream_iterator.__next__()
                    for response in stream_iterator: pass
                validate_extraction_response(response)
            else:
                response = client.documents.extractions.parse(
                    json_schema=json_schema,
                    document=document,
                    model=model,
                    modality=modality
                )
            validate_extraction_response(response)
    
    if client_type == "async":
        async with async_client:
            if response_mode == "stream":
                async with async_client.documents.extractions.stream(
                    json_schema=json_schema,
                    document=document,
                    model=model,
                    modality=modality
                ) as stream_async_iterator:
                    response = await stream_async_iterator.__anext__()
                    async for response in stream_async_iterator: pass
            else:
                response = await async_client.documents.extractions.parse(
                    json_schema=json_schema,
                    document=document,
                    model=model,
                    modality=modality
                )
            validate_extraction_response(response)

