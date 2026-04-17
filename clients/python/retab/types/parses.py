from __future__ import annotations

import datetime
from typing import Literal, Optional

from pydantic import BaseModel, ConfigDict, Field

from .documents.usage import RetabUsage
from .mime import FileRef, MIMEData

TableParsingFormat = Literal["markdown", "yaml", "html", "json"]


class ParseRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    document: MIMEData = Field(..., description="The document to parse")
    model: str = Field(default="retab-small", description="The model to use for parsing")
    table_parsing_format: TableParsingFormat = Field(
        default="html",
        description="Format used to render tables extracted from the document",
    )
    image_resolution_dpi: int = Field(
        default=192,
        ge=96,
        le=300,
        description="DPI used when rasterizing pages for the parser",
    )
    bust_cache: bool = Field(default=False, description="If true, skip the LLM cache and force a fresh completion")


class ParseOutput(BaseModel):
    pages: list[str] = Field(..., description="Text content of each page (1-indexed order)")
    text: str = Field(..., description="Concatenated text content of the full document")


class Parse(BaseModel):
    id: str = Field(..., description="Unique identifier of the parse")
    file: FileRef = Field(..., description="Information about the parsed file")
    model: str = Field(..., description="Model used for parsing")
    table_parsing_format: TableParsingFormat = Field(
        ...,
        description="Format used to render tables extracted from the document",
    )
    image_resolution_dpi: int = Field(..., description="DPI used when rasterizing pages for the parser")
    output: ParseOutput = Field(..., description="The parsed document content")
    usage: Optional[RetabUsage] = Field(default=None, description="Usage information for the parse operation")
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None
