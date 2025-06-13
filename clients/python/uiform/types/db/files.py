import mimetypes
from typing import Any, BinaryIO, Literal, Tuple

from pydantic import BaseModel, ConfigDict, Field, HttpUrl, field_validator


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
    download_url: str = Field(description="The signed URL to download the file")
    expires_in: str = Field(description="The expiration time of the signed URL")
    filename: str = Field(description="The name of the file")

    @field_validator("download_url", mode="after")
    def validate_httpurl(cls, val: Any) -> Any:
        if isinstance(val, str):
            HttpUrl(val)
        return val
