from typing import Literal, Optional
from pydantic import BaseModel, Field
import datetime
import uuid

class Collection(BaseModel):
    object: Literal["collection"] = "collection"
    id: str = Field(default_factory=lambda: f"col_{uuid.uuid4()}", description="Unique identifier of the collection")
    name: str = Field(description="Name of the collection")
    description: Optional[str] = Field(default=None, description="Optional description of the collection")
    created_at: datetime.datetime = Field(default_factory=datetime.datetime.now, description="Timestamp for when the collection was created")
    updated_at: datetime.datetime = Field(default_factory=datetime.datetime.now, description="Timestamp for when the collection was last updated")
