from typing import Any, Optional

from pydantic import BaseModel, EmailStr

from retab.types.documents.extractions import RetabParsedChatCompletion

from ..mime import BaseMIMEData, MIMEData


class WebhookRequest(BaseModel):
    completion: RetabParsedChatCompletion
    user: Optional[EmailStr] = None
    file_payload: MIMEData
    metadata: Optional[dict[str, Any]] = None


class BaseWebhookRequest(BaseModel):
    completion: RetabParsedChatCompletion
    user: Optional[EmailStr] = None
    file_payload: BaseMIMEData
    metadata: Optional[dict[str, Any]] = None
