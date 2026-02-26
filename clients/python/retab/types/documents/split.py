from pydantic import BaseModel, Field
from ..mime import MIMEData

class Subdocument(BaseModel):
    name: str = Field(..., description="The name of the subdocument")
    description: str = Field(..., description="The description of the subdocument")
    partition_key: str | None = Field(default=None, description="The key to partition the subdocument")


class SplitRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to split")
    subdocuments: list[Subdocument] = Field(..., description="The subdocuments to split the document into")
    model: str = Field(default="retab-small", description="The model to use to split the document")
    context: str | None = Field(default=None, description="Additional context for the split operation (e.g., iteration context from a loop)")
    n_consensus: int = Field(default=1, ge=1, le=8, description="Number of consensus split runs to perform. Uses deterministic single-pass when set to 1.")


class Partition(BaseModel):
    key: str = Field(..., description="The partition key value (e.g., property ID, invoice number)")
    pages: list[int] = Field(..., description="The pages of the partition (1-indexed)")
    first_page_y_start: float = Field(default=0.0, description="The y coordinate of the first page of the partition")
    last_page_y_end: float = Field(default=1.0, description="The y coordinate of the last page of the partition")

class SplitVote(BaseModel):
    """A single LLM vote from a consensus run."""
    pages: list[int] = Field(..., description="The pages assigned to this subdocument by this vote (1-indexed)")


class SplitResult(BaseModel):
    name: str = Field(..., description="The name of the subdocument")
    pages: list[int] = Field(..., description="The pages of the subdocument (1-indexed)")
    likelihood: float | None = Field(default=None, description="Likelihood score (0.0-1.0) of the split result. Only set when n_consensus > 1.")
    votes: list[SplitVote] = Field(default_factory=list, description="Individual LLM votes used to build the consensus. Empty when n_consensus <= 1.")
    partitions: list[Partition] = Field(default_factory=list, description="The partitions of the subdocument")


class SplitResponse(BaseModel):
    splits: list[SplitResult] = Field(..., description="The list of document splits with their page ranges")


class GenerateSplitConfigRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to analyze for automatic split configuration generation")
    model: str = Field(default="retab-small", description="The model to use for document analysis")


class GenerateSplitConfigResponse(BaseModel):
    subdocuments: list[Subdocument] = Field(..., description="The auto-generated subdocument definitions with optional partition keys")


class SplitOutputItem(BaseModel):
    """Internal schema item for LLM structured output validation."""
    name: str = Field(..., description="The name of the subdocument")
    start_page: int = Field(..., description="The start page of the subdocument (1-indexed)")
    end_page: int = Field(..., description="The end page of the subdocument (1-indexed, inclusive)")


class SplitOutputSchema(BaseModel):
    """Schema for LLM structured output."""
    splits: list[SplitOutputItem] = Field(
        ...,
        description="List of document sections, each classified into one of the provided subdocuments with their page ranges"
    )
