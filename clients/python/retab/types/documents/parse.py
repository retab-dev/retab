from typing import Literal
from pydantic import BaseModel, ConfigDict, Field

from ..mime import MIMEData, BaseMIMEData
from ..browser_canvas import BrowserCanvas

TableParsingFormat = Literal["markdown", "yaml", "html", "json"]


class RetabUsage(BaseModel):
    """Usage information for document processing."""

    page_count: int = Field(..., description="Number of pages processed")
    credits: float = Field(..., description="Credits consumed for processing")


class ParseRequest(BaseModel):
    """Request model for document parsing."""
    model_config = ConfigDict(extra="ignore")

    document: MIMEData = Field(..., description="Document to parse")
    model: str = Field(default="gemini-2.5-flash", description="Model to use for parsing")
    table_parsing_format: TableParsingFormat = Field(default="html", description="Format for parsing tables")
    image_resolution_dpi: int = Field(default=96, description="DPI for image processing")
    browser_canvas: BrowserCanvas = Field(default="A4", description="Canvas size for document rendering")


class ParseResult(BaseModel):
    """Result of document parsing."""

    document: BaseMIMEData = Field(..., description="Processed document metadata")
    usage: RetabUsage = Field(..., description="Processing usage information")
    pages: list[str] = Field(..., description="Text content of each page")
    text: str = Field(..., description="Text content of the document")
