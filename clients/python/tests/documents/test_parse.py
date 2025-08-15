import asyncio
import time
from typing import Any, Literal, get_args

import httpx
import nanoid  # type: ignore
import pytest

from retab import AsyncRetab, Retab
from retab.types.documents.parse import ParseResult, TableParsingFormat
from retab.types.browser_canvas import BrowserCanvas

# List of AI Models to test (focusing on models that support parsing)
AI_MODELS = Literal[
    "gemini-2.5-flash",
    "gpt-4.1-mini",
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
    browser_canvas: BrowserCanvas = "A4",
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
                browser_canvas=browser_canvas,
            )
    elif client_type == "async":
        async with async_client:
            response = await async_client.documents.parse(
                document=document,
                model=model,
                table_parsing_format=table_parsing_format,
                image_resolution_dpi=image_resolution_dpi,
                browser_canvas=browser_canvas,
            )
    
    validate_parse_response(response)


@pytest.mark.asyncio
@pytest.mark.xdist_group(name="gemini")
@pytest.mark.parametrize("client_type", get_args(ClientType))
@pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
async def test_parse_gemini(
    client_type: ClientType,
    response_mode: ResponseModeType,
    sync_client: Retab,
    async_client: AsyncRetab,
    booking_confirmation_file_path_1: str,
) -> None:
    """Test document parsing with Gemini models."""
    await base_test_parse(
        model="gemini-2.5-flash",
        client_type=client_type,
        response_mode=response_mode,
        sync_client=sync_client,
        async_client=async_client,
        booking_confirmation_file_path_1=booking_confirmation_file_path_1,
    )


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
            model="gemini-2.5-flash",
            table_parsing_format=table_format,
        )
    
    validate_parse_response(response)
    # Verify that the response contains some text content
    assert len(response.text) > 0, f"Should have text content for table format {table_format}"


@pytest.mark.asyncio
@pytest.mark.parametrize("dpi", [72, 96, 150, 300])
async def test_parse_image_resolution(
    sync_client: Retab,
    booking_confirmation_file_path_1: str,
    dpi: int,
) -> None:
    """Test parsing with different image resolutions."""
    with sync_client as client:
        response = client.documents.parse(
            document=booking_confirmation_file_path_1,
            model="gemini-2.5-flash",
            image_resolution_dpi=dpi,
        )
    
    validate_parse_response(response)
    # Higher DPI might result in better text extraction, but we just verify it works
    assert len(response.text) > 0, f"Should have text content for DPI {dpi}"


@pytest.mark.asyncio
@pytest.mark.parametrize("canvas", ["A4", "A3", "A5"])
async def test_parse_browser_canvas(
    sync_client: Retab,
    booking_confirmation_file_path_1: str,
    canvas: BrowserCanvas,
) -> None:
    """Test parsing with different browser canvas sizes."""
    with sync_client as client:
        response = client.documents.parse(
            document=booking_confirmation_file_path_1,
            model="gemini-2.5-flash",
            browser_canvas=canvas,
        )
    
    validate_parse_response(response)
    assert len(response.text) > 0, f"Should have text content for canvas {canvas}"


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
        model="gemini-2.5-flash",
        client_type="async",
        response_mode="parse",
        sync_client=sync_client,
        async_client=async_client,
        booking_confirmation_file_path_1=booking_confirmation_file_path_1,
    )


