from typing import Any, Generic, Optional, TypeVar
from pydantic import BaseModel, Field

# API Standards

# Define a type variable to represent the content type
T = TypeVar("T")

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


