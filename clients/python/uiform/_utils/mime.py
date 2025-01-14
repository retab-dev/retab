from typing import IO, Any, TypeVar, Union, Literal, get_args
from pathlib import Path
import json
import hashlib
import base64
from pathlib import Path
import io
import mimetypes

from ..types.mime import MIMEData
from ..types.modalities import SUPPORTED_TYPES



def generate_sha_hash_from_bytes(bytes_: bytes, hash_algorithm_: Literal['sha256', 'sha1'] = 'sha256') -> str:
    hash_algorithm = hashlib.sha256() if hash_algorithm_ == 'sha256' else hashlib.sha1()
    hash_algorithm.update(bytes_)
    hash_hex = hash_algorithm.hexdigest()
    return hash_hex

def generate_sha_hash_from_base64(base64_string: str, hash_algorithm_: Literal['sha256', 'sha1'] = 'sha256') -> str:
    # Decode the base64 string to bytes, Generate the SHA-256 hash of the bytes, Convert the hash to a hex string
    return generate_sha_hash_from_bytes(base64.b64decode(base64_string), hash_algorithm_=hash_algorithm_)

def generate_sha_hash_from_string(input_string: str, hash_algorithm_: Literal['sha256', 'sha1'] = 'sha256') -> str:
    return generate_sha_hash_from_bytes(input_string.encode('utf-8'), hash_algorithm_=hash_algorithm_)

def generate_sha_hash_from_dict(input_dict: dict, hash_algorithm_: Literal['sha256', 'sha1'] = 'sha256') -> str:
    return generate_sha_hash_from_string(json.dumps(input_dict, sort_keys=True).strip(), hash_algorithm_=hash_algorithm_)


T = TypeVar('T')





def prepare_mime_document(document: Path | str | bytes | io.IOBase) -> MIMEData:
    """
    Convert documents (file paths or file-like objects) to MIMEData objects.
    
    Args:
        document: A path, string, bytes, or file-like object (IO[bytes])
        
    Returns:
        A MIMEData object
    """

    if isinstance(document, bytes):
        # `document` is already the raw bytes
        try: 
            import puremagic
            extension = puremagic.from_string(document)
        except: 
            extension = '.txt'
        file_bytes = document
        filename = "uploaded_file" + extension
    elif isinstance(document, io.IOBase):
        # `document` is a file-like object
        file_bytes = document.read()
        filename = getattr(document, "name", "uploaded_file")
        filename = Path(filename).name
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
    content_hash = hash_obj.hexdigest()

    # Guess MIME type based on file extension
    guessed_type, _ = mimetypes.guess_type(filename)
    mime_type = guessed_type or "application/octet-stream"

    # Build and return the MIMEData object
    mime_data = MIMEData(
        id=content_hash,
        name=filename,
        mime_type=mime_type,
        content=encoded_content
    )
    assert_valid_file_type(mime_data.extension)  # <-- Validate extension as needed

    return mime_data



def prepare_mime_document_list(documents: list[Path | str | bytes | io.IOBase])  -> list[MIMEData]:
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