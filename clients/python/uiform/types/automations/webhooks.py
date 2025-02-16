from pydantic import BaseModel
from typing import Optional, Any
from openai.types.chat.parsed_chat_completion import ParsedChatCompletion
from ..mime import MIMEData, BaseMIMEData
from pydantic import EmailStr

class WebhookRequest(BaseModel):
    completion: ParsedChatCompletion
    user: Optional[EmailStr] = None
    file_payload: MIMEData
    metadata: dict[str, Any]


class BaseWebhookRequest(BaseModel):
    completion: ParsedChatCompletion
    user: Optional[EmailStr] = None
    file_payload: BaseMIMEData
    metadata: dict[str, Any]
