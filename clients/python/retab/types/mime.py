import base64
import datetime
import gzip
import mimetypes
import re
from typing import Any, Optional, Self, Sequence

from pydantic import BaseModel, Field, field_validator
from ..utils.hashing import generate_blake2b_hash_from_base64

import io

# Add webp and heic to the list of supported mime types
mimetypes.add_type("image/webp", ".webp")
mimetypes.add_type("image/heic", ".heic")
mimetypes.add_type("image/heif", ".heif")

# **** OCR DATACLASSES (DocumentAI-compatible) ****
class Point(BaseModel):
    x: int
    y: int


class Matrix(BaseModel):
    """Representation for transformation matrix, compatible with OpenCV format.

    This represents transformation matrices that were applied to the original
    document image to produce the processed page image.
    """

    rows: int = Field(description="Number of rows in the matrix")
    cols: int = Field(description="Number of columns in the matrix")
    type_: int = Field(description="OpenCV data type (e.g., 0 for CV_8U)")
    data: str = Field(description="The matrix data compressed with gzip and encoded as base64 string for JSON serialization")

    @property
    def data_bytes(self) -> bytes:
        """Get the matrix data as bytes."""
        # Decode base64 then decompress with gzip
        compressed_data = base64.b64decode(self.data)
        return gzip.decompress(compressed_data)

    @classmethod
    def from_bytes(cls, rows: int, cols: int, type_: int, data_bytes: bytes) -> Self:
        """Create a Matrix from raw bytes data."""
        # Compress with gzip then encode with base64
        compressed_data = gzip.compress(data_bytes, compresslevel=6)  # Good balance of speed vs compression
        encoded_data = base64.b64encode(compressed_data).decode("utf-8")
        return cls(rows=rows, cols=cols, type_=type_, data=encoded_data)


class TextBox(BaseModel):
    width: int
    height: int
    center: Point
    vertices: tuple[Point, Point, Point, Point] = Field(description="(top-left, top-right, bottom-right, bottom-left)")
    text: str

    # @field_validator("width", "height")
    # @classmethod
    # def check_positive_dimensions(cls, v: int) -> int:
    #     if not isinstance(v, int) or v <= 0:
    #         raise ValueError(f"Dimension must be a positive integer, got {v}")
    #     return v


class Page(BaseModel):
    page_number: int
    width: int
    height: int
    unit: str = Field(default="pixels", description="The unit of the page dimensions")
    blocks: list[TextBox]
    lines: list[TextBox]
    tokens: list[TextBox]
    transforms: list[Matrix] = Field(default=[], description="Transformation matrices applied to the original document image")

    # @field_validator("width", "height")
    # @classmethod
    # def check_positive_dimensions(cls, v: int) -> int:
    #     if not isinstance(v, int) or v <= 0:
    #         raise ValueError(f"Page dimension must be a positive integer, got {v}")
    #     return v


class OCR(BaseModel):
    pages: list[Page]


class MIMEData(BaseModel):
    filename: str = Field(
        description="The filename of the file",
        examples=["file.pdf", "image.png", "data.txt"]
    )
    url: str = Field(
        description="The URL of the file in base64 format",
        examples=["data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIA..."]
    )

    # Internal resource
    _buffer: Optional[io.BytesIO] = None

    @property
    def id(self) -> str:
        return f"file_{generate_blake2b_hash_from_base64(self.content)}"

    @property
    def extension(self) -> str:
        return self.filename.split(".")[-1].lower()

    @property
    def content(self) -> str:
        if self.url.startswith("data:"):
            return self.url.split(",")[1]
        raise ValueError("Content is not available for this file")

    @property
    def mime_type(self) -> str:
        if self.url.startswith("data:"):
            return self.url.split(";")[0].split(":")[1]
        return mimetypes.guess_type(self.filename)[0] or "application/octet-stream"

    @property
    def unique_filename(self) -> str:
        return f"{self.id}.{self.extension}"

    @property
    def size(self) -> int:
        return len(base64.b64decode(self.content))

    # def to_bytesio(self) -> io.BytesIO:
    #     """Decode base64 and return a BytesIO (without leaking references)."""
    #     buf = io.BytesIO(base64.b64decode(self.content))
    #     buf.seek(0)
    #     return buf

    # # -------- Context manager interface --------

    # def __enter__(self) -> io.BytesIO:
    #     """Opens the internal buffer so you can use it like a file."""
    #     if self._buffer is None:
    #         self._buffer = self.to_bytesio()
    #     return self._buffer

    # def __exit__(self, exc_type, exc_val, exc_tb):
    #     """Close and cleanup the buffer."""
    #     if self._buffer is not None:
    #         self._buffer.close()
    #         self._buffer = None

    # # -------- Optional convenience methods --------

    # def open(self) -> io.BytesIO:
    #     """Manual open without `with`."""
    #     return self.__enter__()

    # def close(self):
    #     """Manual close."""
    #     self.__exit__(None, None, None)

    def __str__(self) -> str:
        truncated_url = self.url[:50] + "..." if len(self.url) > 50 else self.url
        return (
            f"MIMEData(filename='{self.filename}', "
            f"url='{truncated_url}', "
            f"mime_type='{self.mime_type}', "
            f"size='{self.size}', "
            f"extension='{self.extension}')"
        )

    def __repr__(self) -> str:
        return self.__str__()


class BaseMIMEData(BaseModel):
    id: str = Field(..., description="ID of the file")
    filename: str = Field(..., description="Filename of the file")
    mime_type: str = Field(..., description="MIME type of the file")

# **** MIME DATACLASSES ****
class AttachmentMetadata(BaseModel):
    is_inline: bool = Field(default=False, description="Whether the attachment is inline or not.")
    inline_cid: Optional[str] = Field(default=None, description="CID reference for inline attachments.")
    source: Optional[str] = Field(
        default=None,
        description="Source of the attachment in dot notation attachment_id, or email_id.attachment_id, allow us to keep track of the origin of the attachment, for search purposes. ",
    )


class BaseAttachmentMIMEData(BaseMIMEData):
    metadata: AttachmentMetadata = Field(default=AttachmentMetadata(), description="Additional metadata about the attachment.")


class AttachmentMIMEData(MIMEData):
    metadata: AttachmentMetadata = Field(default=AttachmentMetadata(), description="Additional metadata about the attachment.")


# **** EMAIL DATACLASSES ****


class EmailAddressData(BaseModel):
    email: str = Field(..., description="The email address")
    display_name: Optional[str] = Field(default=None, description="The display name associated with the email address")

    def __str__(self) -> str:
        if self.display_name:
            return f"{self.display_name} <{self.email}>"
        else:
            return f"<{self.email}>"


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
        cleaned_id = re.sub(r"[\s<>]", "", self.id)
        return f"{cleaned_id}.eml"

    def __repr__(self) -> str:
        recipient_count = len(self.recipients_to) + len(self.recipients_cc) + len(self.recipients_bcc)
        attachment_count = len(self.attachments)

        subject_preview = self.subject
        body_preview = self.body_plain[:5000] + "..." if self.body_plain and len(self.body_plain) > 5000 else self.body_plain

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
    attachments: Sequence[AttachmentMIMEData] = Field([], description="List of email attachments")  # type: ignore
