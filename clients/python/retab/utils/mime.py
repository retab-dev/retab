import base64
import hashlib
import io
import mimetypes
from pathlib import Path
from typing import Sequence, TypeVar, get_args
from urllib.parse import unquote_to_bytes

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
IMAGE_TYPES = Literal[".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".heic", ".heif"]
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


def _is_https_url_string(value: object) -> bool:
    return isinstance(value, str) and value.startswith("https://")


def _passthrough_https_url(url: str) -> MIMEData:
    """Build a MIMEData that references a remote https:// URL without fetching it.

    The backend resolves the URL server-side (see materialize_remote_mime) so large
    files don't traverse the API request body and can bypass the Cloud Run 32 MiB
    request cap. Filename is derived from the URL path; the backend validates the
    fetched content type after download.
    """
    last_segment = url.split("?", 1)[0].rsplit("/", 1)[-1]
    filename = last_segment or "remote_file"
    return MIMEData(filename=filename, url=url)


def _build_mime_data_from_data_url(data_url: str) -> MIMEData:
    """Decode an RFC 2397 ``data:`` URL into a MIMEData.

    Grammar: ``data:[<mediatype>][;base64],<data>``. The default media type
    when omitted is ``text/plain;charset=US-ASCII``. We split manually on the
    first ``,`` to stay transparent (urllib.urlopen would also handle this,
    but at the cost of an opaque dependency). Media-type parameters such as
    ``;charset=...`` are preserved in the rebuilt data URL so MIMEData's
    ``mime_type`` property still returns just the bare type.
    """
    # Strip the "data:" prefix, then split header from payload on the first comma.
    header, _, payload = data_url[len("data:") :].partition(",")

    # Detect and strip the ;base64 marker (case-insensitive, last parameter per RFC).
    is_base64 = False
    if header.lower().endswith(";base64"):
        is_base64 = True
        header = header[: -len(";base64")]

    mediatype = header or "text/plain;charset=US-ASCII"

    if is_base64:
        # Tolerate whitespace that some emitters add to long base64 blobs.
        file_bytes = base64.b64decode(payload, validate=False)
    else:
        file_bytes = unquote_to_bytes(payload)

    encoded_content = base64.b64encode(file_bytes).decode("utf-8")
    content_hash = hashlib.sha256(encoded_content.encode("utf-8")).hexdigest()

    # Derive a filename from the bare mime type so MIMEData.extension stays sensible.
    bare_mime_type = mediatype.split(";", 1)[0]
    extension = mimetypes.guess_extension(bare_mime_type) or ".bin"
    filename = f"inline_{content_hash[:16]}{extension}"

    return MIMEData(filename=filename, url=f"data:{mediatype};base64,{encoded_content}")


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

    if isinstance(document, str) and document.startswith("data:"):
        # RFC 2397 data URL: parse the mime type + (optional) ;base64 + payload.
        # Return a MIMEData directly so callers can pass small inline payloads
        # without writing them to disk first.
        return _build_mime_data_from_data_url(document)

    if _is_https_url_string(document):
        return _passthrough_https_url(document)  # type: ignore[arg-type]

    if hasattr(document, "unicode_string") and callable(getattr(document, "unicode_string")):
        url_str: str = document.unicode_string()  # type: ignore[union-attr]
        if url_str.startswith("https://"):
            return _passthrough_https_url(url_str)

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
        if not Path(filename).suffix:
            # A bare stream (e.g. io.BytesIO) carries no usable extension; sniff
            # the bytes the same way the `bytes` branch does so validation does
            # not reject a valid document just because the stream had no name.
            try:
                sniffed = puremagic.from_string(file_bytes)
                if sniffed.lower() in [".jpg", ".jpeg", ".jfif"]:
                    sniffed = ".jpeg"
            except Exception:
                sniffed = ".txt"
            filename = filename + sniffed
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
        # `document` is a path or a string; cast it to Path. Use an explicit
        # raise (not assert) so validation survives `python -O`.
        if not isinstance(document, (Path, str)):
            raise TypeError(f"Unsupported document type: {type(document).__name__}")
        pathdoc = Path(document)
        with open(pathdoc, "rb") as f:
            file_bytes = f.read()
        filename = pathdoc.name

    # Base64-encode
    encoded_content = base64.b64encode(file_bytes).decode("utf-8")

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
    # Explicit raise (not assert) so the check is not stripped under `python -O`.
    if "." + file_extension not in get_args(SUPPORTED_TYPES):
        raise ValueError(f"Invalid file type: {file_extension}. Must be one of: {get_args(SUPPORTED_TYPES)}")
