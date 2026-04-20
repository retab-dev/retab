from __future__ import annotations

from pydantic import BaseModel, Field

from .documents.usage import RetabUsage
from .mime import MIMEData


class PartitionRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to partition")
    key: str = Field(..., description="The key to partition the document by")
    instructions: str = Field(..., description="Instructions describing how the document should be partitioned")
    model: str = Field(default="retab-small", description="The model to use for partitioning")
    n_consensus: int = Field(
        default=1,
        ge=1,
        le=8,
        description="Number of partitioning runs to use for consensus voting. Uses deterministic single-pass when set to 1.",
    )
    bust_cache: bool = Field(default=False, description="If true, skip the LLM cache and force a fresh completion")


class PartitionChunk(BaseModel):
    key: str = Field(..., description="The partition key value for this chunk")
    pages: list[int] = Field(default_factory=list, description="The pages assigned to this partition chunk (1-indexed)")


class PartitionChunkLikelihood(BaseModel):
    key: float | None = Field(default=None, description="Confidence that this partition key value is correct")
    pages: list[float] = Field(
        default_factory=list,
        description="Confidence for each page in the corresponding partition chunk.pages array",
    )


class PartitionConsensus(BaseModel):
    choices: list[list[PartitionChunk]] = Field(
        default_factory=list,
        description="Alternative partition vote outputs used to build the consolidated result.",
    )
    likelihoods: list[PartitionChunkLikelihood] | None = Field(
        default=None,
        description="Consensus likelihoods aligned with the partition output.",
    )


class PartitionResponse(BaseModel):
    output: list[PartitionChunk] = Field(default_factory=list, description="The partitioned document chunks")
    consensus: PartitionConsensus = Field(
        default_factory=PartitionConsensus,
        description="Consensus metadata for multi-vote partition runs",
    )
    usage: RetabUsage | None = Field(default=None, description="Usage information for the partition operation")
