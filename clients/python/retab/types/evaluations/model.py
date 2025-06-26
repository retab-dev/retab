import datetime
import json
from typing import Any, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, Field, computed_field

from ...utils.json_schema import compute_schema_data_id
from ...utils.mime import generate_blake2b_hash_from_string
from ..inference_settings import InferenceSettings
from .documents import EvaluationDocument
from .iterations import Iteration


# Actual Object stored in DB
class Evaluation(BaseModel):
    id: str = Field(default_factory=lambda: "eval_" + nanoid.generate())
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))

    name: str
    documents: list[EvaluationDocument] = Field(default_factory=list)
    iterations: list[Iteration] = Field(default_factory=list)
    json_schema: dict[str, Any]

    project_id: str = Field(description="The ID of the project", default="default_spreadsheets")
    default_inference_settings: InferenceSettings = Field(
        default=InferenceSettings(), description="The default inference properties for the evaluation (mostly used in the frontend)"
    )

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


class CreateEvaluation(BaseModel):
    name: str
    json_schema: dict[str, Any]
    project_id: str = Field(description="The ID of the project", default="default_spreadsheets")
    default_inference_settings: InferenceSettings = Field(default=InferenceSettings(), description="The default inference properties for the evaluation.")


class ListEvaluationParams(BaseModel):
    project_id: Optional[str] = Field(default=None, description="The ID of the project")
    schema_id: Optional[str] = Field(default=None, description="The ID of the schema")
    schema_data_id: Optional[str] = Field(default=None, description="The ID of the schema data")


class PatchEvaluationRequest(BaseModel):
    name: Optional[str] = Field(default=None, description="The name of the document")
    json_schema: Optional[dict[str, Any]] = Field(default=None, description="The json schema of the evaluation")
    project_id: Optional[str] = Field(default=None, description="The ID of the project")
    default_inference_settings: Optional[InferenceSettings] = Field(default=None, description="The default inference properties for the evaluation (mostly used in the frontend)")


class AddIterationFromJsonlRequest(BaseModel):
    jsonl_gcs_path: str
