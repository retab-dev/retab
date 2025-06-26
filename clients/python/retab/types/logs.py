import datetime
import json
from typing import Any, Dict, List, Literal, Optional

import nanoid  # type: ignore
from openai.types.chat.chat_completion import ChatCompletion
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, EmailStr, Field, HttpUrl, computed_field, field_validator

from ..utils.json_schema import compute_schema_data_id
from ..utils.mime import generate_blake2b_hash_from_string
from ..utils.usage.usage import CostBreakdown, compute_cost_from_model, compute_cost_from_model_with_breakdown
from .ai_models import Amount
from .documents.extractions import RetabParsedChatCompletion
from .mime import BaseMIMEData
from .modalities import Modality
from .pagination import ListMetadata
from .browser_canvas import BrowserCanvas


class ProcessorConfig(BaseModel):
    object: str = Field(default="processor", description="Type of the object")
    id: str = Field(default_factory=lambda: "proc_" + nanoid.generate(), description="Unique identifier for the processor")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    name: str = Field(..., description="Name of the processor")

    modality: Modality
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: BrowserCanvas = Field(
        default="A4", description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type."
    )

    # New attributes
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


class AutomationConfig(BaseModel):
    @computed_field
    @property
    def object(self) -> str:
        return "automation"

    id: str = Field(default_factory=lambda: "auto_" + nanoid.generate(), description="Unique identifier for the automation")
    name: str = Field(..., description="Name of the automation")
    processor_id: str = Field(..., description="ID of the processor to use for the automation")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc), description="Timestamp of the last update")

    default_language: str = Field(default="en", description="Default language for the automation")

    # HTTP Config
    webhook_url: str = Field(..., description="Url of the webhook to send the data to")
    webhook_headers: Dict[str, str] = Field(default_factory=dict, description="Headers to send with the request")

    need_validation: bool = Field(default=False, description="If the automation needs to be validated before running")

    @field_validator("webhook_url", mode="after")
    def validate_httpurl(cls, val: Any) -> Any:
        if isinstance(val, str):
            HttpUrl(val)
        return val


class UpdateProcessorRequest(BaseModel):
    # ------------------------------
    # Processor Parameters
    # ------------------------------
    name: Optional[str] = None
    modality: Optional[Modality] = None
    image_resolution_dpi: Optional[int] = None
    browser_canvas: Optional[BrowserCanvas] = None
    model: Optional[str] = None
    json_schema: Optional[Dict] = None
    temperature: Optional[float] = None
    reasoning_effort: Optional[ChatCompletionReasoningEffort] = None
    n_consensus: Optional[int] = None

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


class UpdateAutomationRequest(BaseModel):
    name: Optional[str] = None
    # processor_id: Optional[str] = None  # TODO: Is it allowed to change the processor_id?

    default_language: Optional[str] = None

    webhook_url: Optional[str] = None
    webhook_headers: Optional[dict[str, str]] = None

    need_validation: Optional[bool] = None

    @field_validator("webhook_url", mode="after")
    def validate_httpurl(cls, val: Any) -> Any:
        if isinstance(val, str):
            HttpUrl(val)
        return val


class OpenAIRequestConfig(BaseModel):
    object: Literal["openai_request"] = "openai_request"
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
    webhook_url: Optional[str]
    request_body: dict[str, Any]
    request_headers: dict[str, str]
    request_at: datetime.datetime

    response_body: dict[str, Any]
    response_headers: dict[str, str]
    response_at: datetime.datetime

    status_code: int
    error: Optional[str] = None
    duration_ms: float

    @field_validator("webhook_url", mode="after")
    def validate_httpurl(cls, val: Any) -> Any:
        if isinstance(val, str):
            HttpUrl(val)
        return val


class LogCompletionRequest(BaseModel):
    json_schema: dict[str, Any]
    completion: ChatCompletion


class AutomationLog(BaseModel):
    object: Literal["automation_log"] = "automation_log"
    id: str = Field(default_factory=lambda: "log_auto_" + nanoid.generate(), description="Unique identifier for the automation log")
    user_email: Optional[EmailStr]  # When the user is logged or when he forwards an email
    organization_id: str
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    automation_snapshot: AutomationConfig
    completion: RetabParsedChatCompletion | ChatCompletion
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
    data: List[AutomationLog]
    list_metadata: ListMetadata
