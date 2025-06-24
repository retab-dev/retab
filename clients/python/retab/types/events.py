import datetime
from typing import Any, Literal, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, Field

metadata_key = Literal[
    "automation",
    "cron",
    "data_structure",
    "dataset",
    "dataset_membership",
    "endpoint",
    "evaluation",
    "extraction",
    "file",
    "files",
    "link",
    "mailbox",
    "organization",
    "outlook",
    "preprocessing",
    "preprocessing",
    "reconciliation",
    "schema",
    "schema_data",
    "template",
    "user",
    "webhook",
]

event_type = Literal[
    "extraction.created",
    "messages.created",
    "document.orientation_corrected",
    "consensus.reconciled",
    "automation.created",
    "automation.updated",
    "automation.deleted",
    "automation.webhook",
    "preprocessing.created",
    "link.created",
    "link.updated",
    "link.deleted",
    "link.webhook",
    "mailbox.created",
    "mailbox.updated",
    "mailbox.deleted",
    "mailbox.webhook",
    "outlook.created",
    "outlook.updated",
    "outlook.deleted",
    "outlook.webhook",
    "schema.generated",
    "schema.promptified",
    "schema.system_promptfile.created",
    "file.updated",
    "file.deleted",
    "template.created",
    "template.deleted",
    "template.sample_document_uploaded",
    "template.sample_document_deleted",
    "template.updated",
]


class Event(BaseModel):
    object: Literal["event"] = "event"
    id: str = Field(default_factory=lambda: "event_" + nanoid.generate(), description="Unique identifier for the event")
    event: str = Field(..., description="A string that distinguishes the event type. Ex: user.created, user.updated, user.deleted, etc.")
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    data: dict[str, Any] = Field(..., description="Event payload. Payloads match the corresponding API objects.")
    metadata: Optional[dict[metadata_key, str]] = Field(
        default=None, description="Ids giving informations about the event. Ex: user.created.metadata = {'user': 'usr_8478973619047837'}"
    )


class StoredEvent(Event):
    organization_id: str = Field(..., description="Organization ID")
