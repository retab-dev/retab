from typing import Literal
from pydantic import BaseModel, Field
import datetime
import uuid
class DatasetMembership(BaseModel):
    object: Literal["dataset_membership"] = "dataset_membership"
    id: str = Field(default_factory=lambda: f"dm_{uuid.uuid4()}", description="Unique identifier for the dataset membership")
    file_id: str = Field(description="ID of the file that belongs to the dataset")
    dataset_id: str = Field(description="ID of the dataset that contains the file") 
    created_at: datetime.datetime = Field(default_factory=datetime.datetime.now, description="Timestamp for when the membership was created")


