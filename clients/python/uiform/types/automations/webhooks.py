from pydantic import BaseModel
from typing import Optional, Any
from ..mime import MIMEData, BaseMIMEData
from uiform.types.documents.extractions import UiParsedChatCompletion
from pydantic import EmailStr


class WebhookRequest(BaseModel):
    completion: UiParsedChatCompletion
    user: Optional[EmailStr] = None
    file_payload: MIMEData
    metadata: dict[str, Any]


class BaseWebhookRequest(BaseModel):
    completion: UiParsedChatCompletion
    user: Optional[EmailStr] = None
    file_payload: BaseMIMEData
    metadata: dict[str, Any]
