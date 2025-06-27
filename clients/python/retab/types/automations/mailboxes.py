import os
import re
from typing import ClassVar, List, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, EmailStr, Field, computed_field, field_validator

from ..logs import AutomationConfig, UpdateAutomationRequest
from ..pagination import ListMetadata

domain_pattern = re.compile(r"^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$")


class Mailbox(AutomationConfig):
    @computed_field
    @property
    def object(self) -> str:
        return "automation.mailbox"

    EMAIL_PATTERN: ClassVar[str] = f".*@{os.getenv('EMAIL_DOMAIN', 'mailbox.retab.com')}$"
    id: str = Field(default_factory=lambda: "mb_" + nanoid.generate(), description="Unique identifier for the mailbox")

    # Email Specific config
    email: str = Field(..., pattern=EMAIL_PATTERN)
    authorized_domains: list[str] = Field(default_factory=list, description="List of authorized domains to receive the emails from")
    authorized_emails: List[EmailStr] = Field(default_factory=list, description="List of emails to access the link")

    # Normalize email fields (case-insensitive)
    @field_validator("email", mode="before")
    def normalize_email(cls, value: str) -> str:
        return value.strip().lower()

    @field_validator("authorized_emails", mode="before")
    def normalize_authorized_emails(cls, emails: List[str]) -> List[str]:
        return [email.strip().lower() for email in emails]

    @field_validator("authorized_domains", mode="before")
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
    # Email Specific config
    # ------------------------------
    authorized_domains: Optional[list[str]] = None
    authorized_emails: Optional[List[EmailStr]] = None

    @field_validator("authorized_emails", mode="before")
    def normalize_authorized_emails(cls, emails: Optional[List[str]]) -> Optional[List[str]]:
        return [email.strip().lower() for email in emails] if emails else None
