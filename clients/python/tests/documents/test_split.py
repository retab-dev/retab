import os
from typing import Literal, get_args

import pytest

from retab import AsyncRetab, Retab
from retab.types.documents.split import Category, SplitResponse, SplitResult

# Get the directory containing the tests
TEST_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# List of AI Models to test for split (models that support vision/document analysis)
AI_MODELS = Literal[
    "gemini-2.5-flash",
]

ClientType = Literal[
    "sync",
    "async",
]


@pytest.fixture(scope="session")
def multi_page_pdf_path() -> str:
    """Return the path to a multi-page PDF for testing."""
    return os.path.join(TEST_DIR, "data", "edit", "fidelity_original.pdf")


@pytest.fixture(scope="session")
def split_categories() -> list[Category]:
    """Return categories for document splitting."""
    return [
        Category(name="form_section", description="Form sections with input fields, checkboxes, or signature areas"),
        Category(name="instructions", description="Instructions, terms and conditions, or explanatory text"),
        Category(name="header", description="Header sections with logos, titles, or document identifiers"),
    ]


def validate_split_response(response: SplitResponse | None, categories: list[Category]) -> None:
    """Validate that the split response has the expected structure and content."""
    # Assert the instance
    assert isinstance(response, SplitResponse), f"Response should be of type SplitResponse, received {type(response)}"

    # Assert the response has required fields
    assert response.splits is not None, "Response splits should not be None"

    # Validate splits
    assert isinstance(response.splits, list), "splits should be a list"
    assert len(response.splits) > 0, "Should have at least one split"

    # Get valid category names
    valid_category_names = {cat.name for cat in categories}

    for split in response.splits:
        assert isinstance(split, SplitResult), f"Each split should be SplitResult, got {type(split)}"
        assert split.name is not None, "Split name should not be None"
        assert split.pages is not None, "Split pages should not be None"
        assert isinstance(split.pages, list), "Split pages should be a list"
        assert len(split.pages) > 0, "Split pages should not be empty"

        # Validate category name is from provided categories
        assert split.name in valid_category_names, f"Split name '{split.name}' should be one of {valid_category_names}"

        # Validate page numbers
        for page in split.pages:
            assert page >= 1, "All pages should be >= 1 (1-indexed)"


def validate_splits_are_ordered(response: SplitResponse) -> None:
    """Validate that splits are ordered by first page."""
    if len(response.splits) <= 1:
        return

    for i in range(1, len(response.splits)):
        prev_first_page = response.splits[i-1].pages[0] if response.splits[i-1].pages else 0
        curr_first_page = response.splits[i].pages[0] if response.splits[i].pages else 0
        assert curr_first_page >= prev_first_page, \
            f"Splits should be ordered by first page: {response.splits[i-1]} should come before {response.splits[i]}"


def validate_splits_non_overlapping(response: SplitResponse) -> None:
    """Validate that splits do not overlap (each page belongs to exactly one section)."""
    if len(response.splits) <= 1:
        return

    # Collect all pages across all splits
    all_pages: list[int] = []
    for split in response.splits:
        all_pages.extend(split.pages)

    # Check for duplicates - each page should only appear once
    seen_pages = set()
    for page in all_pages:
        assert page not in seen_pages, f"Page {page} appears in multiple splits"
        seen_pages.add(page)


# Base test function for split endpoint
async def base_test_split(
    model: str,
    client_type: ClientType,
    sync_client: Retab,
    async_client: AsyncRetab,
    document_path: str,
    categories: list[Category],
) -> SplitResponse:
    response: SplitResponse

    if client_type == "sync":
        with sync_client as client:
            response = client.documents.split(
                document=document_path,
                model=model,
                categories=categories,
            )
    elif client_type == "async":
        async with async_client:
            response = await async_client.documents.split(
                document=document_path,
                model=model,
                categories=categories,
            )
    
    validate_split_response(response, categories)
    return response


@pytest.mark.asyncio
@pytest.mark.xdist_group(name="gemini")
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_split_document(
    client_type: ClientType,
    sync_client: Retab,
    async_client: AsyncRetab,
    multi_page_pdf_path: str,
    split_categories: list[Category],
) -> None:
    """Test document splitting with Gemini models."""
    response = await base_test_split(
        model="gemini-2.5-flash",
        client_type=client_type,
        sync_client=sync_client,
        async_client=async_client,
        document_path=multi_page_pdf_path,
        categories=split_categories,
    )
    
    # Additional validations
    validate_splits_are_ordered(response)


