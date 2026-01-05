"""Types for edit templates."""

from typing import Optional
from pydantic import BaseModel, Field
import datetime

from ..mime import BaseMIMEData, MIMEData
from ..documents.edit import FormField, EditConfig


class EditTemplate(BaseModel):
    """An edit template with pre-defined form fields."""
    
    id: str = Field(..., description="Unique identifier of the template")
    name: str = Field(..., description="Name of the template")
    file: BaseMIMEData = Field(..., description="File information for the empty PDF template")
    form_fields: list[FormField] = Field(..., description="List of form fields in the template")
    organization_id: Optional[str] = Field(default=None, description="Organization that owns this template")
    created_at: datetime.datetime = Field(..., description="Timestamp of creation")
    updated_at: datetime.datetime = Field(..., description="Timestamp of last update")


class CreateEditTemplateRequest(BaseModel):
    """Request model for creating an edit template."""
    
    name: str = Field(..., description="Name of the template")
    document: MIMEData = Field(..., description="The PDF document to use as the template")
    form_fields: list[FormField] = Field(..., description="List of form fields in the template")


class UpdateEditTemplateRequest(BaseModel):
    """Request model for updating an edit template."""
    
    name: Optional[str] = Field(default=None, description="Name of the template")
    form_fields: Optional[list[FormField]] = Field(default=None, description="List of form fields")


class DuplicateEditTemplateRequest(BaseModel):
    """Request model for duplicating an edit template."""
    
    name: Optional[str] = Field(default=None, description="Name for the duplicated template")


class FillTemplateRequest(BaseModel):
    """Request for the fill endpoint.
    Uses pre-defined form fields from the template (PDF only)
    """
    model: str = Field(default="retab-small", description="LLM model to use for inference")
    instructions: str = Field(..., description="Instructions to fill the form")
    template_id: str = Field(..., description="Template ID to use for filling. When provided, uses the template's pre-defined form fields and empty PDF. Only works for PDF documents. Mutually exclusive with document.")
    config: EditConfig = Field(default_factory=EditConfig, description="Configuration for the fill request")
