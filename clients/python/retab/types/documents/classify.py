from __future__ import annotations

from pydantic import BaseModel, Field

from ..mime import MIMEData
from .usage import RetabUsage


class Category(BaseModel):
    name: str = Field(..., description="The name of the category")
    description: str = Field(default="", description="The description of the category")


class ClassifyRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to classify")
    categories: list[Category] = Field(..., description="The categories to classify the document into")
    model: str = Field(default="retab-small", description="The model to use for classification")
    first_n_pages: int | None = Field(
        default=None,
        description="Only use the first N pages of the document for classification. Useful for large documents where classification can be determined from early pages.",
    )
    context: str | None = Field(
        default=None,
        description="Additional context for classification (e.g., iteration context from a loop)",
    )
    n_consensus: int = Field(
        default=1,
        ge=1,
        le=16,
        description="Number of classification runs to use for consensus voting. Uses deterministic single-pass when set to 1.",
    )
    bust_cache: bool = Field(default=False, description="If true, skip the LLM cache and force a fresh completion")


class ClassifyDecision(BaseModel):
    reasoning: str = Field(..., description="The reasoning for the classification decision")
    category: str = Field(..., description="The category name that the document belongs to")


class ClassifyChoice(BaseModel):
    classification: ClassifyDecision = Field(..., description="One alternative classification vote output")


class ClassifyConsensus(BaseModel):
    choices: list[ClassifyChoice] = Field(
        default_factory=list,
        description="Alternative classification vote outputs used to build the consolidated result.",
    )
    likelihood: float | None = Field(
        default=None,
        description="Consensus likelihood score (0.0-1.0) of the winning classification.",
    )


class ClassifyResponse(BaseModel):
    classification: ClassifyDecision = Field(..., description="The classification result with reasoning")
    consensus: ClassifyConsensus = Field(
        default_factory=ClassifyConsensus,
        description="Consensus metadata for multi-vote classification runs",
    )
    usage: RetabUsage = Field(..., description="Usage information for the classification")
