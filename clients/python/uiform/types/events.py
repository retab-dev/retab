from pydantic import BaseModel, Field
from typing import Literal, Any
import uuid
import datetime


class Events(BaseModel):
    object: Literal['event'] = "event"
    id: str = Field(default_factory=lambda: "event_" + str(uuid.uuid4()), description="Unique identifier for the event")
    event: str = Field(..., description="A string that distinguishes the event type. Ex: user.created, user.updated, user.deleted, etc.")
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    data: dict[str, Any] = Field(..., description="Event payload. Payloads match the corresponding API objects.")

