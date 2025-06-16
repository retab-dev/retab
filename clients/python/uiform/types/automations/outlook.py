import re
from typing import Any, Dict, List, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, EmailStr, Field, computed_field, field_validator

from ..logs import AutomationConfig, UpdateAutomationRequest
from ..pagination import ListMetadata

domain_pattern = re.compile(r"^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$")


class AutomationLevel(BaseModel):
    distance_threshold: float = Field(default=0.9, description="Distance threshold for the automation")
    score_threshold: float = Field(default=0.9, description="Score threshold for the automation")


class MatchParams(BaseModel):
    endpoint: str = Field(..., description="Endpoint for matching parameters")
    headers: Dict[str, str] = Field(..., description="Headers for the request")
    path: str = Field(..., description="Path for matching parameters")


class FetchParams(BaseModel):
    endpoint: str = Field(..., description="Endpoint for fetching parameters")
    headers: Dict[str, str] = Field(..., description="Headers for the request")
    name: str = Field(..., description="Name of the fetch parameter")


class Outlook(AutomationConfig):
    @computed_field
    @property
    def object(self) -> str:
        return "automation.outlook"

    id: str = Field(default_factory=lambda: "outlook_" + nanoid.generate(), description="Unique identifier for the outlook")

    authorized_domains: list[str] = Field(default_factory=list, description="List of authorized domains to receive the emails from")
    authorized_emails: List[EmailStr] = Field(default_factory=list, description="List of emails to access the link")

    layout_schema: Optional[dict[str, Any]] = Field(default=None, description="Layout schema format used to display the data")

    # Optional Fields for data integration
    match_params: List[MatchParams] = Field(default_factory=list, description="List of match parameters for the outlook automation")
    fetch_params: List[FetchParams] = Field(default_factory=list, description="List of fetch parameters for the outlook automation")

    # @model_validator(mode="before")
    # @classmethod
    # def compute_layout_schema(cls, values: dict[str, Any]) -> dict[str, Any]:
    #     if values.get("layout_schema") is None:
    #         values["layout_schema"] = convert_schema_to_layout(values["json_schema"])
    #     return values


class ListOutlooks(BaseModel):
    data: list[Outlook]
    list_metadata: ListMetadata


# Inherits from the methods of UpdateAutomationRequest
class UpdateOutlookRequest(UpdateAutomationRequest):
    authorized_domains: Optional[list[str]] = None
    authorized_emails: Optional[List[EmailStr]] = None

    match_params: Optional[List[MatchParams]] = None
    fetch_params: Optional[List[FetchParams]] = None

    layout_schema: Optional[dict[str, Any]] = None

    @field_validator("authorized_emails", mode="before")
    def normalize_authorized_emails(cls, emails: Optional[List[str]]) -> Optional[List[str]]:
        return [email.strip().lower() for email in emails] if emails else None
