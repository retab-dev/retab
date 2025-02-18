from pydantic import BaseModel, Field
from typing import Literal, Any, Optional
import uuid
import datetime



metadata_key = Literal['user', 'organization', 'link', 'mailbox', 'cron', 'outlook', 'extraction', 'webhook', 'schema', 'data_structure']

event_type = Literal['extraction.created', 
                     'link.created', 'link.updated', 'link.deleted', 'link.webhook',
                     'mailbox.created', 'mailbox.updated', 'mailbox.deleted','mailbox.webhook',
                     'outlook.created', 'outlook.updated', 'outlook.deleted','outlook.webhook',
                     'schema.generated', 'schema.promptified'
                     ]

class Event(BaseModel):
    object: Literal['event'] = "event"
    id: str = Field(default_factory=lambda: "event_" + str(uuid.uuid4()), description="Unique identifier for the event")
    event: str = Field(..., description="A string that distinguishes the event type. Ex: user.created, user.updated, user.deleted, etc.")
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    data: dict[str, Any] = Field(..., description="Event payload. Payloads match the corresponding API objects.")
    metadata: Optional[dict[metadata_key, str]] = Field(default=None, description="Ids giving informations about the event. Ex: user.created.metadata = {'user': 'usr_8478973619047837'}")


class StoredEvent(Event):
    organization_id: str = Field(..., description="Organization ID")