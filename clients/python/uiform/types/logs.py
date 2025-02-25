from pydantic import BaseModel, Field,  HttpUrl, EmailStr, computed_field, field_serializer
from typing import Any, Optional, Literal, List, Dict
import datetime
import nanoid # type: ignore
from pydantic_core import Url

from .documents.extractions import DocumentExtractResponse
from .image_settings import ImageSettings
from .pagination import ListMetadata
from .modalities import Modality
from .mime import BaseMIMEData

from .usage import Amount
from .._utils.usage.usage import compute_cost_from_model


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

    
    @field_serializer('webhook_url')
    def url2str(self, val: HttpUrl) -> str:
        return str(val)




class OpenAIRequestConfig(BaseModel):
    object: Literal['openai_request'] = "openai_request"
    id: str = Field(default_factory=lambda: "openai_req_" + nanoid.generate(), description="Unique identifier for the openai request")
    model: str
    json_schema: dict[str, Any]

# ------------------------------
# ------------------------------
# ------------------------------

#from .automations.mailboxes import Mailbox
#from .automations.links import Link
#from .automations.cron import ScrappingConfig
#from .automations.endpoint import Endpoint
#from .automations.outlook import Outlook

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


class AutomationLog(BaseModel):
    object: Literal['automation_log'] = "automation_log"
    id: str = Field(default_factory=lambda: "log_auto_" + nanoid.generate(), description="Unique identifier for the automation log")
    user_email: Optional[EmailStr] # When the user is logged or when he forwards an email
    organization_id:str
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    automation_snapshot:  AutomationConfig
    completion: DocumentExtractResponse | ChatCompletion
    file_metadata: Optional[BaseMIMEData]
    external_request_log: Optional[ExternalRequestLog]
    extraction_id: Optional[str]=Field(default=None, description="ID of the extraction")
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

