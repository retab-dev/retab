from pydantic import BaseModel, Field, ConfigDict
import re

import base64
class BaseMIMEData(BaseModel):
    id: str = Field(..., description="Unique identifier for the attachment")
    name: str = Field(..., description="Original filename of the attachment.")
    mime_type: str = Field(..., description="MIME type of the attachment, determining how it should be handled.")

    def __repr__(self) -> str:
        return f"BaseMIMEData(id={self.id!r}, name={self.name!r}, type={self.mime_type!r})"


    @property
    def extension(self) -> str:
        return self.name.split('.')[-1].lower()

    @property
    def unique_filename(self) -> str:
        cleaned_id = re.sub(r'[\s<>]', '', self.id)
        return f"{cleaned_id}.{self.extension}"
    
    model_config = ConfigDict(extra='ignore')

class MIMEData(BaseMIMEData):
    content: str = Field(..., description="Base64-encoded content of the attachment.")

    @property 
    def size(self) -> int:
        return len(base64.b64decode(self.content))

