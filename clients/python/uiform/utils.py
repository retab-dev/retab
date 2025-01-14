from typing import IO, Any, TypeVar, Union, Literal, get_args
from pathlib import Path
import json
import hashlib
import base64
from pathlib import Path

from .types.mime import MIMEData
from .types.ai_model import AIProvider, OpenAIModel, AnthropicModel, xAI_Model, GeminiModel



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

def load_json_schema(json_schema: Union[dict[str, Any], Path, str]) -> dict[str, Any]:
    """
    Load a JSON schema from either a dictionary or a file path.
    
    Args:
        json_schema: Either a dictionary containing the schema or a path to a JSON file
        
    Returns:
        dict[str, Any]: The loaded JSON schema
        
    Raises:
        JSONDecodeError: If the schema file contains invalid JSON
        FileNotFoundError: If the schema file doesn't exist
    """
    if isinstance(json_schema, (str, Path)):
        with open(json_schema) as f:
            return json.load(f)
    return json_schema




def find_provider_from_model(model: str) -> AIProvider:
    if model in get_args(OpenAIModel):
        return "OpenAI"
    elif ':' in model:
        # Handle fine-tuned models
        ft, base_model, model_id = model.split(':',2)
        if base_model in get_args(OpenAIModel):
            return "OpenAI"
    elif model in get_args(AnthropicModel):
        return "Anthropic"
    elif model in get_args(xAI_Model):
        return "xAI"
    elif model in get_args(GeminiModel):
        return "Gemini"
    raise ValueError(f"Could not determine AI provider for model: {model}")

    

def assert_valid_model_extraction(model: str) -> None:
    if model in get_args(OpenAIModel):
        return
    elif ':' in model:
        # Handle fine-tuned models
        ft, base_model, model_id = model.split(':',2)
        if base_model in get_args(OpenAIModel):
            return
    elif model in get_args(AnthropicModel):
        return
    elif model in get_args(xAI_Model):
        return
    elif model in get_args(GeminiModel):
        return
    raise ValueError(
        f"Invalid model for extraction: {model}.\n"
        f"Valid OpenAI models: {get_args(OpenAIModel)}\n"
        f"Valid Anthropic models: {get_args(AnthropicModel)}\n" 
        f"Valid xAI models: {get_args(xAI_Model)}\n"
        f"Valid Gemini models: {get_args(GeminiModel)}"
    )


def assert_valid_model_schema_generation(model: str) -> None:
    """Assert that the model is either a standard OpenAI model or a valid fine-tuned model.
    
    Valid formats:
    - Standard model: Must be in OpenAIModel
    - Fine-tuned model: Must be {base_model}:{id} where base_model is in OpenAIModel
    
    Raises:
        ValueError: If the model format is invalid
    """
    if model in get_args(OpenAIModel):
        return
    
    try:
        ft, base_model, model_id = model.split(':',2)
        if base_model not in get_args(OpenAIModel):
            raise ValueError(
                f"Invalid base model in fine-tuned model '{model}'. "
                f"Base model must be one of: {get_args(OpenAIModel)}"
            )
        if not model_id or not model_id.strip():
            raise ValueError(f"Model ID cannot be empty in fine-tuned model '{model}'")
    except ValueError as e:
        if ':' not in model:
            raise ValueError(
                f"Invalid model format: {model}. Must be either:\n"
                f"1. A standard model: {get_args(OpenAIModel)}\n"
                f"2. A fine-tuned model in format 'base_model:id' where base_model is one of the standard models"
            ) from None
        raise



import io
import mimetypes
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


from .types.modalities import SUPPORTED_TYPES

def assert_valid_file_type(file_extension: str) -> None:
    assert "." + file_extension in get_args(SUPPORTED_TYPES), f"Invalid file type: {file_extension}. Must be one of: {get_args(SUPPORTED_TYPES)}"