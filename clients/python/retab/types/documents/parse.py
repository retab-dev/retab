from typing import Literal
from pydantic import BaseModel, Field

from ..mime import MIMEData, BaseMIMEData
from ..browser_canvas import BrowserCanvas

TableParsingFormat = Literal["text", "image", "yaml", "html"]


class RetabUsage(BaseModel):
    """Usage information for document processing."""

    page_count: int = Field(..., description="Number of pages processed")
    credits: float = Field(..., description="Credits consumed for processing")


class ParseRequest(BaseModel):
    """Request model for document parsing."""

    document: MIMEData = Field(..., description="Document to parse")
    fast_mode: bool = Field(default=False, description="Use fast mode for parsing (may reduce quality)")
    table_parsing_format: TableParsingFormat = Field(default="html", description="Format for parsing tables")
    image_resolution_dpi: int = Field(default=72, description="DPI for image processing")
    browser_canvas: BrowserCanvas = Field(default="A4", description="Canvas size for document rendering")


class ParseResult(BaseModel):
    """Result of document parsing."""

    document: BaseMIMEData = Field(..., description="Processed document metadata")
    usage: RetabUsage = Field(..., description="Processing usage information")
    pages: list[str] = Field(..., description="Text content of each page")
