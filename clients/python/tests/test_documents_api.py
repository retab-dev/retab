import pytest
import json
import time
import asyncio
import random
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
ClientsFixtureType = dict[str, UiForm | AsyncUiForm]


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
async def test_extract_success(model: AI_MODELS, sync_client: UiForm, async_client: AsyncUiForm, booking_confirmation_file_path: str, booking_confirmation_json_schema: dict[str, Any]) -> None:
    json_schema = booking_confirmation_json_schema
    document=booking_confirmation_file_path
    modality: Literal["text"] = "text"
    
    # Wait a random amount of time between 0 and 2 seconds (to avoid sending multiple requests to the same provider at the same time)
    with sync_client as client:
        print(f"[Sync -- UiForm] Testing {model}...")
        response = client.documents.extractions.parse(
            json_schema=json_schema,
            document=document,
            model=model,
            modality=modality
        )
        validate_extraction_response(response)
    # Wait a little before sending the next request...
    await asyncio.sleep(random.uniform(1, 5))
    
    async with async_client:
        print(f"[Async -- AsyncUiForm] Testing {model}...")
        response = await async_client.documents.extractions.parse(
            json_schema=json_schema,
            document=document,
            model=model,
            modality=modality
        )
        validate_extraction_response(response)

