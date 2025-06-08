from typing import Any, Optional

from pydantic import BaseModel, EmailStr

from uiform.types.documents.extractions import UiParsedChatCompletion

from ..mime import BaseMIMEData, MIMEData


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
