from __future__ import annotations

import datetime
from typing import Optional

from pydantic import BaseModel, Field

from .documents.usage import RetabUsage
from .extractions import ProcessingRequestOrigin
from .mime import FileRef, MIMEData


class Category(BaseModel):
    name: str = Field(..., description="The name of the category")
    description: str = Field(default="", description="The description of the category")


class ClassificationRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to classify")
    categories: list[Category] = Field(..., description="The categories to classify the document into")
    model: str = Field(default="retab-small", description="The model to use for classification")
    first_n_pages: int | None = Field(
        default=None,
        description="Only use the first N pages of the document for classification. Useful for large documents where classification can be determined from early pages.",
    )
    instructions: str | None = Field(
        default=None,
        description="Free-form instructions appended to the system prompt to steer the classification.",
    )
    n_consensus: int = Field(
        default=1,
        ge=1,
        le=16,
        description="Number of classification runs to use for consensus voting. Uses deterministic single-pass when set to 1.",
    )
    bust_cache: bool = Field(default=False, description="If true, skip the LLM cache and force a fresh completion")


class ClassificationDecision(BaseModel):
    reasoning: str = Field(..., description="The reasoning for the classification decision")
    category: str = Field(..., description="The category name that the document belongs to")


class ClassificationConsensus(BaseModel):
    choices: list[ClassificationDecision] = Field(
        default_factory=list,
        description="Alternative classification vote outputs used to build the consolidated result.",
    )
    likelihood: float | None = Field(
        default=None,
        description="Consensus likelihood score (0.0-1.0) of the winning classification.",
    )


class Classification(BaseModel):
    id: str = Field(..., description="Unique identifier of the classification")
    file: FileRef = Field(..., description="Information about the classified file")
    model: str = Field(..., description="Model used for classification")
    categories: list[Category] = Field(..., description="Categories the document was classified against")
    n_consensus: int = Field(default=1, description="Number of consensus votes used")
    instructions: str | None = Field(default=None, description="Free-form instructions supplied with the classification request.")
    output: ClassificationDecision = Field(..., description="The classification result with reasoning")
    consensus: ClassificationConsensus = Field(
        default_factory=ClassificationConsensus,
        description="Consensus metadata for multi-vote classification runs",
    )
    origin: Optional[ProcessingRequestOrigin] = Field(
        default=None,
        description="Origin of the classification request",
    )
    usage: Optional[RetabUsage] = Field(default=None, description="Usage information for the classification")
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None
