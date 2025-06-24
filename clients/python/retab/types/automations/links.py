from typing import Optional

import nanoid  # type: ignore
from pydantic import BaseModel, Field, computed_field

from ..logs import AutomationConfig, UpdateAutomationRequest
from ..pagination import ListMetadata


class Link(AutomationConfig):
    @computed_field
    @property
    def object(self) -> str:
        return "automation.link"

    id: str = Field(default_factory=lambda: "lnk_" + nanoid.generate(), description="Unique identifier for the extraction link")

    # Link Specific Config
    password: Optional[str] = Field(None, description="Password to access the link")


class ListLinks(BaseModel):
    data: list[Link]
    list_metadata: ListMetadata


# Inherits from the methods of UpdateAutomationRequest
class UpdateLinkRequest(UpdateAutomationRequest):
    # ------------------------------
    # Link Config
    # ------------------------------
    password: Optional[str] = None
