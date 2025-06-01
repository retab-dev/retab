import copy
import datetime
import json
import os
import re
from typing import Any, ClassVar, Dict, List, Literal, Optional

import nanoid  # type: ignore
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, EmailStr, Field, HttpUrl, computed_field, field_serializer, field_validator, model_validator
from pydantic_core import Url

from ..._utils.json_schema import clean_schema, convert_schema_to_layout
from ..._utils.mime import generate_blake2b_hash_from_string
from ..logs import AutomationConfig, UpdateAutomationRequest
from ..modalities import Modality
from ..pagination import ListMetadata
from ..schemas.layout import Layout

domain_pattern = re.compile(r"^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$")


class AutomationLevel(BaseModel):
    distance_threshold: float = Field(default=0.9, description="Distance threshold for the automation")
    score_threshold: float = Field(default=0.9, description="Score threshold for the automation")


class MatchParams(BaseModel):
    endpoint: str = Field(..., description="Endpoint for matching parameters")
    headers: Dict[str, str] = Field(..., description="Headers for the request")
    path: str = Field(..., description="Path for matching parameters")


class FetchParams(BaseModel):
    endpoint: str = Field(..., description="Endpoint for fetching parameters")
    headers: Dict[str, str] = Field(..., description="Headers for the request")
    name: str = Field(..., description="Name of the fetch parameter")


class Outlook(AutomationConfig):
    object: Literal['deployment.outlook'] = "deployment.outlook"
    name: str = Field(..., description="Name of the outlook plugin")
    id: str = Field(default_factory=lambda: "outlook_" + nanoid.generate(), description="Unique identifier for the outlook")

    authorized_domains: list[str] = Field(default_factory=list, description="List of authorized domains to receive the emails from")
    authorized_emails: List[EmailStr] = Field(default_factory=list, description="List of emails to access the link")

    # Automation Config
    webhook_url: HttpUrl = Field(..., description="Url of the webhook to send the data to")
    webhook_headers: Dict[str, str] = Field(default_factory=dict, description="Headers to send with the request")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))

    # DocumentExtraction Config
    modality: Modality
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: Literal['A3', 'A4', 'A5'] = Field(default='A4', description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type.")
    model: str = Field(..., description="Model used for chat completion")
    reasoning_effort: ChatCompletionReasoningEffort = Field(
        default="medium", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )
    n_consensus: int = Field(default=1, description="Number of consensus required to validate the data")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    layout_schema: Optional[dict[str, Any]] = Field(default=None, description="Layout schema format used to display the data")

    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])

    # Optional Fields for data integration
    match_params: List[MatchParams] = Field(default_factory=list, description="List of match parameters for the outlook automation")
    fetch_params: List[FetchParams] = Field(default_factory=list, description="List of fetch parameters for the outlook automation")

    @model_validator(mode='before')
    @classmethod
    def compute_layout_schema(cls, values: dict[str, Any]) -> dict[str, Any]:
        if values.get('layout_schema') is None:
            values['layout_schema'] = convert_schema_to_layout(values['json_schema'])
        return values

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


class ListOutlooks(BaseModel):
    data: list[Outlook]
    list_metadata: ListMetadata


# Inherits from the methods of UpdateAutomationRequest
class UpdateOutlookRequest(UpdateAutomationRequest):
    name: Optional[str] = None

    authorized_domains: Optional[list[str]] = None
    authorized_emails: Optional[List[EmailStr]] = None

    match_params: Optional[List[MatchParams]] = None
    fetch_params: Optional[List[FetchParams]] = None

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
    reasoning_effort: Optional[ChatCompletionReasoningEffort] = None

    json_schema: Optional[Dict] = None
    layout_schema: Optional[dict[str, Any]] = None

    @field_validator("authorized_emails", mode="before")
    def normalize_authorized_emails(cls, emails: Optional[List[str]]) -> Optional[List[str]]:
        return [email.strip().lower() for email in emails] if emails else None

    @field_serializer('webhook_url')
    def url2str(self, val: HttpUrl | None) -> str | None:
        if isinstance(val, HttpUrl):
            return str(val)
        return val
