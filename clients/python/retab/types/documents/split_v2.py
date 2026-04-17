from pydantic import BaseModel, Field

from ..mime import MIMEData
from .usage import RetabUsage


class Subdocument(BaseModel):
    name: str = Field(..., description="The name of the subdocument")
    description: str = Field(default="", description="The description of the subdocument")
    partition_key: str | None = Field(default=None, description="The key to partition the subdocument")
    allow_multiple_instances: bool = Field(
        default=False,
        description="When true, this subdocument type can appear more than once in the document — the split will identify each distinct instance (runs an extra vision-based refinement pass).",
    )


class SplitRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to split")
    subdocuments: list[Subdocument] = Field(
        ...,
        min_length=1,
        description="The subdocuments to split the document into",
    )
    model: str = Field(default="retab-small", description="The model to use to split the document")
    context: str | None = Field(
        default=None,
        description="Additional context for the split operation (e.g., iteration context from a loop)",
    )
    n_consensus: int = Field(
        default=1,
        ge=1,
        le=8,
        description="Number of consensus split runs to perform. Uses deterministic single-pass when set to 1.",
    )
    bust_cache: bool = Field(default=False, description="If true, skip the LLM cache and force a fresh completion")


class GenerateSplitConfigRequest(BaseModel):
    document: MIMEData = Field(
        ...,
        description="The document to analyze for automatic split configuration generation",
    )
    model: str = Field(default="retab-small", description="The model to use for document analysis")


class GenerateSplitConfigResponse(BaseModel):
    subdocuments: list[Subdocument] = Field(
        ...,
        description="The auto-generated subdocument definitions with optional partition keys",
    )


class Partition(BaseModel):
    key: str = Field(..., description="The partition key value (e.g., property ID, invoice number)")
    pages: list[int] = Field(..., description="The pages of the partition (1-indexed)")


class SplitResult(BaseModel):
    name: str = Field(..., description="The name of the subdocument")
    pages: list[int] = Field(..., description="The pages of the subdocument (1-indexed)")
    partitions: list[Partition] = Field(default_factory=list, description="The partitions of the subdocument")

class PartitionLikelihood(BaseModel):
    likelihood: float | None = Field(default=None, description="Aggregate confidence for this partition node")
    key: float | None = Field(default=None, description="Confidence that this partition key is correct")
    pages: list[float] = Field(
        default_factory=list,
        description="Confidence for each page in the corresponding partition.pages array",
    )


class SplitSubdocumentLikelihood(BaseModel):
    likelihood: float | None = Field(default=None, description="Aggregate confidence for this split node")
    name: float | None = Field(default=None, description="Confidence that this split label is correct")
    pages: list[float] = Field(
        default_factory=list,
        description="Confidence for each page in the corresponding split.pages array",
    )
    partitions: list[PartitionLikelihood] = Field(
        default_factory=list,
        description="Partition likelihoods aligned with split.partitions",
    )

class SplitConsensus(BaseModel):
    likelihoods: list[SplitSubdocumentLikelihood] | None = Field(
        default=None,
        description="Consensus likelihood tree mirroring the split output",
    )
    choices: list[list[SplitResult]] = Field(
        default_factory=list,
        description="Alternative split vote outputs used to build the consolidated result",
    )


class SplitResponse(BaseModel):
    result: list[SplitResult] = Field(
        ...,
        description="The list of document splits with their assigned pages",
    )
    consensus: SplitConsensus | None = Field(
        default=None,
        description="Consensus metadata for multi-vote split runs",
    )
    usage: RetabUsage = Field(..., description="Usage information for the split operation")
