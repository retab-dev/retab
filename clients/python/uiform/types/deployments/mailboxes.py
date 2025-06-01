import copy
import datetime
import json
import os
import re
from typing import Any, ClassVar, Dict, List, Literal, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, EmailStr, Field, HttpUrl, computed_field, field_serializer, field_validator
from pydantic_core import Url

from ..._utils.json_schema import clean_schema
from ..._utils.mime import generate_blake2b_hash_from_string
from ..modalities import Modality
from ..pagination import ListMetadata

domain_pattern = re.compile(r"^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$")

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort

from ..logs import AutomationConfig, UpdateAutomationRequest


class Mailbox(AutomationConfig):
    EMAIL_PATTERN: ClassVar[str] = f".*@{os.getenv('EMAIL_DOMAIN', 'mailbox.uiform.com')}$"
    object: Literal['deployment.mailbox'] = "deployment.mailbox"
    id: str = Field(default_factory=lambda: "mb_" + nanoid.generate(), description="Unique identifier for the mailbox")

    # Email Specific config
    email: str = Field(..., pattern=EMAIL_PATTERN)
    authorized_domains: list[str] = Field(default_factory=list, description="List of authorized domains to receive the emails from")
    authorized_emails: List[EmailStr] = Field(default_factory=list, description="List of emails to access the link")

    # Automation Config
    webhook_url: HttpUrl = Field(..., description="Url of the webhook to send the data to")
    webhook_headers: Dict[str, str] = Field(default_factory=dict, description="Headers to send with the request")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))

    # DocumentExtraction Config
    modality: Modality
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: Literal['A3', 'A4', 'A5'] = Field(default='A4', description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type.")
    model: str = Field(..., description="Model used for chat completion")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])
    reasoning_effort: ChatCompletionReasoningEffort = Field(
        default="medium", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )
    n_consensus: int = Field(default=1, description="Number of consensus required to validate the data")

    # Normalize email fields (case-insensitive)
    @field_validator("email", mode="before")
    def normalize_email(cls, value: str) -> str:
        return value.strip().lower()

    @field_validator("authorized_emails", mode="before")
    def normalize_authorized_emails(cls, emails: List[str]) -> List[str]:
        return [email.strip().lower() for email in emails]

    @field_validator('authorized_domains', mode='before')
    def validate_domain(cls, list_domains: list[str]) -> list[str]:
        for domain in list_domains:
            if not domain_pattern.match(domain):
                raise ValueError(f"Invalid domain: {domain}")
        return list_domains


class ListMailboxes(BaseModel):
    data: list[Mailbox]
    list_metadata: ListMetadata


# Inherits from the methods of UpdateAutomationRequest
class UpdateMailboxRequest(UpdateAutomationRequest):
    # ------------------------------
    # HTTP Config
    # ------------------------------
    webhook_url: Optional[HttpUrl] = None
    webhook_headers: Optional[Dict[str, str]] = None

    # ------------------------------
    # DocumentProcessing Parameters
    # ------------------------------
    image_resolution_dpi: Optional[int] = None
    browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None
    modality: Optional[Modality] = None
    model: Optional[str] = None
    temperature: Optional[float] = None
    json_schema: Optional[Dict] = None
    reasoning_effort: Optional[ChatCompletionReasoningEffort] = None

    # ------------------------------
    # Email Specific config
    # ------------------------------
    authorized_domains: Optional[list[str]] = None
    authorized_emails: Optional[List[EmailStr]] = None

    @field_validator("authorized_emails", mode="before")
    def normalize_authorized_emails(cls, emails: Optional[List[str]]) -> Optional[List[str]]:
        return [email.strip().lower() for email in emails] if emails else None
