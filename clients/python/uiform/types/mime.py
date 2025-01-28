from pydantic import BaseModel, computed_field
import mimetypes
import base64
import httpx
from typing import Literal
import hashlib

from pydantic import Field

def generate_sha_hash_from_bytes(bytes_: bytes, hash_algorithm_: Literal['sha256', 'sha1'] = 'sha256') -> str:
    hash_algorithm = hashlib.sha256() if hash_algorithm_ == 'sha256' else hashlib.sha1()
    hash_algorithm.update(bytes_)
    hash_hex = hash_algorithm.hexdigest()
    return hash_hex

def generate_sha_hash_from_base64(base64_string: str, hash_algorithm_: Literal['sha256', 'sha1'] = 'sha256') -> str:
    # Decode the base64 string to bytes, Generate the SHA-256 hash of the bytes, Convert the hash to a hex string
    return generate_sha_hash_from_bytes(base64.b64decode(base64_string), hash_algorithm_=hash_algorithm_)

class MIMEData(BaseModel):
    filename: str = Field(description="The filename of the file", examples=["file.pdf", "image.png", "data.txt"])
    url: str = Field(description="The URL of the file", examples=["https://example.com/file.pdf", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIA..."])

    @property
    def extension(self) -> str:
        return self.filename.split('.')[-1].lower()
    
    @property
    @computed_field
    def content(self) -> str:
        if self.url.startswith('data:'):
            # Extract base64 content from data URL
            base64_content = self.url.split(',')[1]
            return base64_content
        else:
            raise ValueError("Content is not available for this file")
    
    @property
    @computed_field
    def mime_type(self) -> str:
        if self.url.startswith('data:'):
            return self.url.split(';')[0].split(':')[1]
        else:
            return mimetypes.guess_type(self.filename)[0] or "application/octet-stream"
        
    @property 
    def unique_filename(self) -> str:
        return f"{generate_sha_hash_from_base64(self.content)}.{self.extension}"

    
    @property
    def size(self) -> int:
        # size in bytes
        return len(base64.b64decode(self.content))
    
    def __str__(self) -> str:
        truncated_url = self.url[:50] + '...' if len(self.url) > 50 else self.url
        truncated_content = self.content[:50] + '...' if len(self.content) > 50 else self.content
        return f"MIMEData(filename='{self.filename}', url='{truncated_url}', content='{truncated_content}', mime_type='{self.mime_type}', size='{self.size}', extension='{self.extension}')"
    
    def __repr__(self) -> str:
        return self.__str__()

    
from typing import Any, Self

class BaseMIMEData(MIMEData):
    @classmethod
    def model_validate(cls, obj: Any, *, strict: bool | None = None, from_attributes: bool | None = None, context: Any | None = None) -> Self:
        if isinstance(obj, MIMEData):
            # Convert MIMEData instance to dict
            obj = obj.model_dump()
        if isinstance(obj, dict) and 'url' in obj:
            obj['url'] = obj['url'][:1000]  # Truncate URL to 1000 chars
        return super().model_validate(obj, strict=strict, from_attributes=from_attributes, context=context)