@pytest.mark.asyncio
async def test_split_response_structure(
    sync_client: Retab,
    multi_page_pdf_path: str,
    split_categories: list[Category],
) -> None:
    """Test that split response has the expected structure and data types."""
    with sync_client as client:
        response = client.documents.split(
            document=multi_page_pdf_path,
            model="gemini-2.5-flash",
            categories=split_categories,
        )
    
    # Validate basic structure
    validate_split_response(response, split_categories)
    
    # Additional specific validations
    assert hasattr(response, 'splits'), "Response should have splits attribute"
    
    # Validate splits have proper structure
    for split in response.splits:
        assert hasattr(split, 'name'), "Split should have name"
        assert hasattr(split, 'pages'), "Split should have pages"
        assert hasattr(split, 'partitions'), "Split should have partitions"

        # Validate types
        assert isinstance(split.name, str), "name should be a string"
        assert isinstance(split.pages, list), "pages should be a list"
        assert all(isinstance(p, int) for p in split.pages), "all pages should be integers"
        assert isinstance(split.partitions, list), "partitions should be a list"


@pytest.mark.asyncio
async def test_split_with_dict_categories(
    sync_client: Retab,
    multi_page_pdf_path: str,
) -> None:
    """Test splitting with categories provided as dictionaries."""
    dict_categories = [
        {"name": "content", "description": "Main content sections"},
        {"name": "footer", "description": "Footer sections with legal text or page numbers"},
    ]
    
    with sync_client as client:
        response = client.documents.split(
            document=multi_page_pdf_path,
            model="gemini-2.5-flash",
            categories=dict_categories,
        )
    
    # Validate response
    assert isinstance(response, SplitResponse), "Response should be SplitResponse"
    assert len(response.splits) > 0, "Should have at least one split"
    
    # Validate category names
    valid_names = {"content", "footer"}
    for split in response.splits:
        assert split.name in valid_names, f"Split name '{split.name}' should be one of {valid_names}"


@pytest.mark.asyncio
async def test_split_page_coverage(
    sync_client: Retab,
    multi_page_pdf_path: str,
    split_categories: list[Category],
) -> None:
    """Test that splits cover the document pages appropriately."""
    with sync_client as client:
        response = client.documents.split(
            document=multi_page_pdf_path,
            model="gemini-2.5-flash",
            categories=split_categories,
        )
    
    validate_split_response(response, split_categories)

    # Validate that all page numbers are positive
    for split in response.splits:
        assert len(split.pages) >= 1, "Each split should have at least 1 page"
        for page in split.pages:
            assert page >= 1, f"All pages should be >= 1, got {page}"


@pytest.mark.asyncio
async def test_split_with_minimal_categories(
    sync_client: Retab,
    multi_page_pdf_path: str,
) -> None:
    """Test splitting with minimal category definitions."""
    minimal_categories = [
        Category(name="document", description="Document content"),
    ]
    
    with sync_client as client:
        response = client.documents.split(
            document=multi_page_pdf_path,
            model="gemini-2.5-flash",
            categories=minimal_categories,
        )
    
    # With only one category, all pages should be classified as that category
    assert isinstance(response, SplitResponse), "Response should be SplitResponse"
    assert len(response.splits) >= 1, "Should have at least one split"
    
    for split in response.splits:
        assert split.name == "document", "All splits should be 'document' category"


@pytest.mark.asyncio
async def test_split_discontinuous_categories(
    sync_client: Retab,
    multi_page_pdf_path: str,
) -> None:
    """Test that the same category can appear multiple times for non-contiguous sections."""
    categories = [
        Category(name="form_fields", description="Sections with form input fields"),
        Category(name="text_content", description="Sections with explanatory text or instructions"),
    ]
    
    with sync_client as client:
        response = client.documents.split(
            document=multi_page_pdf_path,
            model="gemini-2.5-flash",
            categories=categories,
        )
    
    validate_split_response(response, categories)
    
    # Count occurrences of each category
    category_counts: dict[str, int] = {}
    for split in response.splits:
        category_counts[split.name] = category_counts.get(split.name, 0) + 1
    
    # It's valid for a category to appear multiple times (discontinuous sections)
    # This test just verifies the response structure is correct
    total_splits = sum(category_counts.values())
    assert total_splits == len(response.splits), "All splits should be counted"
