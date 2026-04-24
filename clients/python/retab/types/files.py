import datetime
from typing import Literal, Optional

from pydantic import BaseModel, ConfigDict, Field

from .mime import MIMEData


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


class UploadFileResponse(MIMEData):
    model_config = ConfigDict(populate_by_name=True)


class CreateUploadResponse(BaseModel):
    file_id: str = Field(..., description="The ID of the upload session", alias="fileId")
    upload_url: str = Field(..., description="Short-lived signed upload URL", alias="uploadUrl")
    upload_method: str = Field(default="PUT", description="HTTP method for upload", alias="uploadMethod")
    upload_headers: dict[str, str] = Field(default_factory=dict, description="Headers required by the signed upload URL", alias="uploadHeaders")
    mime_data: MIMEData = Field(..., description="Durable Retab MIMEData reference", alias="mimeData")
    expires_at: datetime.datetime = Field(..., description="When the upload URL expires", alias="expiresAt")

    model_config = ConfigDict(populate_by_name=True)
