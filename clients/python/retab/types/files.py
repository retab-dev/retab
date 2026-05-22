import datetime
from typing import Literal, Optional

from pydantic import ConfigDict, Field
from retab.types.base import RetabBaseModel

from .mime import MIMEData


class File(RetabBaseModel):
    """Public file record. Mirrors the ``File`` component in
    ``open-source/docs/api-reference/openapi.json`` — persistence-only
    fields (``sha256``, ``size_bytes``, ``upload_*``, ``ocr_*``,
    ``mime_type``) live on the storage class and are projected away
    before responses leave the server.
    """

    model_config = ConfigDict(extra="ignore")
    object: Literal["file"] = Field(
        default="file",
        description='Resource discriminator. Always ``"file"``.',
    )
    id: str = Field(..., description="The unique identifier of the file")
    filename: str = Field(..., description="The name of the file")
    created_at: datetime.datetime = Field(..., description="When the file was created")
    updated_at: datetime.datetime = Field(..., description="When the file was last updated")
    page_count: Optional[int] = Field(default=None, description="Number of pages in the file")


class FileLink(RetabBaseModel):
    download_url: str = Field(..., description="The signed URL to download the file")
    expires_in: str = Field(..., description="The expiration time of the signed URL")
    filename: str = Field(..., description="The name of the file")
    mime_data: Optional[MIMEData] = Field(
        default=None,
        description="Durable Retab MIMEData reference associated with the file link",
    )


class UploadFileResponse(MIMEData):
    model_config = ConfigDict(populate_by_name=True)


class CreateUploadResponse(RetabBaseModel):
    file_id: str = Field(..., description="The ID of the upload session", alias="fileId")
    upload_url: str = Field(..., description="Short-lived signed upload URL", alias="uploadUrl")
    upload_method: str = Field(default="PUT", description="HTTP method for upload", alias="uploadMethod")
    upload_headers: dict[str, str] = Field(default_factory=dict, description="Headers required by the signed upload URL", alias="uploadHeaders")
    mime_data: MIMEData = Field(..., description="Durable Retab MIMEData reference", alias="mimeData")
    expires_at: datetime.datetime = Field(..., description="When the upload URL expires", alias="expiresAt")

    model_config = ConfigDict(populate_by_name=True)
