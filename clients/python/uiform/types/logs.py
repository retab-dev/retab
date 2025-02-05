from pydantic import BaseModel, Field,  HttpUrl, EmailStr, computed_field
from typing import Any, Optional, Literal, List, Dict
import uuid
import datetime

from .documents.extractions import DocumentExtractResponse
from .image_settings import ImageSettings
from .pagination import ListMetadata
from .modalities import Modality
from .mime import BaseMIMEData, MIMEData

from .usage import Amount
from .._utils.usage.usage import compute_cost_from_model




class HttpOutput(BaseModel):
    completion: DocumentExtractResponse
    user: Optional[EmailStr] = None
    file_payload: MIMEData


class AutomationConfig(BaseModel):
    object: str
    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))

    # HTTP Config
    webhook_url: HttpUrl = Field(..., description = "Url of the webhook to send the data to")
    webhook_headers: Dict[str, str] = Field(default_factory=dict, description = "Headers to send with the request")

    modality: Modality
    image_settings : ImageSettings = Field(default_factory=ImageSettings, description="Preprocessing operations applied to image before sending them to the llm")

    # New attributes
    model: str = Field(..., description="Model used for chat completion")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])

    
class CronSchedule(BaseModel):
    second: Optional[int] = Field(0, ge=0, le=59, description="Second (0-59), defaults to 0")
    minute: int = Field(..., ge=0, le=59, description="Minute (0-59)")
    hour: int = Field(..., ge=0, le=23, description="Hour (0-23)")
    day_of_month: Optional[int] = Field(None, ge=1, le=31, description="Day of the month (1-31), None means any day")
    month: Optional[int] = Field(None, ge=1, le=12, description="Month (1-12), None means every month")
    day_of_week: Optional[int] = Field(None, ge=0, le=6, description="Day of the week (0-6, Sunday = 0), None means any day")

    def to_cron_string(self) -> str:
        return f"{self.second or '*'} {self.minute} {self.hour} " \
               f"{self.day_of_month or '*'} {self.month or '*'} {self.day_of_week or '*'}"


class ScrappingConfig(AutomationConfig):
    object: Literal['scrapping_cron'] = "scrapping_cron"
    id: str = Field(default_factory=lambda: "scrapping_" + str(uuid.uuid4()), description="Unique identifier for the scrapping job")
    
    # Scrapping Specific Config
    link: HttpUrl = Field(..., description="Link to be scrapped")
    schedule: CronSchedule


class OutlookConfig(BaseModel):
    object: Literal['outlook_plugin'] = "outlook_plugin"
    id: str = Field(default_factory=lambda: "outlook_" + str(uuid.uuid4()), description="Unique identifier for the outlook account")
    
    authorized_domains: list[str] = Field(default_factory=list, description = "List of authorized domains to connect to the plugin from")
    authorized_emails: List[EmailStr] = Field(default_factory=list, description = "List of emails to connect to the plugin from")

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


class ExtractionEndpointConfig(BaseModel):
    object: Literal['endpoint'] = "endpoint"
    id: str = Field(default_factory=lambda: "endp" + str(uuid.uuid4()), description="Unique identifier for the extraction endpoint")
    
    # Extraction Endpoint Specific Config
    name: str = Field(..., description="Name of the extraction endpoint")

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


class OpenAIRequestConfig(BaseModel):
    object: Literal['openai_request'] = "openai_request"
    id: str = Field(default_factory=lambda: "openai_req_" + str(uuid.uuid4()), description="Unique identifier for the openai request")
    model: str
    json_schema: dict[str, Any]

# ------------------------------
# ------------------------------
# ------------------------------

from .automations.mailboxes import Mailbox
from .automations.links import Link

#class OpenAILog(BaseModel): 
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


from openai.types.chat import completion_create_params
from openai.types.chat.chat_completion import ChatCompletion

class LogCompletionRequest(BaseModel):
    json_schema: dict[str, Any]
    completion: ChatCompletion


class AutomationLog(BaseModel):
    object: Literal['automation_log'] = "automation_log"
    id: str = Field(default_factory=lambda: "log_auto_" + str(uuid.uuid4()), description="Unique identifier for the automation log")
    user_email: Optional[EmailStr] # When the user is logged or when he forwards an email
    organization_id:str
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    automation_snapshot:  Mailbox| Link| ScrappingConfig| ExtractionEndpointConfig | OpenAIRequestConfig
    completion: DocumentExtractResponse | ChatCompletion
    file_metadata: Optional[BaseMIMEData]
    external_request_log: Optional[ExternalRequestLog]

    @computed_field # type: ignore
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

class ListLogs(BaseModel):
    data: List[AutomationLog]
    list_metadata: ListMetadata

