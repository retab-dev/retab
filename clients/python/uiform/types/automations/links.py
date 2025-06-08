import copy
import datetime
import json
from typing import Any, Dict, Literal, Optional

import nanoid  # type: ignore
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, HttpUrl, computed_field, field_serializer
from pydantic_core import Url

from ..._utils.json_schema import clean_schema
from ..._utils.mime import generate_blake2b_hash_from_string
from ..logs import AutomationConfig, UpdateAutomationRequest
from ..modalities import Modality
from ..pagination import ListMetadata


class Link(AutomationConfig):
    object: Literal['deployment.link'] = "deployment.link"
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
