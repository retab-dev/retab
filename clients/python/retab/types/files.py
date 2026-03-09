import datetime
from typing import Literal, Optional

from pydantic import BaseModel, ConfigDict, Field


class File(BaseModel):
    model_config = ConfigDict(extra="ignore")
    object: Literal["file"] = "file"
    id: str = Field(..., description="The unique identifier of the file")
    filename: str = Field(..., description="The name of the file")
    organization_id: str = Field(..., description="The ID of the organization that owns the file")
    created_at: datetime.datetime = Field(..., description="When the file was created")
    updated_at: datetime.datetime = Field(..., description="When the file was last updated")
    page_count: Optional[int] = Field(default=None, description="Number of pages in the file")


class FileLink(BaseModel):
    download_url: str = Field(..., description="The signed URL to download the file")
    expires_in: str = Field(..., description="The expiration time of the signed URL")
    filename: str = Field(..., description="The name of the file")


class UploadFileResponse(BaseModel):
    file_id: str = Field(..., description="The ID of the uploaded file", alias="fileId")
    filename: str = Field(..., description="The filename of the uploaded file")

    model_config = ConfigDict(populate_by_name=True)
