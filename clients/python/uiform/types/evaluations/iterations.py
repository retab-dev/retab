import copy
import json
from typing import Any, List, Literal, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, Field, computed_field

from ..._utils.json_schema import clean_schema
from ..._utils.mime import generate_blake2b_hash_from_string
from ..ai_models import Amount
from ..jobs.base import InferenceSettings

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
    item_metrics: List[ItemMetric] = Field(description="List of similarity metrics for individual items")
    mean_similarity: float = Field(description="The average similarity score across all items")
    aligned_mean_similarity: float = Field(description="The average similarity score across all items, after alignment")
    metric_type: MetricType = Field(description="The type of similarity metric used for comparison")


class DistancesResult(BaseModel):
    distances: dict[str, Any] = Field(description="List of distances for individual items")
    mean_distance: float = Field(description="The average distance across all items")
    metric_type: MetricType = Field(description="The type of distance metric used for comparison")


class PredictionMetadata(BaseModel):
    extraction_id: Optional[str] = Field(default=None, description="The ID of the extraction")
    likelihoods: Optional[dict[str, Any]] = Field(default=None, description="The likelihoods of the extraction")
    field_locations: Optional[dict[str, Any]] = Field(default=None, description="The field locations of the extraction")
    agentic_field_locations: Optional[dict[str, Any]] = Field(default=None, description="The field locations of the extraction extracted by an llm")
    consensus_details: Optional[list[dict[str, Any]]] = Field(default=None, description="The consensus details of the extraction")
    api_cost: Optional[Amount] = Field(default=None, description="The cost of the API call for this document (if any -- ground truth for example)")


class PredictionData(BaseModel):
    prediction: dict[str, Any] = Field(default={}, description="The result of the extraction or manual annotation")
    metadata: Optional[PredictionMetadata] = Field(default=None, description="The metadata of the prediction")


class Iteration(BaseModel):
    id: str = Field(default_factory=lambda: "eval_iter_" + nanoid.generate())
    inference_settings: InferenceSettings
    json_schema: dict[str, Any]
    predictions: list[PredictionData] = Field(default_factory=list, description="The predictions of the iteration for all the documents")
    metric_results: Optional[MetricResult] = Field(default=None, description="The metric results of the iteration")

    @computed_field  # type: ignore
    @property
    def schema_data_id(self) -> str:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.

        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        return "sch_data_id_" + generate_blake2b_hash_from_string(
            json.dumps(
                clean_schema(
                    copy.deepcopy(self.json_schema),
                    remove_custom_fields=True,
                    fields_to_remove=["description", "default", "title", "required", "examples", "deprecated", "readOnly", "writeOnly"],
                ),
                sort_keys=True,
            ).strip()
        )

    # This is a computed field, it is exposed when serializing the object
    @computed_field  # type: ignore
    @property
    def schema_id(self) -> str:
        """Returns the SHA1 hash of the complete schema.

        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        return "sch_id_" + generate_blake2b_hash_from_string(json.dumps(self.json_schema, sort_keys=True).strip())


class CreateIterationRequest(BaseModel):
    """
    Request model for performing a new iteration with custom inference settings and optional JSON schema.
    """

    inference_settings: InferenceSettings
    json_schema: Optional[dict[str, Any]] = None


class AddIterationFromJsonlRequest(BaseModel):
    jsonl_gcs_path: str
