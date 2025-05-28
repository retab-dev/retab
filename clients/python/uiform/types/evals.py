import copy
import datetime
import json
from typing import Any, List, Literal, Optional, Union

import nanoid  # type: ignore
from pydantic import BaseModel, Field, computed_field


from .._utils.json_schema import clean_schema, compute_schema_data_id
from .._utils.mime import generate_blake2b_hash_from_string
from .ai_models import Amount, LLMModel
from .jobs.base import InferenceSettings
from .mime import MIMEData


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


class AnnotatedDocument(BaseModel):
    mime_data: MIMEData = Field(
        description="The mime data of the document. Can also be a BaseMIMEData, which is why we have this id field (to be able to identify the file, but id is equal to mime_data.id)"
    )
    annotation: dict[str, Any] = Field(default={}, description="The ground truth of the document")


class DocumentItem(AnnotatedDocument):
    annotation_metadata: Optional[PredictionMetadata] = Field(default=None, description="The metadata of the annotation when the annotation is a prediction")


class EvaluationDocument(DocumentItem):
    id: str = Field(description="The ID of the document. Equal to mime_data.id but robust to the case where mime_data is a BaseMIMEData")


class CreateIterationRequest(BaseModel):
    """
    Request model for performing a new iteration with custom inference settings and optional JSON schema.
    """

    inference_settings: InferenceSettings
    json_schema: Optional[dict[str, Any]] = None


class UpdateEvaluationDocumentRequest(BaseModel):
    annotation: Optional[dict[str, Any]] = Field(default=None, description="The ground truth of the document")
    annotation_metadata: Optional[PredictionMetadata] = Field(default=None, description="The metadata of the annotation when the annotation is a prediction")


class UpdateEvaluationRequest(BaseModel):
    name: Optional[str] = Field(default=None, description="The name of the document")
    documents: Optional[list[EvaluationDocument]] = Field(default=None, description="The documents of the evaluation")
    iterations: Optional[list[Iteration]] = Field(default=None, description="The iterations of the evaluation")
    json_schema: Optional[dict[str, Any]] = Field(default=None, description="The json schema of the evaluation")

    project_id: Optional[str] = Field(default=None, description="The ID of the project")

    @computed_field  # type: ignore
    @property
    def schema_data_id(self) -> Optional[str]:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.

        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        if self.json_schema is None:
            return None

        return compute_schema_data_id(self.json_schema)

    # This is a computed field, it is exposed when serializing the object
    @computed_field  # type: ignore
    @property
    def schema_id(self) -> Optional[str]:
        """Returns the SHA1 hash of the complete schema.

        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        if self.json_schema is None:
            return None
        return "sch_id_" + generate_blake2b_hash_from_string(json.dumps(self.json_schema, sort_keys=True).strip())


class Evaluation(BaseModel):
    id: str = Field(default_factory=lambda: "eval_" + nanoid.generate())
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))

    name: str
    documents: list[EvaluationDocument]
    iterations: list[Iteration]
    json_schema: dict[str, Any]

    project_id: str = Field(description="The ID of the project", default="default_spreadsheets")
    default_inference_settings: Optional[InferenceSettings] = Field(default=None, description="The default inference properties for the evaluation (mostly used in the frontend)")

    # @field_validator('iterations')
    # def validate_iterations_content_length(cls: Any, v: list[Iteration], values: Any) -> list[Iteration]:
    #     if 'ground_truth' in values:
    #         ground_truth_length = len(values['ground_truth'])
    #         for iteration in v:
    #             if len(iteration.content) != ground_truth_length:
    #                 raise ValueError(f"Iteration content length must match ground_truth length ({ground_truth_length})")
    #     return v

    @computed_field  # type: ignore
    @property
    def schema_data_id(self) -> str:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.

        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        return compute_schema_data_id(self.json_schema)

    # This is a computed field, it is exposed when serializing the object
    @computed_field  # type: ignore
    @property
    def schema_id(self) -> str:
        """Returns the SHA1 hash of the complete schema.

        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        return "sch_id_" + generate_blake2b_hash_from_string(json.dumps(self.json_schema, sort_keys=True).strip())


class AddIterationFromJsonlRequest(BaseModel):
    jsonl_gcs_path: str
