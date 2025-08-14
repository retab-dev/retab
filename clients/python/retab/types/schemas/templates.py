import datetime
from typing import Any, Literal, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, Field, PrivateAttr

from ...types.mime import MIMEData


class TemplateSchema(BaseModel):
    """A full Schema object with validation."""

    id: str = Field(default_factory=lambda: f"tplt_{nanoid.generate()}")

    name: str
    """The name of the template."""

    object: Literal["template"] = "template"
    """The type of object being preprocessed."""

    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    """The timestamp of when the schema was created."""

    json_schema: dict[str, Any] = {}
    """The JSON schema to use for loading."""

    python_code: Optional[str] = None
    """The Python code to use for creating the Schema."""

    sample_document_filename: Optional[str] = None
    """The filename of the sample document to use for creating the Schema."""



    pydantic_model: type[BaseModel] = Field(default=None, exclude=True, repr=False)  # type: ignore

    _partial_pydantic_model: type[BaseModel] = PrivateAttr()
    """The Pydantic model to use for loading."""


class UpdateTemplateRequest(BaseModel):
    """Request model for updating a template."""

    id: str
    """The ID of the template to update."""

    name: Optional[str] = None
    """The new name of the template."""

    json_schema: Optional[dict[str, Any]] = None
    """The new JSON schema to use for loading."""

    python_code: Optional[str] = None
    """The new Python code to use for creating the Schema."""

    sample_document: Optional[MIMEData] = None
    """The new sample document to use for creating the Schema."""
