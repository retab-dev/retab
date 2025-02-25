from pydantic import BaseModel, Field
import mimetypes
import base64
from typing import Literal, Optional, Sequence, TypeAlias
import datetime
import re
import hashlib

def generate_blake2b_hash_from_bytes(bytes_: bytes) -> str:
    return hashlib.blake2b(bytes_, digest_size=8).hexdigest()

def generate_blake2b_hash_from_base64(base64_string: str) -> str:
    return generate_blake2b_hash_from_bytes(base64.b64decode(base64_string))


class MIMEData(BaseModel):
    filename: str = Field(description="The filename of the file", examples=["file.pdf", "image.png", "data.txt"])
    url: str = Field(description="The URL of the file", examples=["https://example.com/file.pdf", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIA..."])

    @property
    def id(self) -> str:
        return f"file_{generate_blake2b_hash_from_base64(self.content)}"
    
    @property
    def extension(self) -> str:
        return self.filename.split('.')[-1].lower()
    
    @property
    def content(self) -> str:
        if self.url.startswith('data:'):
            # Extract base64 content from data URL
            base64_content = self.url.split(',')[1]
            return base64_content
        else:
            raise ValueError("Content is not available for this file")
    
    @property
    def mime_type(self) -> str:
        if self.url.startswith('data:'):
            return self.url.split(';')[0].split(':')[1]
        else:
            return mimetypes.guess_type(self.filename)[0] or "application/octet-stream"
        
    @property 
    def unique_filename(self) -> str:
        return f"{self.id}.{self.extension}"

    
    @property
    def size(self) -> int:
        # size in bytes
        return len(base64.b64decode(self.content))
    
    def __str__(self) -> str:
        truncated_url = self.url[:50] + '...' if len(self.url) > 50 else self.url
        # truncated_content = self.content[:50] + '...' if len(self.content) > 50 else self.content
        return f"MIMEData(filename='{self.filename}', url='{truncated_url}', mime_type='{self.mime_type}', size='{self.size}', extension='{self.extension}')"
    
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

    def __str__(self) -> str:
        truncated_url = self.url[:50] + '...' if len(self.url) > 50 else self.url
        truncated_content = self.content[:50] + '...' if len(self.content) > 50 else self.content
        return f"BaseMIMEData(filename='{self.filename}', url='{truncated_url}', content='{truncated_content}', mime_type='{self.mime_type}', extension='{self.extension}')"
    
    def __repr__(self) -> str:
        return self.__str__()

    


# **** MIME DATACLASSES ****

DisplayType: TypeAlias = Literal["image", "pdf", "txt"]

class DisplayMetadata(BaseModel):
    url: str = Field(..., description="URL of the attachment for display purposes.")
    type : DisplayType = Field(..., description="Type of the attachment, determining how it should be handled.")

class AttachmentMetadata(BaseModel):
    is_inline: bool = Field(default=False, description="Whether the attachment is inline or not.")
    inline_cid: Optional[str] = Field(default=None, description="CID reference for inline attachments.")
    url: Optional[str] = Field(default=None, description="URL of the attachment.")
    display_metadata: Optional[DisplayMetadata] = Field(default=None, description="Display Metadata")
    source: Optional[str] = Field(default=None, description="Source of the attachment in dot notation attachment_id, or email_id.attachment_id, allow us to keep track of the origin of the attachment, for search purposes. ")

class BaseAttachmentMIMEData(BaseMIMEData):
    metadata: AttachmentMetadata = Field(AttachmentMetadata(), description="Additional metadata about the attachment.")

class AttachmentMIMEData(MIMEData):
    metadata: AttachmentMetadata = Field(AttachmentMetadata(), description="Additional metadata about the attachment.")




# **** EMAIL DATACLASSES ****

class EmailAddressData(BaseModel):
    email: str = Field(..., description="The email address")
    display_name: Optional[str] = Field(default=None, description="The display name associated with the email address")

    def __str__(self) -> str:
        if self.display_name:
            return (f"{self.display_name} <{self.email}>")
        else:
            return (f"<{self.email}>")

# Light EmailData object that can conveniently be stored in mongoDB for search
class BaseEmailData(BaseModel):
    id: str = Field(..., description="The Message-ID header of the email")
    tree_id: str = Field(..., description="The root email ID, which is references[0] if it exists, otherwise the email's ID")

    subject: Optional[str] = Field(default=None, description="The subject of the email")
    body_plain: Optional[str] = Field(default=None, description="The plain text body of the email")
    body_html: Optional[str] = Field(default=None, description="The HTML body of the email")
    sender: EmailAddressData = Field(..., description="The sender's email address information")
    recipients_to: list[EmailAddressData] = Field(..., description="List of primary recipients' email address information")
    recipients_cc: list[EmailAddressData] = Field(default=[], description="List of carbon copy recipients' email address information")
    recipients_bcc: list[EmailAddressData] = Field(default=[], description="List of blind carbon copy recipients' email address information")
    sent_at: datetime.datetime = Field(..., description="The date and time when the email was sent")
    received_at: Optional[datetime.datetime] = Field(default=None, description="The date and time when the email was received")
 
    in_reply_to: Optional[str] = Field(default=None, description="The Message-ID of the email this is replying to")
    references: list[str] = Field(default=[], description="List of Message-IDs this email references")
    headers: dict[str, str] = Field(default={}, description="Dictionary of email headers")
    
    url: Optional[str] = Field(default=None, description="URL where the email content can be accessed")

    attachments: Sequence[BaseAttachmentMIMEData] = Field(default=[], description="List of email attachments")

    @property
    def unique_filename(self) -> str:
        cleaned_id = re.sub(r'[\s<>]', '', self.id)
        return f"{cleaned_id}.eml"
    
    def __repr__(self) -> str:
        recipient_count = len(self.recipients_to) + len(self.recipients_cc) + len(self.recipients_bcc)
        attachment_count = len(self.attachments)
        
        subject_preview = self.subject
        body_preview = self.body_plain[:5000] + '...' if self.body_plain and len(self.body_plain) > 5000 else self.body_plain
        
        return (
            f"BaseEmailData("
            f"id='{self.id}', "
            f"subject='{subject_preview}', "
            f"body='{body_preview}', "
            f"sender='{self.sender.email}', "
            f"recipients={recipient_count}, "
            f"attachments={attachment_count}, "
            f"sent_at='{self.sent_at.strftime('%Y-%m-%d %H:%M:%S')}'"
            f")"
        )
    
    def __str__(self) -> str:
        return self.__repr__()

class EmailData(BaseEmailData):
    attachments: Sequence[AttachmentMIMEData] = Field([], description="List of email attachments") # type: ignore

