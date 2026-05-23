import base64
import datetime
import gzip
import io
import mimetypes
import re
from typing import Optional, Self, Sequence
from urllib.parse import urlsplit

from pydantic import Field
from retab.types.base import RetabBaseModel
from retab.utils.hashing import generate_blake2b_hash_from_base64


mimetypes.add_type("image/webp", ".webp")
mimetypes.add_type("image/heic", ".heic")
mimetypes.add_type("image/heif", ".heif")


class Point(RetabBaseModel):
    x: int
    y: int


class Matrix(RetabBaseModel):
    """Representation for transformation matrix, compatible with OpenCV format."""

    rows: int = Field(description="Number of rows in the matrix")
    cols: int = Field(description="Number of columns in the matrix")
    type_: int = Field(description="OpenCV data type")
    data: str = Field(description="The matrix data compressed with gzip and encoded as base64 string")

    @property
    def data_bytes(self) -> bytes:
        compressed_data = base64.b64decode(self.data)
        return gzip.decompress(compressed_data)

    @classmethod
    def from_bytes(cls, rows: int, cols: int, type_: int, data_bytes: bytes) -> Self:
        compressed_data = gzip.compress(data_bytes, compresslevel=6)
        encoded_data = base64.b64encode(compressed_data).decode("utf-8")
        return cls(rows=rows, cols=cols, type_=type_, data=encoded_data)


class TextBox(RetabBaseModel):
    width: int
    height: int
    center: Point
    vertices: tuple[Point, Point, Point, Point] = Field(description="top-left, top-right, bottom-right, bottom-left")
    text: str


class Page(RetabBaseModel):
    page_number: int
    width: int
    height: int
    unit: str = Field(default="pixels", description="The unit of the page dimensions")
    blocks: list[TextBox]
    lines: list[TextBox]
    tokens: list[TextBox]
    transforms: list[Matrix] = Field(default_factory=list, description="Transformation matrices applied to the original document image")


class OCR(RetabBaseModel):
    pages: list[Page]


class MIMEData(RetabBaseModel):
    filename: str = Field(description="The filename of the file", examples=["file.pdf", "image.png", "data.txt"])
    url: str = Field(description="The URL of the file in base64 format", examples=["data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIA..."])

    _buffer: Optional[io.BytesIO] = None

    @property
    def id(self) -> str:
        parsed_url = urlsplit(self.url)
        if parsed_url.scheme == "https" and parsed_url.hostname == "storage.retab.com" and not parsed_url.query and not parsed_url.fragment:
            path_parts = [part for part in parsed_url.path.split("/") if part]
            if len(path_parts) == 2:
                file_id, separator, extension = path_parts[1].rpartition(".")
                if path_parts[0] and file_id and separator == "." and extension:
                    return file_id
        return f"file_{generate_blake2b_hash_from_base64(self.content)}"

    @property
    def extension(self) -> str:
        return self.filename.split(".")[-1].lower()

    @property
    def content(self) -> str:
        if self.url.startswith("data:"):
            return self.url.split(",", 1)[1]
        raise ValueError("Content is not available for this file")

    @property
    def mime_type(self) -> str:
        if self.url.startswith("data:"):
            return self.url.split(";", 1)[0].split(":", 1)[1]
        return mimetypes.guess_type(self.filename)[0] or "application/octet-stream"

    @property
    def unique_filename(self) -> str:
        return f"{self.id}.{self.extension}"

    @property
    def size(self) -> int:
        return len(base64.b64decode(self.content))

    def __str__(self) -> str:
        truncated_url = self.url[:50] + "..." if len(self.url) > 50 else self.url
        try:
            size: int | str = self.size
        except ValueError:
            size = "unavailable"
        return f"MIMEData(filename='{self.filename}', url='{truncated_url}', mime_type='{self.mime_type}', size='{size}', extension='{self.extension}')"

    def __repr__(self) -> str:
        return self.__str__()


class FileRef(RetabBaseModel):
    """Public/shared file reference used across SDK and customer-facing APIs."""

    id: str = Field(..., description="ID of the file")
    filename: str = Field(..., description="Filename of the file")
    mime_type: str = Field(..., description="MIME type of the file")


class AttachmentMetadata(RetabBaseModel):
    is_inline: bool = Field(default=False, description="Whether the attachment is inline or not.")
    inline_cid: Optional[str] = Field(default=None, description="CID reference for inline attachments.")
    source: Optional[str] = Field(default=None, description="Source of the attachment in dot notation.")


class BaseAttachmentMIMEData(FileRef):
    metadata: AttachmentMetadata = Field(default_factory=AttachmentMetadata, description="Additional metadata about the attachment.")


class AttachmentMIMEData(MIMEData):
    metadata: AttachmentMetadata = Field(default_factory=AttachmentMetadata, description="Additional metadata about the attachment.")


class EmailAddressData(RetabBaseModel):
    email: str = Field(..., description="The email address")
    display_name: Optional[str] = Field(default=None, description="The display name associated with the email address")

    def __str__(self) -> str:
        if self.display_name:
            return f"{self.display_name} <{self.email}>"
        return f"<{self.email}>"


class BaseEmailData(RetabBaseModel):
    id: str = Field(..., description="The Message-ID header of the email")
    tree_id: str = Field(..., description="The root email ID")
    subject: Optional[str] = Field(default=None, description="The subject of the email")
    body_plain: Optional[str] = Field(default=None, description="The plain text body of the email")
    body_html: Optional[str] = Field(default=None, description="The HTML body of the email")
    sender: EmailAddressData = Field(..., description="The sender's email address information")
    recipients_to: list[EmailAddressData] = Field(..., description="List of primary recipients")
    recipients_cc: list[EmailAddressData] = Field(default_factory=list, description="List of carbon copy recipients")
    recipients_bcc: list[EmailAddressData] = Field(default_factory=list, description="List of blind carbon copy recipients")
    sent_at: datetime.datetime = Field(..., description="The date and time when the email was sent")
    received_at: Optional[datetime.datetime] = Field(default=None, description="The date and time when the email was received")
    in_reply_to: Optional[str] = Field(default=None, description="The Message-ID of the email this is replying to")
    references: list[str] = Field(default_factory=list, description="List of Message-IDs this email references")
    headers: dict[str, str] = Field(default_factory=dict, description="Dictionary of email headers")
    url: Optional[str] = Field(default=None, description="URL where the email content can be accessed")
    attachments: Sequence[BaseAttachmentMIMEData] = Field(default_factory=list, description="List of email attachments")

    @property
    def unique_filename(self) -> str:
        cleaned_id = re.sub(r"[\s<>]", "", self.id)
        return f"{cleaned_id}.eml"

    def __repr__(self) -> str:
        recipient_count = len(self.recipients_to) + len(self.recipients_cc) + len(self.recipients_bcc)
        attachment_count = len(self.attachments)
        subject_preview = self.subject
        body_preview = self.body_plain[:5000] + "..." if self.body_plain and len(self.body_plain) > 5000 else self.body_plain
        return (
            f"BaseEmailData(id='{self.id}', subject='{subject_preview}', body='{body_preview}', "
            f"sender='{self.sender.email}', recipients={recipient_count}, attachments={attachment_count}, "
            f"sent_at='{self.sent_at.strftime('%Y-%m-%d %H:%M:%S')}')"
        )

    def __str__(self) -> str:
        return self.__repr__()


class EmailData(BaseEmailData):
    attachments: Sequence[AttachmentMIMEData] = Field(default_factory=list, description="List of email attachments")  # type: ignore[assignment]
