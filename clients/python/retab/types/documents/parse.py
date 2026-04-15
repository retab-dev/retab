from typing import Literal
from pydantic import BaseModel, ConfigDict, Field

from ..mime import MIMEData, FileRef
from .usage import RetabUsage
TableParsingFormat = Literal["markdown", "yaml", "html", "json"]




class ParseRequest(BaseModel):
    """Request model for document parsing."""
    model_config = ConfigDict(extra="ignore")

    document: MIMEData = Field(..., description="Document to parse")
    model: str = Field(default="retab-small", description="Model to use for parsing")
    table_parsing_format: TableParsingFormat = Field(default="html", description="Format for parsing tables")
    image_resolution_dpi: int = Field(default=192, description="DPI for image processing", ge=96, le=300)
    bust_cache: bool = Field(default=False, description="If true, skip the LLM cache and force a fresh completion")


class ParseResponse(BaseModel):
    """Result of document parsing."""

    document: FileRef = Field(..., description="Processed document metadata")
    usage: RetabUsage = Field(..., description="Processing usage information")
    pages: list[str] = Field(..., description="Text content of each page")
    text: str = Field(..., description="Text content of the document")
