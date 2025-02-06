from pydantic import BaseModel
from typing import Optional
from openai.types.chat.parsed_chat_completion import ParsedChatCompletion
from ..mime import MIMEData
from pydantic import EmailStr

class WebhookRequest(BaseModel):
    completion: ParsedChatCompletion
    user: Optional[EmailStr] = None
    file_payload: MIMEData
