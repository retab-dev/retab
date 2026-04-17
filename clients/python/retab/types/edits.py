from __future__ import annotations

import datetime
from enum import Enum
from typing import Optional

from pydantic import BaseModel, Field

from .documents.usage import RetabUsage
from .mime import FileRef, MIMEData


class BBox(BaseModel):
    left: float = Field(
        ...,
        description="Left coordinate of the bounding box, relative to page width (0.0 = left edge, 1.0 = right edge).",
    )
    top: float = Field(
        ...,
        description="Top coordinate of the bounding box, relative to page height (0.0 = top edge, 1.0 = bottom edge).",
    )
    width: float = Field(..., description="Width of the bounding box, relative to page width (0.0–1.0).")
    height: float = Field(..., description="Height of the bounding box, relative to page height (0.0–1.0).")
    page: int = Field(..., description="1-based index of the page where this field appears.")


class FieldType(str, Enum):
    TEXT = "text"
    CHECKBOX = "checkbox"


class BaseFormField(BaseModel):
    bbox: BBox = Field(..., description="Position and size of this field on the page.")
    description: str = Field(..., description="Human-readable description of the field, including label and instructions.")
    type: FieldType = Field(..., description="Type of field. Currently supported: 'text' and 'checkbox'.")
    key: str = Field(..., description="Stable key identifying the field in the form data.")


class FormField(BaseFormField):
    value: Optional[str] = Field(
        default=None,
        description="Filled value of the field as text. Null when no filled value is set.",
    )


class FormSchema(BaseModel):
    form_fields: list[FormField] = Field(..., description="Ordered list of form fields defining the structure of the form.")


class EditConfig(BaseModel):
    color: str = Field(default="#000080", description="Hex code of the color to use for the filled text.")


class EditRequest(BaseModel):
    instructions: str = Field(..., description="Instructions describing how to fill the form fields.")
    document: MIMEData | None = Field(
        default=None,
        description="Input document (PDF, DOCX, XLSX, or PPTX). Mutually exclusive with template_id.",
    )
    template_id: str | None = Field(
        default=None,
        description="EditTemplate id to fill. When provided, uses the template's pre-defined form fields and empty PDF. Mutually exclusive with document.",
    )
    model: str = Field(default="retab-small", description="The model to use for edit inference.")
    config: EditConfig = Field(default_factory=EditConfig, description="Edit configuration (rendering options).")
    bust_cache: bool = Field(default=False, description="If true, skip the LLM cache and force a fresh completion.")


class EditResult(BaseModel):
    form_data: list[FormField] = Field(..., description="Filled form fields (positions, descriptions, and filled values).")
    filled_document: MIMEData = Field(..., description="PDF with the filled form values rendered in.")


class Edit(BaseModel):
    id: str = Field(..., description="Unique identifier of the edit.")
    file: FileRef = Field(..., description="Information about the source file (input document or template PDF).")
    model: str = Field(..., description="Model used for the edit operation.")
    instructions: str = Field(..., description="Instructions supplied with the edit request.")
    config: EditConfig = Field(..., description="Configuration used for the edit operation.")
    template_id: str | None = Field(
        default=None,
        description="Template id used when the edit was created from a template; null for direct-document edits.",
    )
    data: EditResult = Field(..., description="The edit result: filled form fields and the rendered PDF.")
    usage: RetabUsage = Field(..., description="Usage information for the edit operation.")
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None


class EditTemplateRequest(BaseModel):
    name: str = Field(..., description="Name of the template.")
    document: MIMEData = Field(..., description="The PDF document to use as the empty template.")
    form_fields: list[FormField] = Field(..., description="Form fields to attach to the template.")


class UpdateEditTemplateRequest(BaseModel):
    name: Optional[str] = Field(default=None, description="New name for the template.")
    form_fields: Optional[list[FormField]] = Field(default=None, description="Replacement list of form fields.")


class DuplicateEditTemplateRequest(BaseModel):
    name: Optional[str] = Field(
        default=None,
        description="Name for the duplicated template. Defaults to '<original> (copy)'.",
    )


class EditTemplate(BaseModel):
    id: str = Field(..., description="Unique identifier of the template.")
    name: str = Field(..., description="Name of the template.")
    file: FileRef = Field(..., description="File information for the empty PDF template.")
    form_fields: list[FormField] = Field(default_factory=list, description="Form fields attached to the template.")
    field_count: int = Field(default=0, description="Number of form fields in the template.")
    organization_id: Optional[str] = Field(default=None, description="Organization that owns this template.")
    created_at: datetime.datetime = Field(..., description="Timestamp of creation.")
    updated_at: datetime.datetime = Field(..., description="Timestamp of last update.")
