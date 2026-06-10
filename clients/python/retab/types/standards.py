from enum import Enum
from typing import Any, List, Literal, Optional, Tuple, TypeVar
from typing_extensions import TypedDict

from pydantic import Field
from retab.types.base import RetabBaseModel


T = TypeVar("T")


class _Unset(Enum):
    """Sentinel for parameters that were not provided by the caller."""

    UNSET = "UNSET"


UNSET = _Unset.UNSET

# Deprecated alias kept for backward compatibility.
FieldUnset: Any = UNSET


class ErrorDetail(RetabBaseModel):
    code: str
    message: str
    details: Optional[dict] = None


class StandardErrorResponse(RetabBaseModel):
    detail: ErrorDetail


class StreamingBaseModel(RetabBaseModel):
    streaming_error: ErrorDetail | None = None


class DocumentPreprocessResponseContent(RetabBaseModel):
    messages: list[dict[str, Any]] = Field(..., description="Messages generated during the preprocessing of the document")
    json_schema: dict[str, Any] = Field(..., description="Generated JSON Schema for Structured Output OpenAI Completions")


class PreparedRequest(RetabBaseModel):
    method: Literal["POST", "GET", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE"]
    url: str
    data: Any = None
    params: dict | None = None
    form_data: dict | None = None
    files: dict | List[Tuple[str, Tuple[str, bytes, str]]] | None = None
    idempotency_key: str | None = None
    raise_for_status: bool = False


class DeleteResponse(TypedDict):
    """Response from a delete operation."""

    success: bool
    id: str


class ExportResponse(TypedDict):
    """Response from an export operation."""

    success: bool
    path: str


class CountResponse(RetabBaseModel):
    """Response from a resource `GET /count` route: the number of matching records.

    A shared named component so every count route exposes a typed `count` field
    (not a loose `dict`), keeping the OpenAPI->TypeScript bridge typed.
    """

    count: int
