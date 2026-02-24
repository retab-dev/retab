import os
from typing import Literal, get_args

import pytest

from retab import AsyncRetab, Retab
from retab.types.documents.split import Subdocument, SplitResponse, SplitResult, GenerateSplitConfigResponse

# Get the directory containing the tests
TEST_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# List of AI Models to test for split (models that support vision/document analysis)
AI_MODELS = Literal[
    "retab-small",
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
def split_subdocuments() -> list[Subdocument]:
    """Return subdocuments for document splitting."""
    return [
        Subdocument(name="form_section", description="Form sections with input fields, checkboxes, or signature areas"),
        Subdocument(name="instructions", description="Instructions, terms and conditions, or explanatory text"),
        Subdocument(name="header", description="Header sections with logos, titles, or document identifiers"),
    ]


def validate_split_response(response: SplitResponse | None, subdocuments: list[Subdocument]) -> None:
    """Validate that the split response has the expected structure and content."""
    # Assert the instance
    assert isinstance(response, SplitResponse), f"Response should be of type SplitResponse, received {type(response)}"

    # Assert the response has required fields
    assert response.splits is not None, "Response splits should not be None"

    # Validate splits
    assert isinstance(response.splits, list), "splits should be a list"
    assert len(response.splits) > 0, "Should have at least one split"

    # Get valid subdocument names
    valid_subdocument_names = {subdoc.name for subdoc in subdocuments}

    # Track whether at least one split has pages (not all should be empty)
    has_non_empty_split = False

    for split in response.splits:
        assert isinstance(split, SplitResult), f"Each split should be SplitResult, got {type(split)}"
        assert split.name is not None, "Split name should not be None"
        assert split.pages is not None, "Split pages should not be None"
        assert isinstance(split.pages, list), "Split pages should be a list"

        # Validate subdocument name is from provided subdocuments
        assert split.name in valid_subdocument_names, f"Split name '{split.name}' should be one of {valid_subdocument_names}"

        # Validate page numbers (only for non-empty splits)
        if len(split.pages) > 0:
            has_non_empty_split = True
            for page in split.pages:
                assert page >= 1, "All pages should be >= 1 (1-indexed)"

    # At least one split should have pages assigned
    assert has_non_empty_split, "At least one split should have pages assigned"


def validate_splits_are_ordered(response: SplitResponse) -> None:
    """Validate that non-empty splits are ordered by first page."""
    # Filter to only splits with pages for ordering validation
    non_empty_splits = [s for s in response.splits if s.pages]

    if len(non_empty_splits) <= 1:
        return

    for i in range(1, len(non_empty_splits)):
        prev_first_page = non_empty_splits[i-1].pages[0]
        curr_first_page = non_empty_splits[i].pages[0]
        assert curr_first_page >= prev_first_page, \
            f"Splits should be ordered by first page: {non_empty_splits[i-1]} should come before {non_empty_splits[i]}"


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
    subdocuments: list[Subdocument],
) -> SplitResponse:
    response: SplitResponse

    if client_type == "sync":
        with sync_client as client:
            response = client.documents.split(
                document=document_path,
                model=model,
                subdocuments=subdocuments,
            )
    elif client_type == "async":
        async with async_client:
            response = await async_client.documents.split(
                document=document_path,
                model=model,
                subdocuments=subdocuments,
            )
    
    validate_split_response(response, subdocuments)
    return response



@pytest.mark.asyncio
async def test_split_response_structure(
    sync_client: Retab,
    multi_page_pdf_path: str,
    split_subdocuments: list[Subdocument],
) -> None:
    """Test that split response has the expected structure and data types."""
    with sync_client as client:
        response = client.documents.split(
            document=multi_page_pdf_path,
            model="retab-small",
            subdocuments=split_subdocuments,
        )
    
    # Validate basic structure
    validate_split_response(response, split_subdocuments)
    
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
async def test_split_with_dict_subdocuments(
    sync_client: Retab,
    multi_page_pdf_path: str,
) -> None:
    """Test splitting with subdocuments provided as dictionaries."""
    dict_subdocuments = [
        {"name": "content", "description": "Main content sections"},
        {"name": "footer", "description": "Footer sections with legal text or page numbers"},
    ]
    
    with sync_client as client:
        response = client.documents.split(
            document=multi_page_pdf_path,
            model="retab-small",
            subdocuments=dict_subdocuments,
        )
    
    # Validate response
    assert isinstance(response, SplitResponse), "Response should be SplitResponse"
    assert len(response.splits) > 0, "Should have at least one split"
    
    # Validate subdocument names match what we provided
    valid_names = {"content", "footer"}
    for split in response.splits:
        assert split.name in valid_names, f"Split name '{split.name}' should be one of {valid_names}"


