import copy
import datetime
import json
from typing import Any, Literal, Optional

import nanoid  # type: ignore
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, HttpUrl, computed_field, field_serializer

from ..._utils.json_schema import clean_schema
from ..._utils.mime import generate_blake2b_hash_from_string
from ..logs import AutomationConfig, UpdateAutomationRequest
from ..modalities import Modality
from ..pagination import ListMetadata


class Endpoint(AutomationConfig):
    object: Literal['deployment.endpoint'] = "deployment.endpoint"
    id: str = Field(default_factory=lambda: "endp_" + nanoid.generate(), description="Unique identifier for the extraction endpoint")

class ListEndpoints(BaseModel):
    data: list[Endpoint]
    list_metadata: ListMetadata


# Inherits from the methods of UpdateAutomationRequest
class UpdateEndpointRequest(UpdateAutomationRequest):
    pass
