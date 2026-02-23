import json
import os
from typing import Literal, get_args

import pytest

from retab import AsyncRetab, Retab
from retab.types.documents.edit import EditResponse, FormField, OCRResult
from retab.types.mime import MIMEData

# Get the directory containing the tests
TEST_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# List of AI Models to test for edit (models that support form filling)
AI_MODELS = Literal[
    "retab-large",
    "retab-small",
]

ClientType = Literal[
    "sync",
    "async",
]


@pytest.fixture(scope="session")
def fidelity_form_path() -> str:
    """Return the path to the Fidelity form PDF."""
    return os.path.join(TEST_DIR, "data", "edit", "fidelity_original.pdf")


@pytest.fixture(scope="session")
def fidelity_instructions() -> str:
    """Return the filling instructions for the Fidelity form."""
    instructions_path = os.path.join(TEST_DIR, "data", "edit", "instructions.txt")
    with open(instructions_path) as f:
        return f.read()


def validate_edit_response(response: EditResponse | None) -> None:
    """Validate that the edit response has the expected structure and content."""
    # Assert the instance
    assert isinstance(response, EditResponse), f"Response should be of type EditResponse, received {type(response)}"
    
    # Assert the response has required fields
    assert response.form_data is not None, "Response form_data should not be None"
    
    # Validate form_data
    assert isinstance(response.form_data, list), "form_data should be a list"
    assert len(response.form_data) > 0, "Should have at least one form field"
    
    for field in response.form_data:
        assert isinstance(field, FormField), f"Each field should be FormField, got {type(field)}"
        assert field.bbox is not None, "Field bbox should not be None"
        assert field.description is not None, "Field description should not be None"
        assert field.type is not None, "Field type should not be None"
        # value can be None for unfilled fields
        
        # Validate bbox
        assert 0 <= field.bbox.left <= 1, "bbox left should be normalized (0-1)"
        assert 0 <= field.bbox.top <= 1, "bbox top should be normalized (0-1)"
        assert 0 <= field.bbox.width <= 1, "bbox width should be normalized (0-1)"
        assert 0 <= field.bbox.height <= 1, "bbox height should be normalized (0-1)"
        assert field.bbox.page >= 1, "bbox page should be >= 1"
    
    # Validate filled_document
    if response.filled_document is not None:
        assert isinstance(response.filled_document, MIMEData), "filled_document should be MIMEData"
        assert response.filled_document.filename is not None, "filled_document filename should not be None"
        assert response.filled_document.url.startswith("data:application/pdf;base64,"), "filled_document url should be a PDF data URI"


def validate_fidelity_form_fields(response: EditResponse, instructions: dict) -> None:
    """Validate that specific Fidelity form fields were filled correctly."""
    # Check that some expected values from the instructions are present in the filled fields
    # Account owners
    for owner in instructions.get("account_owners", []):
        # The name might appear in various forms (split, uppercase, etc.)
        owner_parts = owner.upper().split()
        assert any(
            any(part in (field.value or "").upper() for part in owner_parts)
            for field in response.form_data if field.value
        ), "Expected to find account owner name parts from '{}' in filled fields".format(owner)
    
    # Check for at least one account number
    account_numbers = instructions.get("fidelity_accounts", [])
    if account_numbers:
        assert any(
            any(acc in (field.value or "") for acc in account_numbers)
            for field in response.form_data if field.value
        ), "Expected to find at least one account number in filled fields"


# Base test function for edit endpoint
async def base_test_edit(
    model: str,
    client_type: ClientType,
    sync_client: Retab,
    async_client: AsyncRetab,
    document_path: str,
    instructions: str,
) -> EditResponse:
    response: EditResponse 

    if client_type == "sync":
        with sync_client as client:
            response = client.documents.edit(
                document=document_path,
                instructions=instructions,
                model=model,
            )
    elif client_type == "async":
        async with async_client:
            response = await async_client.documents.edit(
                document=document_path,
                instructions=instructions,
                model=model,
            )
    
    validate_edit_response(response)
    return response



@pytest.mark.asyncio
async def test_edit_response_structure(
    sync_client: Retab,
    fidelity_form_path: str,
    fidelity_instructions: str,
) -> None:
    """Test that edit response has the expected structure and data types."""
    with sync_client as client:
        response = client.documents.edit(
            document=fidelity_form_path,
            instructions=fidelity_instructions,
            model="retab-small",
        )
    
    # Validate basic structure
    validate_edit_response(response)
    
    # Additional specific validations
    assert hasattr(response, 'form_data'), "Response should have form_data attribute"
    assert hasattr(response, 'filled_document'), "Response should have filled_document attribute"
    
    # Validate form_data has fields with proper structure
    for field in response.form_data:
        assert hasattr(field, 'bbox'), "Field should have bbox"
        assert hasattr(field, 'description'), "Field should have description"
        assert hasattr(field, 'type'), "Field should have type"
        assert hasattr(field, 'value'), "Field should have value"



@pytest.mark.asyncio
async def test_edit_filled_document_is_valid(
    sync_client: Retab,
    fidelity_form_path: str,
    fidelity_instructions: str,
) -> None:
    """Test that the filled PDF is a valid PDF."""
    import base64
    
    with sync_client as client:
        response = client.documents.edit(
            document=fidelity_form_path,
            instructions=fidelity_instructions,
            model="retab-small",
        )
    
    validate_edit_response(response)
    
    # Extract and validate PDF content
    assert response.filled_document is not None, "Should have a filled PDF"
    
    # Extract base64 content from data URI
    base64_content = response.filled_document.url.split(",")[1]
    pdf_bytes = base64.b64decode(base64_content)
    
    # Check PDF magic bytes (PDF files start with %PDF-)
    assert pdf_bytes[:5] == b"%PDF-", "Filled PDF should be a valid PDF file"
    assert len(pdf_bytes) > 1000, "Filled PDF should have substantial content"


@pytest.mark.asyncio
async def test_edit_form_data_has_filled_values(
    sync_client: Retab,
    fidelity_form_path: str,
    fidelity_instructions: str,
) -> None:
    """Test that at least some form fields have filled values."""
    with sync_client as client:
        response = client.documents.edit(
            document=fidelity_form_path,
            instructions=fidelity_instructions,
            model="retab-small",
        )
    
    validate_edit_response(response)
    
    # Count fields with values
    filled_fields = [field for field in response.form_data if field.value]
    
    assert len(filled_fields) > 0, "Should have at least one filled field"
    
    # The Fidelity form has multiple fields, expect a reasonable number to be filled
    total_fields = len(response.form_data)
    fill_ratio = len(filled_fields) / total_fields if total_fields > 0 else 0
    
    # At least 20% of fields should be filled with the provided instructions
    assert fill_ratio >= 0.2, f"Expected at least 20% of fields to be filled, got {fill_ratio:.1%}"
