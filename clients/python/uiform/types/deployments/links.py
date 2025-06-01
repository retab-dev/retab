import copy
import datetime
import json
from typing import Any, Dict, Literal, Optional

import nanoid  # type: ignore
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, HttpUrl, computed_field, field_serializer
from pydantic_core import Url

from ..._utils.json_schema import clean_schema
from ..._utils.mime import generate_blake2b_hash_from_string
from ..logs import AutomationConfig, UpdateAutomationRequest
from ..modalities import Modality
from ..pagination import ListMetadata


class Link(AutomationConfig):
    object: Literal['deployment.link'] = "deployment.link"
    id: str = Field(default_factory=lambda: "lnk_" + nanoid.generate(), description="Unique identifier for the extraction link")

    # Link Specific Config
    name: str = Field(..., description="Name of the link")
    password: Optional[str] = Field(None, description="Password to access the link")

    # Automation Config
    webhook_url: HttpUrl = Field(..., description="Url of the webhook to send the data to")
    webhook_headers: Dict[str, str] = Field(default_factory=dict, description="Headers to send with the request")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    # DocumentExtraction Config
    modality: Modality
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: Literal['A3', 'A4', 'A5'] = Field(default='A4', description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type.")
    model: str = Field(..., description="Model used for chat completion")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])
    reasoning_effort: ChatCompletionReasoningEffort = Field(
        default="medium", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )
    n_consensus: int = Field(default=1, description="Number of consensus required to validate the data")

    @computed_field  # type: ignore
    @property
    def schema_data_id(self) -> str:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.

        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        return "sch_data_id_" + generate_blake2b_hash_from_string(
            json.dumps(
                clean_schema(
                    copy.deepcopy(self.json_schema),
                    remove_custom_fields=True,
                    fields_to_remove=["description", "default", "title", "required", "examples", "deprecated", "readOnly", "writeOnly"],
                ),
                sort_keys=True,
            ).strip()
        )

    # This is a computed field, it is exposed when serializing the object
    @computed_field  # type: ignore
    @property
    def schema_id(self) -> str:
        """Returns the SHA1 hash of the complete schema.

        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        return "sch_id_" + generate_blake2b_hash_from_string(json.dumps(self.json_schema, sort_keys=True).strip())

    @field_serializer('webhook_url')
    def url2str(self, val: HttpUrl) -> str:
        return str(val)


class ListLinks(BaseModel):
    data: list[Link]
    list_metadata: ListMetadata


# Inherits from the methods of UpdateAutomationRequest
class UpdateLinkRequest(UpdateAutomationRequest):
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
    image_resolution_dpi: Optional[int] = None
    browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None
    modality: Optional[Modality] = None
    # Others DocumentExtraction Parameters
    model: Optional[str] = None
    temperature: Optional[float] = None
    json_schema: Optional[Dict] = None
    reasoning_effort: Optional[ChatCompletionReasoningEffort] = None

    @field_serializer('webhook_url')
    def url2str(self, val: HttpUrl | None) -> str | None:
        if isinstance(val, HttpUrl):
            return str(val)
        return val
