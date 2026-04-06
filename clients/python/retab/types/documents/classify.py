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
    first_n_pages: int | None = Field(default=None, description="Only use the first N pages of the document for classification. Useful for large documents where classification can be determined from early pages.")
    context: str | None = Field(default=None, description="Additional context for classification (e.g., iteration context from a loop)")
    n_consensus: int = Field(default=1, ge=1, le=8, description="Number of classification runs to use for consensus voting. Uses deterministic single-pass when set to 1.")
    bust_cache: bool = Field(default=False, description="If true, skip the LLM cache and force a fresh completion")


class ClassifyResult(BaseModel):
    reasoning: str = Field(..., description="The reasoning for the classification decision")
    classification: str = Field(..., description="The category name that the document belongs to")


class ClassifyVote(BaseModel):
    """A single LLM vote from a consensus classification run."""

    reasoning: str = Field(..., description="The reasoning produced by this classification vote")
    classification: str = Field(..., description="The category chosen by this classification vote")


class ClassifyResponse(BaseModel):
    result: ClassifyResult = Field(..., description="The classification result with reasoning")
    likelihood: float | None = Field(default=None, description="Likelihood score (0.0-1.0) of the consensus classification. Only set when n_consensus > 1 and at least two votes succeed.")
    votes: list[ClassifyVote] = Field(default_factory=list, description="Individual LLM votes used to build the consensus. Empty when n_consensus <= 1.")
    usage: RetabUsage = Field(..., description="Usage information for the classification")


class ClassifyOutputSchema(BaseModel):
    """Schema for LLM structured output."""
    reasoning: str = Field(
        ..., 
        description="Step-by-step reasoning explaining why this document belongs to the chosen category"
    )
    classification: str = Field(
        ...,
        description="The category name that this document belongs to"
    )
