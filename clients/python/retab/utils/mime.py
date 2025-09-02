import base64
import hashlib
import io
import json
import mimetypes
from pathlib import Path
from typing import Sequence, TypeVar, get_args

import httpx
import PIL.Image
import puremagic
from pydantic import HttpUrl

from ..types.mime import MIMEData
from typing import Literal

EXCEL_TYPES = Literal[".xls", ".xlsx", ".ods"]
WORD_TYPES = Literal[".doc", ".docx", ".odt"]
PPT_TYPES = Literal[".ppt", ".pptx", ".odp"]
PDF_TYPES = Literal[".pdf"]
IMAGE_TYPES = Literal[".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp"]
TEXT_TYPES = Literal[
    ".txt",
    ".csv",
    ".tsv",
    ".md",
    ".log",
    ".xml",
    ".json",
    ".yaml",
    ".yml",
    ".rtf",
    ".ini",
    ".conf",
    ".cfg",
    ".nfo",
    ".srt",
    ".sql",
    ".sh",
    ".bat",
    ".ps1",
    ".js",
    ".jsx",
    ".ts",
    ".tsx",
    ".py",
    ".java",
    ".c",
    ".cpp",
    ".cs",
    ".rb",
    ".php",
    ".swift",
    ".kt",
    ".go",
    ".rs",
    ".pl",
    ".r",
    ".m",
    ".scala",
]
HTML_TYPES = Literal[".html", ".htm"]
WEB_TYPES = Literal[".mhtml"]
EMAIL_TYPES = Literal[".eml", ".msg"]
AUDIO_TYPES = Literal[".mp3", ".mp4", ".mpeg", ".mpga", ".m4a", ".wav", ".webm"]
SUPPORTED_TYPES = Literal[EXCEL_TYPES, WORD_TYPES, PPT_TYPES, PDF_TYPES, IMAGE_TYPES, TEXT_TYPES, HTML_TYPES, WEB_TYPES, EMAIL_TYPES, AUDIO_TYPES]


T = TypeVar("T")

def convert_pil_image_to_mime_data(image: PIL.Image.Image) -> MIMEData:
    """Convert a PIL Image object to a MIMEData object.

    Args:
        image: PIL Image object to convert

    Returns:
        MIMEData object containing the image data
    """
    # Convert PIL image to base64 string
    buffered = io.BytesIO()
    choosen_format = image.format if (image.format and image.format.lower() in ["png", "jpeg", "gif", "webp"]) else "JPEG"
    image.save(buffered, format=choosen_format)
    base64_content = base64.b64encode(buffered.getvalue()).decode("utf-8")

    content_hash = hashlib.sha256(base64_content.encode("utf-8")).hexdigest()

    # Create MIMEData object
    return MIMEData(filename=f"image_{content_hash}.{choosen_format.lower()}", url=f"data:image/{choosen_format.lower()};base64,{base64_content}")


def convert_mime_data_to_pil_image(mime_data: MIMEData) -> PIL.Image.Image:
    """Convert a MIMEData object to a PIL Image object.

    Args:
        mime_data: MIMEData object containing image data

    Returns:
        PIL Image object

    Raises:
        ValueError: If the MIMEData object does not contain image data
    """
    if not mime_data.mime_type.startswith("image/"):
        raise ValueError("MIMEData object does not contain image data")

    # Decode base64 content to bytes
    image_bytes = base64.b64decode(mime_data.content)

    # Create PIL Image from bytes
    image = PIL.Image.open(io.BytesIO(image_bytes))

    return image


def prepare_mime_document(document: Path | str | bytes | io.IOBase | MIMEData | PIL.Image.Image | HttpUrl) -> MIMEData:
    """
    Convert documents (file paths or file-like objects) to MIMEData objects.

    Args:
        document: A path, string, bytes, or file-like object (IO[bytes])

    Returns:
        A MIMEData object
    """
    # Check if document is a HttpUrl (Pydantic type)

    if isinstance(document, PIL.Image.Image):
        return convert_pil_image_to_mime_data(document)

    if isinstance(document, MIMEData):
        return document

    if isinstance(document, bytes):
        # `document` is already the raw bytes
        try:
            extension = puremagic.from_string(document)
            if extension.lower() in [".jpg", ".jpeg", ".jfif"]:
                extension = ".jpeg"
        except Exception:
            extension = ".txt"
        file_bytes = document
        filename = "uploaded_file" + extension
    elif isinstance(document, io.IOBase):
        # `document` is a file-like object
        file_bytes = document.read()
        filename = getattr(document, "name", "uploaded_file")
        filename = Path(filename).name
    elif hasattr(document, "unicode_string") and callable(getattr(document, "unicode_string")):
        with httpx.Client() as client:
            url: str = document.unicode_string()  # type: ignore
            response = client.get(url)
            response.raise_for_status()
            try:
                extension = puremagic.from_string(response.content)
                if extension.lower() in [".jpg", ".jpeg", ".jfif"]:
                    extension = ".jpeg"
            except Exception:
                extension = ".txt"
            file_bytes = response.content  # Fix: Use response.content instead of document
            filename = "uploaded_file" + extension
    else:
        # `document` is a path or a string; cast it to Path
        assert isinstance(document, (Path, str))
        pathdoc = Path(document)
        with open(pathdoc, "rb") as f:
            file_bytes = f.read()
        filename = pathdoc.name

    # Base64-encode
    encoded_content = base64.b64encode(file_bytes).decode("utf-8")
    # Compute SHA-256 hash over the *base64-encoded* content
    hash_obj = hashlib.sha256(encoded_content.encode("utf-8"))
    hash_obj.hexdigest()

    # Guess MIME type based on file extension
    guessed_type, _ = mimetypes.guess_type(filename)
    mime_type = guessed_type or "application/octet-stream"
    # Build and return the MIMEData object
    mime_data = MIMEData(filename=filename, url=f"data:{mime_type};base64,{encoded_content}")
    assert_valid_file_type(mime_data.extension)  # <-- Validate extension as needed

    return mime_data


def prepare_mime_document_list(documents: Sequence[Path | str | bytes | MIMEData | io.IOBase | PIL.Image.Image]) -> list[MIMEData]:
    """
    Convert documents (file paths or file-like objects) to MIMEData objects.

    Args:
        documents: List of document paths or file-like objects

    Returns:
        List of MIMEData objects
    """
    return [prepare_mime_document(doc) for doc in documents]


def assert_valid_file_type(file_extension: str) -> None:
    assert "." + file_extension in get_args(SUPPORTED_TYPES), f"Invalid file type: {file_extension}. Must be one of: {get_args(SUPPORTED_TYPES)}"
