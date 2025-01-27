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
from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.modalities import Modality

from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.documents.image_operations import ImageOperations
from ...types.documents.parse import DocumentExtractRequest, DocumentExtractResponse
from ...types.documents.text_operations import TextOperations
from ...types.mime import MIMEData


import secrets
import string


class DocumentProcessingConfig(BaseModel):
    text_operations: TextOperations = Field(default=TextOperations(), description="Additional context to be used by the AI model")
    image_operations: ImageOperations = Field(default=ImageOperations(**{
        "correct_image_orientation": True,
        "dpi": 72,
        "image_to_text": "ocr",
        "browser_canvas": "A4"
    }), description="Additional context to be used by the AI model")
    modality: Literal["native"] = "native"

    
    

class AutomationConfig(DocumentProcessingConfig):
    id: str
    organization_id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    json_schema: Dict = Field(..., description="JSON schema to validate the email data")
    additional_messages: list[ChatCompletionUiformMessage] = []
    model: str = "gpt-4o-mini"
    temperature: float = 0


class WebhookOutput(BaseModel):
    # The output of the webhook that the person will receive
    mime_data: MIMEData # or some tidy version of it
    extraction: DocumentExtractResponse

class WebhookConfig(BaseModel):
    endpoint: HttpUrl = Field(..., description = "Endpoint to send the data to")
    method: Literal["POST"]= "POST"
    headers: Dict[str, str] = Field(default_factory=dict, description = "Headers to send with the request")
    outgoing_ip: Literal["34.163.38.96"] = Field("34.163.38.96", description = "IP address of the server that will send the data to the endpoint")


class MailboxConfig(AutomationConfig):
    object: Literal['mailbox'] = "mailbox"
    id: str = "mb_" + str(uuid.uuid4())
    email: str = Field(..., pattern=r".*@mailbox\.uiform\.com$")
    follow_up: bool = Field(default=False, description = "Whether to send a follow-up email to the user to confirm the success of the email forwarding")
    webhook_config: WebhookConfig

class LinkProtection(BaseModel): 
    is_secured: bool = Field(default=False, description = "Whether the link is secured")
    password: str | None = Field(default=None, description = "Password to access the link")

    def __init__(self, **data: Any)->None:
        super().__init__(**data)
        if self.is_secured and self.password is None:
            # Generate a random 12 character password with letters and numbers
            self.password = 'pwd_' + ''.join(secrets.choice(string.ascii_letters + string.digits) for _ in range(12))

class ExtractionLinkConfig(AutomationConfig):
    object: Literal['extraction_link'] = "extraction_link"
    max_file_size: int = Field(default=10, description = "Maximum file size in MB")
    protection: LinkProtection

class ReconciliationLinkConfig(AutomationConfig):
    object: Literal['reconciliation_link'] = "reconciliation_link"
    max_file_size: int = Field(default=10, description = "Maximum file size in MB")
    protection: LinkProtection
    webhook_config: WebhookConfig



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
    object: Literal['scrapping'] = "scrapping"
    id: str = Field(default_factory=lambda: "scrapping_" + str(uuid.uuid4()), description="Unique identifier for the scrapping job")
    link: HttpUrl = Field(..., description="Link to be scrapped")

# ------------------------------
# ------------------------------
# ------------------------------




class MailboxLog(BaseModel):
    # Logs that will be stored in the database and displayed in the frontend
    sender: str
    receiver: str
    subject: Optional[str]
    body: Optional[str]
    body_html: Optional[str]
    sent_at: datetime.datetime
    received_at: Optional[datetime.datetime]
    processed_at: datetime.datetime
    email: str
    organization_id: str

class ExtractionLinkLog(BaseModel):
    # Logs that will be stored in the database and displayed in the frontend
    sender: str
    file_name: str
    file_size: int  # in bytes
    mime_type: str
    uploaded_at: datetime.datetime
    processed_at: datetime.datetime
    link_id: str
    organization_id: str
    protection_used: bool  # whether the link was accessed with password protection

class CreateMailBoxRequest(BaseModel):
    email: str #= Field(..., pattern=r".*@uiform\.com$")
    webhook_config: WebhookConfig
    json_schema: Dict = Field(..., description="JSON schema to validate the email data")
    text_operations: TextOperations = Field(default=TextOperations(), description="Additional context to be used by the AI model")
    image_operations: ImageOperations = Field(default=ImageOperations(**{
        "correct_image_orientation": True,
        "dpi": 72,
        "image_to_text": "ocr",
        "browser_canvas": "A4"
    }), description="Additional context to be used by the AI model")
    modality: Literal["native"] = "native"
    model: str = "gpt-4o-mini"
    temperature: float = 0
    additional_messages: list[ChatCompletionUiformMessage] = []


class UpdateMailBoxRequest(BaseModel):
    webhook_config: Optional[WebhookConfig] = None
    text_operations: Optional[TextOperations] = None
    image_operations: Optional[ImageOperations] = None
    modality: Optional[Literal["native"]] = None
    model: Optional[str] = None
    temperature: Optional[float] = None
    additional_messages: Optional[list[ChatCompletionUiformMessage]] = None
    json_schema: Optional[Dict] = None
