from pydantic import Field
from retab.types.base import RetabBaseModel

from ..mime import MIMEData
from .usage import RetabUsage


class Subdocument(RetabBaseModel):
    name: str = Field(..., description="The name of the subdocument")
    description: str = Field(default="", description="The description of the subdocument")
    partition_key: str | None = Field(default=None, description="The key to partition the subdocument")
    allow_multiple_instances: bool = Field(
        default=False,
        description="When true, this subdocument type can appear more than once in the document — the split will identify each distinct instance (runs an extra vision-based refinement pass).",
    )


class SplitRequest(RetabBaseModel):
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


class GenerateSplitConfigRequest(RetabBaseModel):
    document: MIMEData = Field(
        ...,
        description="The document to analyze for automatic split configuration generation",
    )
    model: str = Field(default="retab-small", description="The model to use for document analysis")


class GenerateSplitConfigResponse(RetabBaseModel):
    subdocuments: list[Subdocument] = Field(
        ...,
        description="The auto-generated subdocument definitions with optional partition keys",
    )


class Partition(RetabBaseModel):
    key: str = Field(..., description="The partition key value (e.g., property ID, invoice number)")
    pages: list[int] = Field(..., description="The pages of the partition (1-indexed)")


class SplitResult(RetabBaseModel):
    name: str = Field(..., description="The name of the subdocument")
    pages: list[int] = Field(..., description="The pages of the subdocument (1-indexed)")
    partitions: list[Partition] = Field(default_factory=list, description="The partitions of the subdocument")


class SplitChoice(RetabBaseModel):
    splits: list[SplitResult] = Field(default_factory=list, description="One alternative split vote output")


class PartitionLikelihood(RetabBaseModel):
    likelihood: float | None = Field(default=None, description="Aggregate confidence for this partition node")
    key: float | None = Field(default=None, description="Confidence that this partition key is correct")
    pages: list[float] = Field(
        default_factory=list,
        description="Confidence for each page in the corresponding partition.pages array",
    )


class SplitSubdocumentLikelihood(RetabBaseModel):
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


class SplitLikelihoodTree(RetabBaseModel):
    splits: list[SplitSubdocumentLikelihood] = Field(
        default_factory=list,
        description="Likelihood tree aligned with the top-level splits array",
    )


class SplitConsensus(RetabBaseModel):
    likelihoods: SplitLikelihoodTree | None = Field(
        default=None,
        description="Consensus likelihood tree mirroring the split output",
    )
    choices: list[SplitChoice] = Field(
        default_factory=list,
        description="Alternative split vote outputs used to build the consolidated result",
    )


class SplitResponse(RetabBaseModel):
    splits: list[SplitResult] = Field(
        ...,
        description="The list of document splits with their assigned pages",
    )
    consensus: SplitConsensus | None = Field(
        default=None,
        description="Consensus metadata for multi-vote split runs",
    )
    usage: RetabUsage = Field(..., description="Usage information for the split operation")
