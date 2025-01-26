from typing import Literal, Dict, Any, Optional
from pydantic import BaseModel, Field, computed_field
import datetime
import uuid
import copy
import json

from ..._utils.json_schema import clean_schema
from ..._utils.mime import generate_sha_hash_from_string

class Dataset(BaseModel):
    object: Literal["dataset"] = "dataset"
    id: str = Field(default_factory=lambda: f"ds_{uuid.uuid4()}", description="Unique identifier of the dataset")
    name: str = Field(description="Name of the dataset")
    description: Optional[str] = Field(default=None, description="Optional description of the dataset")
    collection_id: str = Field(description="ID of the collection this dataset belongs to")
    json_schema: Dict[str, Any] = Field(default_factory=dict, description="JSON schema defining the structure of annotations in this dataset")
    created_at: datetime.datetime = Field(default_factory=datetime.datetime.now, description="Timestamp for when the dataset was created")
    updated_at: datetime.datetime = Field(default_factory=datetime.datetime.now, description="Timestamp for when the dataset was last updated")
    
    @computed_field   # type: ignore
    @property
    def schema_data_version(self) -> str:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.
        
        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        return generate_sha_hash_from_string(
            json.dumps(
                clean_schema(copy.deepcopy(self.json_schema), remove_custom_fields=True, fields_to_remove=["description", "default", "title", "required", "examples", "deprecated", "readOnly", "writeOnly"]),
                sort_keys=True).strip(), 
            "sha1")

    # This is a computed field, it is exposed when serializing the object
    @computed_field   # type: ignore
    @property
    def schema_version(self) -> str:
        """Returns the SHA1 hash of the complete schema.
        
        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        return generate_sha_hash_from_string(json.dumps(self.json_schema, sort_keys=True).strip(), "sha1")
    
