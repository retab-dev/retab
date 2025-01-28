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
from ...types.documents.image_operations import ImageOperations
from ...types.documents.parse import DocumentExtractRequest, DocumentExtractResponse, DocumentExtractionConfig
from ...types.documents.text_operations import TextOperations
from ...types.ai_model import LLMModel

import secrets
import string

from ...types.mime import MIMEData, BaseMIMEData

# Never used anywhere in the logs, but will be useful
class HttpOutput(BaseModel): 
    extraction: dict[str,Any]
    payload: Optional[MIMEData] # Only if forward_file is true -> MIMEData forwarded to the mailbox, or to the link
    user_email: Optional[EmailStr] # When the user is logged or when he forwards an email


class HttpConfig(BaseModel):
    endpoint: HttpUrl = Field(..., description = "Endpoint to send the data to")
    method: Literal["POST"]= "POST"
    headers: Dict[str, str] = Field(default_factory=dict, description = "Headers to send with the request")
    outgoing_ip: Literal["34.163.38.96"] = Field("34.163.38.96", description = "IP address of the server that will send the data to the endpoint")
    max_file_size: int = Field(default=50, description = "Maximum file size in MB")
    forward_file: bool = Field(default=False, description = "Whether to forward the file to the endpoint")

class AutomationConfig(DocumentExtractionConfig):
    object: str
    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    http_config: HttpConfig



class MailboxConfig(AutomationConfig):
    object: Literal['mailbox'] = "mailbox"
    id: str = Field(default_factory=lambda: "mb_" + str(uuid.uuid4()), description="Unique identifier for the mailbox")
    
    # Email Specific config
    email: str = Field(..., pattern=r".*@mailbox\.uiform\.com$")
    follow_up: bool = Field(default=False, description = "Whether to send a follow-up email to the user to confirm the success of the email forwarding")
    authorized_domains: list[str] = Field(default_factory=list, pattern=r"^[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$", description = "List of authorized domains to receive the emails from")
    authorized_emails: List[EmailStr] = Field(default_factory=list, description = "List of emails to access the link")
    


class LinkProtection(BaseModel): 
    protection_type: Literal["none", "password", "invitations"] = "none"
    password: str | None = Field(default=None, description = "Password to access the link")
    invitations: List[EmailStr] = Field(default_factory=list, description = "List of emails allowed to access the link")

    def __init__(self, **data: Any)->None:
        super().__init__(**data)
        if self.protection_type == "password" and self.password is None:
            # Generate a random 12 character password with letters and numbers
            self.password = 'pwd_' + ''.join(secrets.choice(string.ascii_letters + string.digits) for _ in range(12))


class ExtractionLinkConfig(AutomationConfig):
    object: Literal['extraction_link'] = "extraction_link"
    id: str = Field(default_factory=lambda: "el_" + str(uuid.uuid4()), description="Unique identifier for the extraction link")
    
    # Link Specific Config
    name: str = Field(..., description = "Name of the link")
    protection: LinkProtection



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



class ExtractionEndpointConfig(AutomationConfig):
    object: Literal['extraction_endpoint'] = "extraction_endpoint"
    id: str = Field(default_factory=lambda: "extraction_endpoint_" + str(uuid.uuid4()), description="Unique identifier for the extraction endpoint")
    
    # Extraction Endpoint Specific Config
    name: str = Field(..., description="Name of the extraction endpoint")


# ------------------------------
# ------------------------------
# ------------------------------



class RequestLog(BaseModel):
    endpoint: HttpUrl
    request_body: dict[str, Any]
    request_headers: dict[str, str]
    request_timestamp: datetime.datetime 
    
    response_body: dict[str, Any]
    response_headers: dict[str, str]
    response_timestamp: datetime.datetime 
    
    status_code: int
    error: Optional[str] = None
    duration_ms: float 


class AutomationLog(BaseModel):
    id: str = Field(default_factory=lambda: "log_auto_" + str(uuid.uuid4()), description="Unique identifier for the automation log")
    automation_snapshot: AutomationConfig # i.e MailboxConfig, ExtractionLinkConfig, ScrappingConfig, ExtractionEndpointConfig

    file_metadata: BaseMIMEData
    user_email: Optional[EmailStr] # When the user is logged or when he forwards an email

    request_log: RequestLog

# ------------------------------
# ------------------------------
# ------------------------------

class UpdateMailBoxRequest(BaseModel):
    id: str

    follow_up: Optional[bool] = None
    authorized_domains: Optional[list[str]] = None
    authorized_emails: Optional[List[EmailStr]] = None
    http_config: Optional[HttpConfig] = None

    # ------------------------------
    # DocumentExtraction Parameters
    # ------------------------------
    # DocumentProcessing Parameters
    text_operations: Optional[TextOperations] = None
    image_operations: Optional[ImageOperations] = None
    modality: Optional[Literal["native"]] = None
    # Others DocumentExtraction Parameters
    model: Optional[LLMModel] = None
    temperature: Optional[float] = None
    additional_messages: Optional[list[ChatCompletionUiformMessage]] = None
    json_schema: Optional[Dict] = None

   
    
class UpdateExtractionLinkRequest(BaseModel):
    id: str

    name: Optional[str] = None
    protection: Optional[LinkProtection] = None
    http_config: Optional[HttpConfig] = None
    
    # ------------------------------
    # DocumentExtraction Parameters
    # ------------------------------
    # DocumentProcessing Parameters
    text_operations: Optional[TextOperations] = None
    image_operations: Optional[ImageOperations] = None
    modality: Optional[Literal["native"]] = None
    # Others DocumentExtraction Parameters
    model: Optional[LLMModel] = None
    temperature: Optional[float] = None
    additional_messages: Optional[list[ChatCompletionUiformMessage]] = None
    json_schema: Optional[Dict] = None

