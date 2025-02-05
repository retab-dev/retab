from pydantic import BaseModel
from typing import Optional
from ..documents.extractions import DocumentExtractResponse
from ..mime import MIMEData
from pydantic import EmailStr

class WebhookRequest(BaseModel):
    completion: DocumentExtractResponse
    user: Optional[EmailStr] = None
    file_payload: MIMEData