@pytest.mark.asyncio
async def test_split_page_coverage(
    sync_client: Retab,
    multi_page_pdf_path: str,
    split_subdocuments: list[Subdocument],
) -> None:
    """Test that splits cover the document pages appropriately."""
    with sync_client as client:
        response = client.documents.split(
            document=multi_page_pdf_path,
            model="retab-small",
            subdocuments=split_subdocuments,
        )
    
    validate_split_response(response, split_subdocuments)

    # Validate that all page numbers are positive (for non-empty splits)
    for split in response.splits:
        for page in split.pages:
            assert page >= 1, f"All pages should be >= 1, got {page}"


@pytest.mark.asyncio
async def test_split_with_minimal_subdocuments(
    sync_client: Retab,
    multi_page_pdf_path: str,
) -> None:
    """Test splitting with minimal subdocument definitions."""
    minimal_subdocuments = [
        Subdocument(name="document", description="Document content"),
    ]
    
    with sync_client as client:
        response = client.documents.split(
            document=multi_page_pdf_path,
            model="retab-small",
            subdocuments=minimal_subdocuments,
        )
    
    # With only one subdocument, all pages should be classified as that subdocument
    assert isinstance(response, SplitResponse), "Response should be SplitResponse"
    assert len(response.splits) >= 1, "Should have at least one split"
    
    for split in response.splits:
        assert split.name == "document", "All splits should be 'document' subdocument"


@pytest.mark.asyncio
async def test_split_discontinuous_subdocuments(
    sync_client: Retab,
    multi_page_pdf_path: str,
) -> None:
    """Test that the same subdocument can appear multiple times for non-contiguous sections."""
    discontinuous_subdocuments = [
        Subdocument(name="form_fields", description="Sections with form input fields"),
        Subdocument(name="text_content", description="Sections with explanatory text or instructions"),
    ]
    
    with sync_client as client:
        response = client.documents.split(
            document=multi_page_pdf_path,
            model="retab-small",
            subdocuments=discontinuous_subdocuments,
        )
    
    validate_split_response(response, discontinuous_subdocuments)
    
    # Count occurrences of each subdocument
    subdocument_counts: dict[str, int] = {}
    for split in response.splits:
        subdocument_counts[split.name] = subdocument_counts.get(split.name, 0) + 1
    
    # It's valid for a subdocument to appear multiple times (discontinuous sections)
    # This test just verifies the response structure is correct
    total_splits = sum(subdocument_counts.values())
    assert total_splits == len(response.splits), "All splits should be counted"


# ═══════════════════════════════════════════════════════════════════════════════
# Generate Split Config tests
# ═══════════════════════════════════════════════════════════════════════════════


@pytest.mark.asyncio
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_generate_split_config_fidelity(
    client_type: ClientType,
    sync_client: Retab,
    async_client: AsyncRetab,
    multi_page_pdf_path: str,
) -> None:
    """Test generate_split_config with the fidelity PDF."""
    response: GenerateSplitConfigResponse

    if client_type == "sync":
        with sync_client as client:
            response = client.documents.generate_split_config(
                document=multi_page_pdf_path,
                model="retab-small",
            )
    elif client_type == "async":
        async with async_client:
            response = await async_client.documents.generate_split_config(
                document=multi_page_pdf_path,
                model="retab-small",
            )

    assert isinstance(response, GenerateSplitConfigResponse)
    assert len(response.subdocuments) > 0, "Should detect at least one subdocument type"
    for subdoc in response.subdocuments:
        assert subdoc.name, "Subdocument should have a name"
        assert subdoc.description, "Subdocument should have a description"


@pytest.mark.asyncio
@pytest.mark.parametrize("pdf_path", [
    "/workspaces/monorepo/ilovepdf_merged.pdf",
    "/workspaces/monorepo/Use Case 1_Res Vendor KVFREF24774193-GA.pdf",
])
async def test_generate_split_config_test_pdfs(
    sync_client: Retab,
    pdf_path: str,
) -> None:
    """Test generate_split_config with the two test PDFs."""
    if not os.path.exists(pdf_path):
        pytest.skip(f"Test PDF not found: {pdf_path}")

    with sync_client as client:
        response = client.documents.generate_split_config(
            document=pdf_path,
            model="retab-small",
        )

    assert isinstance(response, GenerateSplitConfigResponse)
    assert len(response.subdocuments) > 0, f"Should detect subdocuments in {os.path.basename(pdf_path)}"

    print(f"\n{'=' * 60}")
    print(f"PDF: {os.path.basename(pdf_path)}")
    print(f"Subdocuments detected: {len(response.subdocuments)}")
    for subdoc in response.subdocuments:
        pk = f" (partition_key: {subdoc.partition_key})" if subdoc.partition_key else ""
        print(f"  - {subdoc.name}: {subdoc.description}{pk}")
    print(f"{'=' * 60}")
