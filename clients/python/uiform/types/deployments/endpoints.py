import copy
import datetime
import json
from typing import Any, Literal, Optional

import nanoid  # type: ignore
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, HttpUrl, computed_field, field_serializer

from ..._utils.json_schema import clean_schema
from ..._utils.mime import generate_blake2b_hash_from_string
from ..logs import AutomationConfig, UpdateAutomationRequest
from ..modalities import Modality
from ..pagination import ListMetadata


class Endpoint(AutomationConfig):
    object: Literal['deployment.endpoint'] = "deployment.endpoint"
    id: str = Field(default_factory=lambda: "endp_" + nanoid.generate(), description="Unique identifier for the extraction endpoint")

    # Extraction Endpoint Specific Config
    name: str = Field(..., description="Name of the extraction endpoint")

    # Automation Config
    webhook_url: HttpUrl = Field(..., description="Url of the webhook to send the data to")
    webhook_headers: dict[str, str] = Field(default_factory=dict, description="Headers to send with the request")
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


class ListEndpoints(BaseModel):
    data: list[Endpoint]
    list_metadata: ListMetadata


# Inherits from the methods of UpdateAutomationRequest
class UpdateEndpointRequest(UpdateAutomationRequest):
    # ------------------------------
    # Endpoint Config
    # ------------------------------
    name: Optional[str] = None

    # ------------------------------
    # HTTP Config
    # ------------------------------
    webhook_url: Optional[HttpUrl] = None
    webhook_headers: Optional[dict[str, str]] = None

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
    json_schema: Optional[dict[str, Any]] = None
    reasoning_effort: Optional[ChatCompletionReasoningEffort] = None

    @field_serializer('webhook_url')
    def url2str(self, val: HttpUrl | None) -> str | None:
        if isinstance(val, HttpUrl):
            return str(val)
        return val
