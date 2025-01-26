from typing import Literal
from pydantic import BaseModel, Field
import datetime
import uuid
class CollectionMembership(BaseModel):
    object: Literal["collection_membership"] = "collection_membership"
    id: str = Field(default_factory=lambda: f"cm_{uuid.uuid4()}", description="Unique identifier for the collection membership")
    file_id: str = Field(description="ID of the file that belongs to the collection")
    collection_id: str = Field(description="ID of the collection that contains the file") 
    created_at: datetime.datetime = Field(default_factory=datetime.datetime.now, description="Timestamp for when the membership was created")
    updated_at: datetime.datetime = Field(default_factory=datetime.datetime.now, description="Timestamp for when the membership was last updated")

