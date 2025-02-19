from pydantic import BaseModel, Field,  HttpUrl, computed_field, field_serializer
from pydantic_core import Url
from typing import Any, Literal, Dict, Optional
import uuid
import datetime
import copy
import json

from ..image_settings import ImageSettings
from ..modalities import Modality
from ..pagination import ListMetadata

from ..._utils.json_schema import clean_schema
from ..._utils.mime import generate_sha_hash_from_string




class Link(BaseModel):
    object: Literal['link'] = "link"
    id: str = Field(default_factory=lambda: "lnk_" + str(uuid.uuid4()), description="Unique identifier for the extraction link")
    
    # Link Specific Config
    name: str = Field(..., description = "Name of the link")
    password: Optional[str] = Field(None, description = "Password to access the link")

    # Automation Config
    webhook_url: HttpUrl = Field(..., description = "Url of the webhook to send the data to")
    webhook_headers: Dict[str, str] = Field(default_factory=dict, description = "Headers to send with the request")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    # DocumentExtraction Config
    modality: Modality
    image_settings : ImageSettings = Field(default_factory=ImageSettings, description="Preprocessing operations applied to image before sending them to the llm")
    model: str = Field(..., description="Model used for chat completion")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])

    @computed_field   # type: ignore
    @property
    def schema_data_id(self) -> str:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.
        
        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        return "sch_data_id_" + generate_sha_hash_from_string(
            json.dumps(
                clean_schema(copy.deepcopy(self.json_schema), remove_custom_fields=True, fields_to_remove=["description", "default", "title", "required", "examples", "deprecated", "readOnly", "writeOnly"]),
                sort_keys=True).strip(), 
            "sha1")

    # This is a computed field, it is exposed when serializing the object
    @computed_field   # type: ignore
    @property
    def schema_id(self) -> str:
        """Returns the SHA1 hash of the complete schema.
        
        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        return "sch_id_" + generate_sha_hash_from_string(json.dumps(self.json_schema, sort_keys=True).strip(), "sha1")

    @field_serializer('webhook_url')
    def url2str(self, val) -> str:
        if isinstance(val, Url): ### This magic! If isinstance(val, HttpUrl) - error
            return str(val)
        return val

class ListLinks(BaseModel):
    data: list[Link]
    list_metadata: ListMetadata






class UpdateLinkRequest(BaseModel):
    id: str

    # ------------------------------
    # Link Config
    # ------------------------------
    name: Optional[str] = None
    password: Optional[str] = None

    # ------------------------------
    # HTTP Config
    # ------------------------------
    webhook_url: Optional[HttpUrl] = None
    webhook_headers: Optional[Dict[str, str]] = None

    # ------------------------------
    # DocumentExtraction Parameters
    # ------------------------------
    # DocumentProcessing Parameters
    image_settings: Optional[ImageSettings] = None
    modality: Optional[Modality] = None
    # Others DocumentExtraction Parameters
    model: Optional[str] = None
    temperature: Optional[float] = None
    json_schema: Optional[Dict] = None

