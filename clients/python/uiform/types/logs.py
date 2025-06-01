import copy
import datetime
import json
from typing import Any, Dict, List, Literal, Optional

import nanoid  # type: ignore
from openai import OpenAI
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, EmailStr, Field, HttpUrl, computed_field, field_serializer
from pydantic_core import Url

from .._utils.json_schema import clean_schema, compute_schema_data_id
from .._utils.mime import generate_blake2b_hash_from_string
from .._utils.usage.usage import compute_cost_from_model, compute_cost_from_model_with_breakdown, CostBreakdown
from .ai_models import Amount
from .documents.extractions import UiParsedChatCompletion
from .mime import BaseMIMEData
from .modalities import Modality
from .pagination import ListMetadata


class AutomationConfig(BaseModel):
    object: str
    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    default_language: str = Field(default="en", description="Default language for the automation")

    # HTTP Config
    webhook_url: HttpUrl = Field(..., description="Url of the webhook to send the data to")
    webhook_headers: Dict[str, str] = Field(default_factory=dict, description="Headers to send with the request")

    modality: Modality
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: Literal['A3', 'A4', 'A5'] = Field(default='A4', description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type.")

    # New attributes
    model: str = Field(..., description="Model used for chat completion")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])
    reasoning_effort: ChatCompletionReasoningEffort = Field(
        default="medium", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )
    need_validation: bool = Field(default=False, description="If the automation needs to be validated before running")
    n_consensus: int = Field(default=1, description="Number of consensus required to validate the data")

    @computed_field  # type: ignore
    @property
    def schema_data_id(self) -> str:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.

        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        return compute_schema_data_id(self.json_schema)

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


class UpdateAutomationRequest(BaseModel):
    # ------------------------------
    # HTTP Config
    # ------------------------------
    webhook_url: Optional[HttpUrl] = None
    webhook_headers: Optional[Dict[str, str]] = None

    # ------------------------------
    # DocumentProcessing Parameters
    # ------------------------------
    image_resolution_dpi: Optional[int] = None
    browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None
    modality: Optional[Modality] = None
    model: Optional[str] = None
    temperature: Optional[float] = None
    json_schema: Optional[Dict] = None
    reasoning_effort: Optional[ChatCompletionReasoningEffort] = None
    need_validation: Optional[bool] = None
    n_consensus: Optional[int] = None

    @field_serializer('webhook_url')
    def url2str(self, val: HttpUrl | None) -> str | None:
        if isinstance(val, HttpUrl):
            return str(val)
        return val

    @computed_field  # type: ignore
    @property
    def schema_data_id(self) -> Optional[str]:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.

        Returns:
            str: A SHA1 hash string representing the schema data version.
        """

        if self.json_schema is None:
            return None
        return compute_schema_data_id(self.json_schema)

    @computed_field  # type: ignore
    @property
    def schema_id(self) -> Optional[str]:
        """Returns the SHA1 hash of the complete schema.

        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        if self.json_schema is None:
            return None
        return "sch_id_" + generate_blake2b_hash_from_string(json.dumps(self.json_schema, sort_keys=True).strip())


class OpenAIRequestConfig(BaseModel):
    object: Literal['openai_request'] = "openai_request"
    id: str = Field(default_factory=lambda: "openai_req_" + nanoid.generate(), description="Unique identifier for the openai request")
    model: str
    json_schema: dict[str, Any]
    reasoning_effort: Optional[ChatCompletionReasoningEffort] = None


# ------------------------------
# ------------------------------
# ------------------------------

# from .automations.mailboxes import Mailbox
# from .automations.links import Link
# from .automations.cron import ScrappingConfig
# from .automations.outlook import Outlook

# class OpenAILog(BaseModel):
#    request_config: OpenAIRequestConfig
#    completion: ChatCompletion


class ExternalRequestLog(BaseModel):
    webhook_url: Optional[HttpUrl]
    request_body: dict[str, Any]
    request_headers: dict[str, str]
    request_at: datetime.datetime

    response_body: dict[str, Any]
    response_headers: dict[str, str]
    response_at: datetime.datetime

    status_code: int
    error: Optional[str] = None
    duration_ms: float

    @field_serializer('webhook_url')
    def url2str(self, val: HttpUrl | None) -> str | None:
        if isinstance(val, HttpUrl):
            return str(val)
        return val


from openai.types.chat import completion_create_params
from openai.types.chat.chat_completion import ChatCompletion


class LogCompletionRequest(BaseModel):
    json_schema: dict[str, Any]
    completion: ChatCompletion


class DeploymentLog(BaseModel):
    object: Literal['automation_log'] = "automation_log"
    id: str = Field(default_factory=lambda: "log_auto_" + nanoid.generate(), description="Unique identifier for the automation log")
    user_email: Optional[EmailStr]  # When the user is logged or when he forwards an email
    organization_id: str
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    deployment_snapshot: AutomationConfig
    completion: UiParsedChatCompletion | ChatCompletion
    file_metadata: Optional[BaseMIMEData]
    external_request_log: Optional[ExternalRequestLog]
    extraction_id: Optional[str] = Field(default=None, description="ID of the extraction")

    @computed_field  # type: ignore
    @property
    def api_cost(self) -> Optional[Amount]:
        if self.completion and self.completion.usage:
            try:
                cost = compute_cost_from_model(self.completion.model, self.completion.usage)
                return cost
            except Exception as e:
                print(f"Error computing cost: {e}")
                return None
        return None
    
    @computed_field  # type: ignore
    @property
    def cost_breakdown(self) -> Optional[CostBreakdown]:
        if self.completion and self.completion.usage:
            try:
                cost = compute_cost_from_model_with_breakdown(self.completion.model, self.completion.usage)
                return cost
            except Exception as e:
                print(f"Error computing cost: {e}")
                return None
        return None


class ListLogs(BaseModel):
    data: List[DeploymentLog]
    list_metadata: ListMetadata
