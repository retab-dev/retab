from pydantic import BaseModel, Field
from typing import Literal, Optional, Dict

from typing import Any, Optional, Literal, List, Dict

from ..._resource import SyncAPIResource, AsyncAPIResource

from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.modalities import Modality

from typing import Any, Optional, Literal, List, Dict
import uuid
import datetime
from pydantic import HttpUrl, EmailStr

from ..._resource import SyncAPIResource, AsyncAPIResource
from ...types.modalities import Modality

from ...types.documents.create_messages import ChatCompletionUiformMessage 
from ...types.documents.image_settings import ImageSettings
from ...types.documents.parse import DocumentExtractRequest, DocumentExtractResponse, DocumentExtractionConfig
from ...types.ai_model import LLMModel


from ...types.mime import MIMEData, BaseMIMEData

# Never used anywhere in the logs, but will be useful
class HttpOutput(BaseModel): 
    extraction: dict[str,Any]
    payload: Optional[MIMEData] # MIMEData forwarded to the mailbox, or to the link
    user_email: Optional[EmailStr] # When the user is logged or when he forwards an email

class AutomationConfig(DocumentExtractionConfig):
    object: str
    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))

    # HTTP Config
    webhook_url: HttpUrl = Field(..., description = "Url of the webhook to send the data to")
    webhook_headers: Dict[str, str] = Field(default_factory=dict, description = "Headers to send with the request")

    def model_dump(
        self,
        **kwargs: Any,
    ) -> dict[str, Any]:
        data = super().model_dump(**kwargs)
        parent_fields = set(DocumentExtractionConfig.model_fields.keys())
        child_keys = [k for k in data if k not in parent_fields]
        parent_keys = [k for k in data if k in parent_fields]
        return {k: data[k] for k in (child_keys + parent_keys)}
    
    def __str__(self) -> str:
        data = self.model_dump()
        items = [f"{k}={repr(v)}" for k, v in data.items()]
        return f"{self.__class__.__name__}({', '.join(items)})"
    
    def __repr__(self) -> str:
        return self.__str__()
    
from typing import ClassVar
import os
class MailboxConfig(AutomationConfig):
    EMAIL_PATTERN: ClassVar[str] = f".*@{os.getenv('EMAIL_DOMAIN', 'devmail.uiform.com')}$"
    object: Literal['mailbox'] = "mailbox"
    id: str = Field(default_factory=lambda: "mb_" + str(uuid.uuid4()), description="Unique identifier for the mailbox")
    
    # Email Specific config
    email: str = Field(..., pattern=EMAIL_PATTERN)
    authorized_domains: list[str] = Field(default_factory=list, description = "List of authorized domains to receive the emails from")
    authorized_emails: List[EmailStr] = Field(default_factory=list, description = "List of emails to access the link")

    def model_dump(
        self,
        **kwargs: Any,
    ) -> dict[str, Any]:
        data = super().model_dump(**kwargs)
        parent_fields = set(DocumentExtractionConfig.model_fields.keys())
        child_keys = [k for k in data if k not in parent_fields]
        parent_keys = [k for k in data if k in parent_fields]
        return {k: data[k] for k in (child_keys + parent_keys)}
    
    def __str__(self) -> str:
        data = self.model_dump()
        items = [f"{k}={repr(v)}" for k, v in data.items()]
        return f"{self.__class__.__name__}({', '.join(items)})"
    
    def __repr__(self) -> str:
        return self.__str__()
    
    
    


LinkProtection = Literal["unprotected", "invite-only"]

class ExtractionLinkConfig(AutomationConfig):
    object: Literal['extraction_link'] = "extraction_link"
    id: str = Field(default_factory=lambda: "el_" + str(uuid.uuid4()), description="Unique identifier for the extraction link")
    
    # Link Specific Config
    name: str = Field(..., description = "Name of the link")
    password: Optional[str] = Field(None, description = "Password to access the link")

    def model_dump(
        self,
        **kwargs: Any,
    ) -> dict[str, Any]:
        data = super().model_dump(**kwargs)
        parent_fields = set(DocumentExtractionConfig.model_fields.keys())
        child_keys = [k for k in data if k not in parent_fields]
        parent_keys = [k for k in data if k in parent_fields]
        return {k: data[k] for k in (child_keys + parent_keys)}

    def __str__(self) -> str:
        data = self.model_dump()
        items = [f"{k}={repr(v)}" for k, v in data.items()]
        return f"{self.__class__.__name__}({', '.join(items)})"
    
    def __repr__(self) -> str:
        return self.__str__()
    
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


class OutlookConfig(AutomationConfig):
    object: Literal['outlook_plugin'] = "outlook_plugin"
    id: str = Field(default_factory=lambda: "outlook_" + str(uuid.uuid4()), description="Unique identifier for the outlook account")
    
    authorized_domains: list[str] = Field(default_factory=list, description = "List of authorized domains to connect to the plugin from")
    authorized_emails: List[EmailStr] = Field(default_factory=list, description = "List of emails to connect to the plugin from")

class ExtractionEndpointConfig(AutomationConfig):
    object: Literal['extraction_endpoint'] = "extraction_endpoint"
    id: str = Field(default_factory=lambda: "extraction_endpoint_" + str(uuid.uuid4()), description="Unique identifier for the extraction endpoint")
    
    # Extraction Endpoint Specific Config
    name: str = Field(..., description="Name of the extraction endpoint")


# ------------------------------
# ------------------------------
# ------------------------------


class ExternalRequestLog(BaseModel):
    webhook_url: Optional[HttpUrl]
    request_body: dict[str, Any]
    request_headers: dict[str, str]
    request_timestamp: datetime.datetime

    response_body: dict[str, Any]
    response_headers: dict[str, str]
    response_timestamp: datetime.datetime

    status_code: int
    error: Optional[str] = None
    duration_ms: float

class InternalLog(BaseModel):
    automation_snapshot:  Optional[MailboxConfig| ExtractionLinkConfig| ScrappingConfig| ExtractionEndpointConfig]
    file_metadata: BaseMIMEData
    extraction: Optional[DocumentExtractResponse]
    received_timestamp: Optional[datetime.datetime]

class AutomationLog(BaseModel):
    object: Literal['automation_log'] = "automation_log"
    id: str = Field(default_factory=lambda: "log_auto_" + str(uuid.uuid4()), description="Unique identifier for the automation log")
    user_email: Optional[EmailStr] # When the user is logged or when he forwards an email
    organization_id:str
    internal_log: InternalLog
    external_request_log: ExternalRequestLog



# ------------------------------
# ------------------------------
# ------------------------------

class UpdateMailBoxRequest(BaseModel):

    authorized_domains: Optional[list[str]] = None
    authorized_emails: Optional[List[EmailStr]] = None

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
    model: Optional[LLMModel] = None
    temperature: Optional[float] = None
    json_schema: Optional[Dict] = None

   

class UpdateExtractionLinkRequest(BaseModel):
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
    model: Optional[LLMModel] = None
    temperature: Optional[float] = None
    json_schema: Optional[Dict] = None

