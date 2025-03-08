from pydantic import BaseModel, Field, computed_field, PrivateAttr
from typing import Any, Literal, Optional
import datetime

from ..._utils.json_schema import generate_schema_data_id, generate_schema_id


class TemplateSchema(BaseModel):
    """A full Schema object with validation."""

    name: str
    """The name of the template."""

    object: Literal["schema"] = "schema"
    """The type of object being preprocessed."""

    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    """The timestamp of when the schema was created."""
 
    json_schema: dict[str, Any] = {}
    """The JSON schema to use for loading."""

    python_code: Optional[str] = None
    """The Python code to use for creating the Schema."""

    sample_document_filename: Optional[str] = None
    """The filename of the sample document to use for creating the Schema."""

    # This is a computed field, it is exposed when serializing the object
    @computed_field   # type: ignore
    @property
    def data_id(self) -> str:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.
        
        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        return generate_schema_data_id(self.json_schema)

    # This is a computed field, it is exposed when serializing the object
    @computed_field   # type: ignore
    @property
    def id(self) -> str:
        """Returns the SHA1 hash of the complete schema.
        
        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        return generate_schema_id(self.json_schema)

    pydantic_model: type[BaseModel] = Field(default=None, exclude=True, repr=False)     # type: ignore

    _partial_pydantic_model: type[BaseModel] = PrivateAttr()
    """The Pydantic model to use for loading."""

   