import asyncio
import time
from typing import Any, Literal, get_args

import httpx
import nanoid  # type: ignore
import pytest

from retab import AsyncRetab, Retab
from retab.types.documents.parse import ParseResult, TableParsingFormat

# List of AI Models to test (focusing on models that support parsing)
AI_MODELS = Literal[
    "retab-micro",
    "retab-micro",
]

ClientType = Literal[
    "sync",
    "async",
]

# Parse doesn't have streaming, so we only test the basic parse response
ResponseModeType = Literal[
    "parse",
]


def validate_parse_response(response: ParseResult | None) -> None:
    """Validate that the parse response has the expected structure and content."""
    # Assert the instance
    assert isinstance(response, ParseResult), f"Response should be of type ParseResult, received {type(response)}"
    
    # Assert the response has required fields
    assert response.document is not None, "Response document should not be None"
    assert response.usage is not None, "Response usage should not be None"
    assert response.pages is not None, "Response pages should not be None"
    assert response.text is not None, "Response text should not be None"
    
    # Assert usage information is valid
    assert response.usage.page_count > 0, "Page count should be greater than 0"
    assert response.usage.credits >= 0, "Credits should be non-negative"
    
    # Assert pages is a list of strings
    assert isinstance(response.pages, list), "Pages should be a list"
    assert all(isinstance(page, str) for page in response.pages), "All pages should be strings"
    assert len(response.pages) > 0, "Should have at least one page"
    
    # Assert text is a non-empty string
    assert isinstance(response.text, str), "Text should be a string"
    assert len(response.text.strip()) > 0, "Text should not be empty"


# Base test function for parse endpoint
async def base_test_parse(
    model: AI_MODELS,
    client_type: ClientType,
    response_mode: ResponseModeType,
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    table_parsing_format: TableParsingFormat = "html",
    image_resolution_dpi: int = 96,
) -> None:
    document = booking_confirmation_file_path_1
    response: ParseResult | None = None

    if client_type == "sync":
        with sync_client as client:
            response = client.documents.parse(
                document=document,
                model=model,
                table_parsing_format=table_parsing_format,
                image_resolution_dpi=image_resolution_dpi,
            )
    elif client_type == "async":
        async with async_client:
            response = await async_client.documents.parse(
                document=document,
                model=model,
                table_parsing_format=table_parsing_format,
                image_resolution_dpi=image_resolution_dpi,
            )
    
    validate_parse_response(response)


@pytest.mark.asyncio
@pytest.mark.parametrize("table_format", get_args(TableParsingFormat))
async def test_parse_table_formats(
    sync_client: Retab,
    booking_confirmation_file_path_1: str,
    table_format: TableParsingFormat,
) -> None:
    """Test parsing with different table formats."""
    with sync_client as client:
        response = client.documents.parse(
            document=booking_confirmation_file_path_1,
            model="retab-micro",
            table_parsing_format=table_format,
        )
    
    validate_parse_response(response)
    # Verify that the response contains some text content
    assert len(response.text) > 0, f"Should have text content for table format {table_format}"


@pytest.mark.asyncio
@pytest.mark.parametrize("dpi", [96, 128, 196])
async def test_parse_image_resolution(
    sync_client: Retab,
    booking_confirmation_file_path_1: str,
    dpi: int,
) -> None:
    """Test parsing with different image resolutions."""
    with sync_client as client:
        response = client.documents.parse(
            document=booking_confirmation_file_path_1,
            model="retab-micro",
            image_resolution_dpi=dpi,
        )
    
    validate_parse_response(response)
    # Higher DPI might result in better text extraction, but we just verify it works
    assert len(response.text) > 0, f"Should have text content for DPI {dpi}"



@pytest.mark.asyncio
@pytest.mark.parametrize("request_number", range(5))
async def test_parse_overload(
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
    request_number: int,
) -> None:
    """Test multiple concurrent parse requests to verify system stability."""
    await asyncio.sleep(request_number * 0.1)
    await base_test_parse(
        model="retab-micro",
        client_type="async",
        response_mode="parse",
        sync_client=sync_client,
        async_client=async_client,
        booking_confirmation_file_path_1=booking_confirmation_file_path_1,
    )


@pytest.mark.asyncio
async def test_parse_response_structure(
    sync_client: Retab,
    booking_confirmation_file_path_1: str,
) -> None:
    """Test that parse response has the expected structure and data types."""
    with sync_client as client:
        response = client.documents.parse(
            document=booking_confirmation_file_path_1,
            model="retab-micro",
        )
    
    # Validate basic structure
    validate_parse_response(response)
    
    # Additional specific validations
    assert hasattr(response, 'document'), "Response should have document attribute"
    assert hasattr(response, 'usage'), "Response should have usage attribute"
    assert hasattr(response, 'pages'), "Response should have pages attribute"
    assert hasattr(response, 'text'), "Response should have text attribute"
    
    # Validate document metadata
    assert hasattr(response.document, 'mime_type'), "Document should have mime_type"
    
    # Validate usage information
    assert isinstance(response.usage.page_count, int), "Page count should be an integer"
    assert isinstance(response.usage.credits, (int, float)), "Credits should be a number"
    
    # Validate that pages and text are consistent
    # The text might be formatted differently than just joining pages, but should contain similar content
    assert len(response.text) > 0, "Text should not be empty"
    assert len(response.pages) == response.usage.page_count, "Number of pages should match page count in usage"
