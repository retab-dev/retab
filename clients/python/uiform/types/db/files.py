from typing import Literal, Tuple, BinaryIO
from pydantic import BaseModel, Field, HttpUrl, ConfigDict
import uuid
import mimetypes



class DBFile(BaseModel):
    """Represents the core file object in your new spec."""
    object: Literal["file"] = "file"
    id: str = Field(default_factory=lambda: f"file_{uuid.uuid4()}", description="The unique identifier of the file")
    filename: str = Field(description="The name of the file")

    @property
    def mime_type(self) -> str:
        return mimetypes.guess_type(self.filename)[0] or "application/octet-stream"
    
    @property
    def extension(self) -> str:
        return self.filename.split(".")[-1]
    
    model_config = ConfigDict(arbitrary_types_allowed=True)


FileData = Tuple[str, BinaryIO, str]
FileTuple = Tuple[str, FileData]

class FileLink(BaseModel):
    download_url: HttpUrl = Field(description="The signed URL to download the file")
    expires_in: str = Field(description="The expiration time of the signed URL")
    filename: str = Field(description="The name of the file")
