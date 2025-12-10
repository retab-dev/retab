from enum import Enum
from typing import List, Optional
from pydantic import BaseModel, Field

from ..mime import MIMEData


class BBox(BaseModel):
    """Bounding box of a field, in normalized coordinates."""

    left: float = Field(
        ...,
        description=(
            "Left coordinate of the bounding box, relative to page width "
            "(0.0 = left edge, 1.0 = right edge)."
        ),
    )
    top: float = Field(
        ...,
        description=(
            "Top coordinate of the bounding box, relative to page height "
            "(0.0 = top edge, 1.0 = bottom edge)."
        ),
    )
    width: float = Field(
        ...,
        description="Width of the bounding box, relative to page width (0.0–1.0).",
    )
    height: float = Field(
        ...,
        description="Height of the bounding box, relative to page height (0.0–1.0).",
    )
    page: int = Field(
        ...,
        description="1-based index of the page where this field appears in the processed document.",
    )
   


class FieldType(str, Enum):
    TEXT = "text"
    CHECKBOX = "checkbox"


class BaseFormField(BaseModel):
    """Single field in the form schema (text input, checkbox, etc.)."""

    bbox: BBox = Field(
        ...,
        description="Bounding box specifying the position and size of this field on the page.",
    )
    description: str = Field(
        ...,
        description=(
            "Human-readable description of the field, including its label, "
            "section, and instructions for how it should be filled."
        ),
    )
    type: FieldType = Field(
        ...,
        description="Type of field. Currently supported values: 'text' and 'checkbox'.",
    )
    

class FormField(BaseFormField):
    """Single field in the form schema (text input, checkbox, etc.)."""

    value: Optional[str] = Field(
        None,
        description=(
            "Current or default value of the field as text. "
            "May be null/None if no value is set."
        ),
    )

class FormSchema(BaseModel):
    """Top-level object holding the list of fields for a given form template."""

    form_fields: List[FormField] = Field(
        ...,
        description=(
            "List of form fields (with positions, descriptions, and metadata) "
            "that define the structure of the form."
        ),
    )

class FilledFormField(BaseFormField):
    """Single field in the form schema (text input, checkbox, etc.)."""

    value: Optional[str] = Field(
        None,
        description=(
            "Filled value of the field as text. "
            "May be null/None if no filled value is set."
        ),
    )


class OCRTextElement(BaseModel):
    """A single OCR-detected text element with its bounding box."""
    
    text: str = Field(..., description="The detected text content")
    bbox: BBox = Field(..., description="Bounding box in normalized coordinates")
    element_type: str = Field(..., description="Type of element: 'block', 'line', or 'token'")


class OCRResult(BaseModel):
    """Result of OCR processing on a document."""
    
    elements: list[OCRTextElement] = Field(default_factory=list, description="All detected text elements")
    formatted_text: str = Field(..., description="Human-readable formatted string of all elements")
    annotated_pdf: MIMEData = Field(..., description="PDF with bbox annotations")


class InferFormSchemaRequest(BaseModel):
    """Request to infer form schema from a PDF."""
    
    document: MIMEData = Field(..., description="Input document (PDF)")
    model: str = Field(default="gemini-2.5-pro", description="LLM model to use for inference")


class InferFormSchemaResponse(BaseModel):
    """Response containing the inferred form schema."""
    
    form_schema: FormSchema = Field(..., description="The inferred form schema")
    ocr_result: OCRResult = Field(..., description="The OCR results used for inference")
    form_fields_pdf: MIMEData = Field(..., description="PDF with form field bounding boxes")


class EditRequest(InferFormSchemaRequest):
    """Request for the infer_and_fill_schema endpoint."""
    
    filling_instructions: str = Field(..., description="Instructions to fill the form")

class EditResponse(BaseModel):
    """Response from the fill_form endpoint.
    """
    ocr_result: OCRResult = Field(..., description="The OCR results used for inference")
    form_data: list[FilledFormField] = Field(
        ...,
        description=(
            "List of form fields (with positions, descriptions, and metadata) "
            "that define the structure of the form."
        ),
    )
    filled_pdf: MIMEData = Field(..., description="PDF with filled form values")


class ProcessOCRRequest(BaseModel):
    """Request to process a PDF with OCR only (step 1)."""
    
    document: MIMEData = Field(..., description="Input document (PDF)")

