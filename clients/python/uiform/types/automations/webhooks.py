from pydantic import BaseModel
from typing import Optional, Any
from ..mime import MIMEData, BaseMIMEData
from uiform.types.documents.extractions import UiParsedChatCompletion
from pydantic import EmailStr


class WebhookRequest(BaseModel):
    completion: UiParsedChatCompletion
    user: Optional[EmailStr] = None
    file_payload: MIMEData
    metadata: Optional[dict[str, Any]] = None


class BaseWebhookRequest(BaseModel):
    completion: UiParsedChatCompletion
    user: Optional[EmailStr] = None
    file_payload: BaseMIMEData
    metadata: Optional[dict[str, Any]] = None
