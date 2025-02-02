from pydantic import BaseModel, Field,  HttpUrl, EmailStr, field_validator, computed_field
from typing import Any, Literal, List, Dict, ClassVar, Optional
import uuid
import datetime
import os
import re 
import copy
import json

from ..image_settings import ImageSettings
from ..modalities import Modality
from ..pagination import ListMetadata

from ..._utils.json_schema import clean_schema
from ..._utils.mime import generate_sha_hash_from_string

domain_pattern = re.compile(r"^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$")

class Mailbox(BaseModel):
    EMAIL_PATTERN: ClassVar[str] = f".*@{os.getenv('EMAIL_DOMAIN', 'devmail.uiform.com')}$"
    object: Literal['mailbox'] = "mailbox"
    id: str = Field(default_factory=lambda: "mb_" + str(uuid.uuid4()), description="Unique identifier for the mailbox")
    
    # Email Specific config
    email: str = Field(..., pattern=EMAIL_PATTERN)
    authorized_domains: list[str] = Field(default_factory=list, description = "List of authorized domains to receive the emails from")
    authorized_emails: List[EmailStr] = Field(default_factory=list, description = "List of emails to access the link")

    # Automation Config
    webhook_url: HttpUrl = Field(..., description = "Url of the webhook to send the data to")
    webhook_headers: Dict[str, str] = Field(default_factory=dict, description = "Headers to send with the request")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))

    # DocumentExtraction Config
    modality: Modality
    image_settings : ImageSettings = Field(default_factory=ImageSettings, description="Preprocessing operations applied to image before sending them to the llm")
    model: str = Field(..., description="Model used for chat completion")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])


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
    
    @computed_field   # type: ignore
    @property
    def schema_data_id(self) -> str:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.
        
        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        return "sch_data_id_" + generate_sha_hash_from_string(
            json.dumps(
                clean_schema(copy.deepcopy(self.json_schema), remove_custom_fields=True, fields_to_remove=["description", "default", "title", "required", "examples", "deprecated", "readOnly", "writeOnly"]),
                sort_keys=True).strip(), 
            "sha1")

    # This is a computed field, it is exposed when serializing the object
    @computed_field   # type: ignore
    @property
    def schema_id(self) -> str:
        """Returns the SHA1 hash of the complete schema.
        
        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        return "sch_id_" + generate_sha_hash_from_string(json.dumps(self.json_schema, sort_keys=True).strip(), "sha1")



class ListMailboxes(BaseModel):
    data: list[Mailbox]
    list_metadata: ListMetadata


class UpdateMailboxRequest(BaseModel):

    authorized_domains: Optional[list[str]] = None
    authorized_emails: Optional[List[EmailStr]] = None

    # ------------------------------
    # HTTP Config
    # ------------------------------
    webhook_url: Optional[HttpUrl] = None
    webhook_headers: Optional[Dict[str, str]] = None

    # ------------------------------
    # DocumentExtraction Parameters
    # ------------------------------
    # DocumentProcessing Parameters
    image_settings: Optional[ImageSettings] = None
    modality: Optional[Modality] = None
    # Others DocumentExtraction Parameters
    model: Optional[str] = None
    temperature: Optional[float] = None
    json_schema: Optional[Dict] = None

    @field_validator("authorized_emails", mode="before")
    def normalize_authorized_emails(cls, emails: List[str]) -> List[str]:
        return [email.strip().lower() for email in emails]

   