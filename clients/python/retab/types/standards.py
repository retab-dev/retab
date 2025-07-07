from typing import Any, List, Literal, Optional, Tuple, TypeVar, TypedDict

from pydantic import BaseModel, Field
from pydantic_core import PydanticUndefined

# API Standards

# Define a type variable to represent the content type
T = TypeVar("T")

FieldUnset: Any = PydanticUndefined


# Define the ErrorDetail model
class ErrorDetail(BaseModel):
    code: str
    message: str
    details: Optional[dict] = None


class StandardErrorResponse(BaseModel):
    detail: ErrorDetail


class StreamingBaseModel(BaseModel):
    streaming_error: ErrorDetail | None = None


class DocumentPreprocessResponseContent(BaseModel):
    messages: list[dict[str, Any]] = Field(..., description="Messages generated during the preprocessing of the document")
    json_schema: dict[str, Any] = Field(..., description="Generated JSON Schema for Structured Output OpenAI Completions")


class PreparedRequest(BaseModel):
    method: Literal["POST", "GET", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE"]
    url: str
    data: dict | None = None
    params: dict | None = None
    form_data: dict | None = None
    files: dict | List[Tuple[str, Tuple[str, bytes, str]]] | None = None
    idempotency_key: str | None = None
    raise_for_status: bool = False


class DeleteResponse(TypedDict):
    """Response from a delete operation"""

    success: bool
    id: str


class ExportResponse(TypedDict):
    """Response from an export operation"""

    success: bool
    path: str