@pytest.mark.asyncio
async def test_parse_with_idempotency(
    sync_client: Retab, 
    booking_confirmation_file_path_1: str,
) -> None:
    """Test that parse requests with the same idempotency key return identical results."""
    idempotency_key = nanoid.generate()
    model = "gemini-2.5-flash"

    # First request
    response_initial = sync_client.documents.parse(
        document=booking_confirmation_file_path_1,
        model=model,
        table_parsing_format="html",
        image_resolution_dpi=96,
        browser_canvas="A4",
        idempotency_key=idempotency_key,
    )
    
    await asyncio.sleep(2)
    t0 = time.time()
    
    # Second request with same idempotency key
    response_second = sync_client.documents.parse(
        document=booking_confirmation_file_path_1,
        model=model,
        table_parsing_format="html",
        image_resolution_dpi=96,
        browser_canvas="A4",
        idempotency_key=idempotency_key,
    )
    
    t1 = time.time()
    assert t1 - t0 < 10, "Request should take less than 10 seconds"
    
    # Verify responses are identical
    assert response_initial.text == response_second.text, "Response text should be the same"
    assert response_initial.pages == response_second.pages, "Response pages should be the same"
    assert response_initial.usage.page_count == response_second.usage.page_count, "Page count should be the same"
    assert response_initial.usage.credits == response_second.usage.credits, "Credits should be the same"


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "error_scenario",
    ["invalid_model", "missing_document", "invalid_dpi", "invalid_canvas"],
)
async def test_parse_with_idempotency_error_scenarios(
    sync_client: Retab, 
    booking_confirmation_file_path_1: str, 
    error_scenario: str,
) -> None:
    """Test idempotency behavior when various error conditions occur during parsing."""
    idempotency_key = str(nanoid.generate())
    
    # Set up error scenarios
    if error_scenario == "invalid_model":
        model = "invalid-model-name"  # type: ignore
        document = booking_confirmation_file_path_1
        table_parsing_format = "html"
        image_resolution_dpi = 96
        browser_canvas = "A4"
    elif error_scenario == "missing_document":
        model = "gemini-2.5-flash"
        document = "/nonexistent/file.pdf"
        table_parsing_format = "html"
        image_resolution_dpi = 96
        browser_canvas = "A4"
    elif error_scenario == "invalid_dpi":
        model = "gemini-2.5-flash"
        document = booking_confirmation_file_path_1
        table_parsing_format = "html"
        image_resolution_dpi = -1  # Invalid DPI
        browser_canvas = "A4"
    elif error_scenario == "invalid_canvas":
        model = "gemini-2.5-flash"
        document = booking_confirmation_file_path_1
        table_parsing_format = "html"
        image_resolution_dpi = 96
        browser_canvas = "InvalidCanvas"  # type: ignore
    else:
        pytest.fail(f"Unknown error scenario: {error_scenario}")
    
    response1: Any = None
    response2: Any = None
    raised_exception_1: Exception | None = None
    raised_exception_2: Exception | None = None
    
    # First request attempt
    try:
        response1 = sync_client.documents.parse(
            document=document,
            model=model,  # type: ignore
            table_parsing_format=table_parsing_format,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,  # type: ignore
            idempotency_key=idempotency_key,
        )
    except Exception as e:
        raised_exception_1 = e

    await asyncio.sleep(2)
    t0 = time.time()
    
    # Second request attempt with same idempotency key
    try:
        response2 = sync_client.documents.parse(
            document=document,
            model=model,  # type: ignore
            table_parsing_format=table_parsing_format,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,  # type: ignore
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
        validate_parse_response(response1)
        validate_parse_response(response2)
        assert response1.text == response2.text, "Response text should be identical"
        assert response1.pages == response2.pages, "Response pages should be identical"


@pytest.mark.asyncio
async def test_parse_response_structure(
    sync_client: Retab,
    booking_confirmation_file_path_1: str,
) -> None:
    """Test that parse response has the expected structure and data types."""
    with sync_client as client:
        response = client.documents.parse(
            document=booking_confirmation_file_path_1,
            model="gemini-2.5-flash",
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
    assert hasattr(response.document, 'size'), "Document should have size"
    
    # Validate usage information
    assert isinstance(response.usage.page_count, int), "Page count should be an integer"
    assert isinstance(response.usage.credits, (int, float)), "Credits should be a number"
    
    # Validate that pages and text are consistent
    # The text might be formatted differently than just joining pages, but should contain similar content
    assert len(response.text) > 0, "Text should not be empty"
    assert len(response.pages) == response.usage.page_count, "Number of pages should match page count in usage"


# Commented out tests for other providers - can be enabled when those models support parsing
# @pytest.mark.asyncio
# @pytest.mark.xdist_group(name="openai")
# @pytest.mark.parametrize("client_type", get_args(ClientType))
# @pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
# async def test_parse_openai(
#     client_type: ClientType,
#     response_mode: ResponseModeType,
#     sync_client: Retab,
#     async_client: AsyncRetab,
#     booking_confirmation_file_path_1: str,
# ) -> None:
#     """Test document parsing with OpenAI models."""
#     await base_test_parse(
#         model="gpt-4o-mini",
#         client_type=client_type,
#         response_mode=response_mode,
#         sync_client=sync_client,
#         async_client=async_client,
#         booking_confirmation_file_path_1=booking_confirmation_file_path_1,
#     )

# @pytest.mark.asyncio
# @pytest.mark.xdist_group(name="anthropic")
# @pytest.mark.parametrize("client_type", get_args(ClientType))
# @pytest.mark.parametrize("response_mode", get_args(ResponseModeType))
# async def test_parse_anthropic(
#     client_type: ClientType,
#     response_mode: ResponseModeType,
#     sync_client: Retab,
#     async_client: AsyncRetab,
#     booking_confirmation_file_path_1: str,
# ) -> None:
#     """Test document parsing with Anthropic models."""
#     await base_test_parse(
#         model="claude-3-5-sonnet-latest",
#         client_type=client_type,
#         response_mode=response_mode,
#         sync_client=sync_client,
#         async_client=async_client,
#         booking_confirmation_file_path_1=booking_confirmation_file_path_1,
#     )