from typing import Literal, Dict, Any, Optional, List
from pydantic import BaseModel, Field, computed_field
import datetime
import nanoid # type: ignore
import copy
import json

from ..._utils.json_schema import clean_schema
from ..._utils.mime import generate_blake2b_hash_from_string

class Dataset(BaseModel):
    """This is the base class for all datasets. It contains the common fields for all datasets. Mostly useful for default_datasets and custom_datasets."""
    object: Literal["dataset.default", "dataset.custom"]
    id: str = Field(default_factory=lambda: f"ds_{nanoid.generate()}", description="Unique identifier of the dataset")
    name: str = Field(description="Name of the dataset")
    json_schema: Optional[Dict[str, Any]] = Field(default=None, description="JSON schema defining the structure of annotations in this dataset")
    description: Optional[str] = Field(default=None, description="Optional description of the dataset")
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc), description="Timestamp for when the dataset was created")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc), description="Timestamp for when the dataset was last updated")
    
    @computed_field   # type: ignore
    @property
    def schema_data_id(self) -> str:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.
        
        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        if self.json_schema is None:
            return ""
        return "sch_data_id_"+generate_blake2b_hash_from_string(
            json.dumps(
                clean_schema(copy.deepcopy(self.json_schema), remove_custom_fields=True, fields_to_remove=["description", "default", "title", "required", "examples", "deprecated", "readOnly", "writeOnly"]),
                sort_keys=True).strip()
            )

    # This is a computed field, it is exposed when serializing the object
    @computed_field   # type: ignore
    @property
    def schema_id(self) -> str:
        """Returns the SHA1 hash of the complete schema.
        
        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        if self.json_schema is None:
            return ""
        
        return "sch_id_"+generate_blake2b_hash_from_string(json.dumps(self.json_schema, sort_keys=True).strip())

class DatasetAnnotationStatus(BaseModel):
    total_files: int
    files_with_empty_annotations: List[str]
    files_with_incomplete_annotations: List[str]
    files_with_completed_annotations: List[str]
