from pydantic import BaseModel, Field
from typing import Literal, Optional, Dict

from typing import Any, Optional, Literal, List, Dict

from ..._resource import SyncAPIResource, AsyncAPIResource

from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.modalities import Modality

from typing import Any, Optional, Literal, List, Dict
import uuid
import datetime
from pydantic import HttpUrl

from ..._resource import SyncAPIResource, AsyncAPIResource
from ...types.modalities import Modality

from ...types.documents.create_messages import ChatCompletionUiformMessage, DocumentProcessingConfig
from ...types.documents.image_operations import ImageOperations
from ...types.documents.parse import DocumentExtractRequest, DocumentExtractResponse
from ...types.documents.text_operations import TextOperations
from ...types.mime import MIMEData, BaseMIMEData
from ...types.ai_model import LLMModel

import secrets
import string




class WebhookConfig(BaseModel):
    endpoint: HttpUrl = Field(..., description = "Endpoint to send the data to")
    method: Literal["POST"]= "POST"
    headers: Dict[str, str] = Field(default_factory=dict, description = "Headers to send with the request")
    outgoing_ip: Literal["34.163.38.96"] = Field("34.163.38.96", description = "IP address of the server that will send the data to the endpoint")



class AutomationConfig(DocumentProcessingConfig):
    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    json_schema: Dict = Field(..., description="JSON schema to validate the email data")
    additional_messages: list[ChatCompletionUiformMessage] = []
    model: LLMModel = "gpt-4o-mini"
    temperature: float = 0
    webhook_config: WebhookConfig


class WebhookOutput(BaseModel):
    # The output of the webhook that the person will receive
    mime_data: MIMEData # or some tidy version of it
    extraction: DocumentExtractResponse


class MailboxConfig(AutomationConfig):
    object: Literal['mailbox'] = "mailbox"
    id: str = Field(default_factory=lambda: "mb_" + str(uuid.uuid4()), description="Unique identifier for the mailbox")
    email: str = Field(..., pattern=r".*@mailbox\.uiform\.com$")
    follow_up: bool = Field(default=False, description = "Whether to send a follow-up email to the user to confirm the success of the email forwarding")

from pydantic import EmailStr

class LinkProtection(BaseModel): 
    protection_type: Literal["none", "password", "organization", "invitations"] = "none"
    password: str | None = Field(default=None, description = "Password to access the link")
    invitations: List[EmailStr] = Field(default_factory=list, description = "List of emails to access the link")

    def __init__(self, **data: Any)->None:
        super().__init__(**data)
        if self.protection_type == "password" and self.password is None:
            # Generate a random 12 character password with letters and numbers
            self.password = 'pwd_' + ''.join(secrets.choice(string.ascii_letters + string.digits) for _ in range(12))




class ReconciliationLinkConfig(AutomationConfig):
    object: Literal['reconciliation_link'] = "reconciliation_link"
    max_file_size: int = Field(default=10, description = "Maximum file size in MB")
    protection: LinkProtection



class ExtractionLinkConfig(AutomationConfig):
    object: Literal['extraction_link'] = "extraction_link"
    id: str = Field(default_factory=lambda: "el_" + str(uuid.uuid4()), description="Unique identifier for the extraction link")
    
    name: str = Field(..., description = "Name of the link")
    max_file_size: int = Field(default=10, description = "Maximum file size in MB")
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

class CronJobConfig(AutomationConfig):
    object: Literal['cron_job'] = "cron_job"
    id: str = Field(default_factory=lambda: "cron_" + str(uuid.uuid4()), description="Unique identifier for the cron job")
    organization_id: str = Field(..., description="Organization ID that owns the cron job")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    schedule: CronSchedule

class ScrappingConfig(CronJobConfig):
    id: str = Field(default_factory=lambda: "scrapping_" + str(uuid.uuid4()), description="Unique identifier for the scrapping job")
    link: HttpUrl = Field(..., description="Link to be scrapped")

# ------------------------------
# ------------------------------
# ------------------------------

class EmailMetadata(BaseModel):
    sender: str
    recipient: str
    subject: Optional[str]
    body: Optional[str]
    body_html: Optional[str]
    sent_at: datetime.datetime
    received_at: Optional[datetime.datetime]
    processed_at: datetime.datetime

class RequestResult(BaseModel):
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

class MailboxLog(BaseModel):
    email: str
    organization_id: str

    email_metadata: EmailMetadata
    request_result: RequestResult

class ExtractionLinkLog(BaseModel):
    link_id: str 
    organization_id: str

    name: str = Field(..., description = "Name of the link")
    max_file_size: int = Field(default=10, description = "Maximum file size in MB")
    protection: LinkProtection
    file_metadata: BaseMIMEData

    request_result: RequestResult















# ------------------------------
# ------------------------------
# ------------------------------

class UpdateMailBoxRequest(BaseModel):
    webhook_config: Optional[WebhookConfig] = None
    text_operations: Optional[TextOperations] = None
    image_operations: Optional[ImageOperations] = None
    modality: Optional[Literal["native"]] = None
    model: Optional[LLMModel] = None
    temperature: Optional[float] = None
    additional_messages: Optional[list[ChatCompletionUiformMessage]] = None
    json_schema: Optional[Dict] = None
    
class UpdateExtractionLinkRequest(BaseModel):
    id: str
    name: Optional[str] = None
    max_file_size: Optional[int] = None
    protection: Optional[LinkProtection] = None
    webhook_config: Optional[WebhookConfig] = None
    text_operations: Optional[TextOperations] = None
    image_operations: Optional[ImageOperations] = None
    modality: Optional[Modality] = None
    model: Optional[str] = None
    temperature: Optional[float] = None
    additional_messages: Optional[list[ChatCompletionUiformMessage]] = None
    json_schema: Optional[Dict] = None

