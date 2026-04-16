from typing import Any

from pydantic import BaseModel, Field, model_validator
from ..mime import MIMEData
from .usage import RetabUsage

class Subdocument(BaseModel):
    name: str = Field(..., description="The name of the subdocument")
    description: str = Field(default="", description="The description of the subdocument")
    partition_key: str | None = Field(default=None, description="The key to partition the subdocument")

class SplitRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to split")
    subdocuments: list[Subdocument] = Field(..., min_length=1, description="The subdocuments to split the document into")
    model: str = Field(default="retab-small", description="The model to use to split the document")
    context: str | None = Field(default=None, description="Additional context for the split operation (e.g., iteration context from a loop)")
    n_consensus: int = Field(default=1, ge=1, le=8, description="Number of consensus split runs to perform. Uses deterministic single-pass when set to 1.")
    bust_cache: bool = Field(default=False, description="If true, skip the LLM cache and force a fresh completion")


class Partition(BaseModel):
    key: str = Field(..., description="The partition key value (e.g., property ID, invoice number)")
    pages: list[int] = Field(..., description="The pages of the partition (1-indexed)")

class SplitVote(BaseModel):
    """A single LLM vote from a consensus run."""
    pages: list[int] = Field(..., description="The pages assigned to this subdocument by this vote (1-indexed)")


class SplitResult(BaseModel):
    name: str = Field(..., description="The name of the subdocument")
    pages: list[int] = Field(..., description="The pages of the subdocument (1-indexed)")
    partitions: list[Partition] = Field(default_factory=list, description="The partitions of the subdocument")


class SplitChoice(BaseModel):
    splits: list[SplitResult] = Field(default_factory=list, description="One alternative split vote output")


class PartitionLikelihood(BaseModel):
    key: float | None = Field(default=None, description="Confidence that this partition key is correct")
    pages: list[float] = Field(default_factory=list, description="Confidence for each page in the corresponding partition.pages array")


class SplitLikelihood(BaseModel):
    name: float | None = Field(default=None, description="Confidence that this split label is correct")
    pages: list[float] = Field(default_factory=list, description="Confidence for each page in the corresponding split.pages array")
    partitions: list[PartitionLikelihood] = Field(default_factory=list, description="Partition likelihoods aligned with split.partitions")


class SplitLikelihoodTree(BaseModel):
    splits: list[SplitLikelihood] = Field(default_factory=list, description="Likelihood tree aligned with the top-level splits array")


def _normalize_split_metric_token(value: str) -> str:
    normalized = "".join(
        character.lower() if character.isalnum() or character == "_" else "_"
        for character in value.strip()
    )
    normalized = normalized.strip("_")
    while "__" in normalized:
        normalized = normalized.replace("__", "_")
    return normalized or "split"


def _build_split_metric_key(raw_split: dict[str, Any], split_index: int) -> str:
    raw_metric_key = raw_split.get("metric_key")
    if isinstance(raw_metric_key, str) and raw_metric_key.strip():
        return raw_metric_key.strip()
    raw_name = raw_split.get("name")
    if isinstance(raw_name, str) and raw_name.strip():
        return _normalize_split_metric_token(raw_name)
    return f"split_{split_index}"


def _coerce_float(value: Any) -> float | None:
    if isinstance(value, (int, float)):
        return float(value)
    return None


def _build_split_likelihood_tree_from_flat_map(
    raw_splits: list[dict[str, Any]],
    raw_likelihoods: dict[str, Any],
) -> dict[str, Any]:
    likelihood_splits: list[dict[str, Any]] = []

    for split_index, raw_split in enumerate(raw_splits):
        metric_key = _build_split_metric_key(raw_split, split_index)
        split_path = f"splits.{metric_key}"
        split_likelihood = raw_likelihoods.get(split_path)

        raw_pages = raw_split.get("pages", [])
        pages = [page for page in raw_pages if isinstance(page, int)]

        raw_partitions = raw_split.get("partitions", [])
        partitions: list[dict[str, Any]] = []
        if isinstance(raw_partitions, list):
            for partition_index, raw_partition in enumerate(raw_partitions):
                if not isinstance(raw_partition, dict):
                    continue
                partition_key = str(raw_partition.get("key") or f"partition_{partition_index}")
                partition_path = f"{split_path}.partitions.{partition_key}"
                partition_likelihood = raw_likelihoods.get(partition_path, split_likelihood)
                partition_pages = [
                    page for page in raw_partition.get("pages", []) if isinstance(page, int)
                ]
                partitions.append(
                    {
                        "key": partition_likelihood,
                        "pages": [
                            raw_likelihoods.get(f"{partition_path}.pages.{page}", partition_likelihood)
                            for page in partition_pages
                        ],
                    }
                )

        likelihood_splits.append(
            {
                "name": raw_likelihoods.get(f"{split_path}.name", split_likelihood),
                "pages": [
                    raw_likelihoods.get(f"{split_path}.pages.{page}", split_likelihood)
                    for page in pages
                ],
                "partitions": partitions,
            }
        )

    return {"splits": likelihood_splits}


