from typing import Any, Literal, Optional
from pydantic import BaseModel, Field

# Define the type alias for MetricType
MetricType = Literal["levenshtein", "jaccard", "hamming"]


# Define the structure for an individual item metric
class ItemMetric(BaseModel):
    id: str = Field(description="The ID of the item being measured")
    name: str = Field(description="The name of the item being measured")
    similarity: float = Field(description="The similarity score between 0 and 1")
    similarities: dict[str, Any] = Field(description="The similarity scores for each item in the list")
    flat_similarities: dict[str, Optional[float]] = Field(description="The similarity scores for each item in the list in dot notation format")
    aligned_similarity: float = Field(description="The similarity score between 0 and 1, after alignment")
    aligned_similarities: dict[str, Any] = Field(description="The similarity scores for each item in the list, after alignment")
    aligned_flat_similarities: dict[str, Optional[float]] = Field(description="The similarity scores for each item in the list in dot notation format, after alignment")


# Define the main MetricResult model
class MetricResult(BaseModel):
    item_metrics: list[ItemMetric] = Field(description="List of similarity metrics for individual items")
    mean_similarity: float = Field(description="The average similarity score across all items")
    aligned_mean_similarity: float = Field(description="The average similarity score across all items, after alignment")
    metric_type: MetricType = Field(description="The type of similarity metric used for comparison")


class DistancesResult(BaseModel):
    distances: dict[str, Any] = Field(description="List of distances for individual items")
    mean_distance: float = Field(description="The average distance across all items")
    metric_type: MetricType = Field(description="The type of distance metric used for comparison")
