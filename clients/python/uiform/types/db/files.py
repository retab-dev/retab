import mimetypes
from typing import BinaryIO, Literal, Tuple

from pydantic import BaseModel, ConfigDict, Field, HttpUrl, field_serializer


class DBFile(BaseModel):
    """Represents the core file object in your new spec."""

    object: Literal["file"] = "file"
    id: str = Field(..., description="The unique identifier of the file. It is generated from the content 'file_{sha256(content)}'")
    filename: str = Field(..., description="The name of the file")

    @property
    def mime_type(self) -> str:
        return mimetypes.guess_type(self.filename)[0] or "application/octet-stream"

    @property
    def extension(self) -> str:
        return self.filename.split(".")[-1].lower()

    model_config = ConfigDict(arbitrary_types_allowed=True)


FileData = Tuple[str, BinaryIO, str]
FileTuple = Tuple[str, FileData]


class FileLink(BaseModel):
    download_url: HttpUrl = Field(description="The signed URL to download the file")
    expires_in: str = Field(description="The expiration time of the signed URL")
    filename: str = Field(description="The name of the file")

    @field_serializer('download_url')
    def url2str(self, val: HttpUrl) -> str:
        return str(val)