def _normalize_native_likelihood_tree(raw_likelihoods: dict[str, Any]) -> dict[str, Any]:
    raw_splits = raw_likelihoods.get("splits")
    if not isinstance(raw_splits, list):
        return {"splits": []}

    normalized_splits: list[dict[str, Any]] = []
    for raw_split in raw_splits:
        if not isinstance(raw_split, dict):
            continue
        pages = [float(page) for page in raw_split.get("pages", []) if isinstance(page, (int, float))]
        partitions: list[dict[str, Any]] = []
        raw_partitions = raw_split.get("partitions", [])
        if isinstance(raw_partitions, list):
            for raw_partition in raw_partitions:
                if not isinstance(raw_partition, dict):
                    continue
                partition_pages = [
                    float(page) for page in raw_partition.get("pages", []) if isinstance(page, (int, float))
                ]
                partition_likelihood = _coerce_float(raw_partition.get("likelihood"))
                if partition_likelihood is None:
                    partition_likelihood = _coerce_float(raw_partition.get("key"))
                if partition_likelihood is None and partition_pages:
                    partition_likelihood = sum(partition_pages) / len(partition_pages)
                partitions.append(
                    {
                        "key": _coerce_float(raw_partition.get("key")) or partition_likelihood,
                        "pages": partition_pages,
                    }
                )

        split_likelihood = _coerce_float(raw_split.get("likelihood"))
        if split_likelihood is None:
            split_likelihood = _coerce_float(raw_split.get("name"))
        if split_likelihood is None and pages:
            split_likelihood = sum(pages) / len(pages)

        normalized_splits.append(
            {
                "name": _coerce_float(raw_split.get("name")) or split_likelihood,
                "pages": pages,
                "partitions": partitions,
            }
        )

    return {"splits": normalized_splits}


def _coerce_float(value: Any) -> float | None:
    return float(value) if isinstance(value, (int, float)) else None


class SplitConsensus(BaseModel):
    likelihoods: SplitLikelihoodTree | None = Field(default=None, description="Consensus likelihood tree mirroring the split output")
    choices: list[SplitChoice] = Field(
        default_factory=list,
        description="Alternative split vote outputs used to build the consolidated result",
    )

class SplitResponse(BaseModel):
    splits: list[SplitResult] = Field(..., description="The list of document splits with their assigned pages")
    consensus: SplitConsensus | None = Field(default=None, description="Consensus metadata for multi-vote split runs")
    usage: RetabUsage = Field(..., description="Usage information for the split operation")

    @model_validator(mode="before")
    @classmethod
    def _coerce_native_split_shape(cls, value: Any) -> Any:
        if not isinstance(value, dict):
            return value

        raw_consensus = value.get("consensus")
        if not isinstance(raw_consensus, dict):
            return value

        raw_likelihoods = raw_consensus.get("likelihoods")
        if not isinstance(raw_likelihoods, dict):
            return value
        if "splits" in raw_likelihoods:
            normalized = dict(value)
            normalized["consensus"] = {
                **raw_consensus,
                "likelihoods": _normalize_native_likelihood_tree(raw_likelihoods),
            }
            return normalized

        raw_splits = value.get("splits", [])
        normalized_splits = [
            raw_split
            for raw_split in raw_splits
            if isinstance(raw_split, dict)
        ]
        normalized = dict(value)
        normalized["consensus"] = {
            **raw_consensus,
            "likelihoods": _build_split_likelihood_tree_from_flat_map(
                normalized_splits,
                raw_likelihoods,
            ),
        }
        return normalized

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
