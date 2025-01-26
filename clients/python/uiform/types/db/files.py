from typing import Literal, Optional
from pydantic import BaseModel, Field
import uuid

class DBFile(BaseModel):
    """Represents the core file object in your new spec."""
    object: Literal["file"] = "file"
    id: str = Field(default_factory=lambda: f"file_{uuid.uuid4()}", description="The unique identifier of the file")
    filename: str = Field(description="The name of the file")
    mime_type: str = Field(description="The MIME type of the file (e.g., 'application/pdf', 'image/png')")
    size: Optional[int] = Field(default=None, description="The size of the file in bytes")
    organization_id: str = Field(description="The ID of the organization that owns the file")

    @property
    def extension(self) -> str:
        return self.filename.split(".")[-1]
    
    @property
    def gcs_path(self) -> str:
        return f"{self.organization_id}/file/{self.id}.{self.extension}" 

    class Config:
        from_attributes = True

