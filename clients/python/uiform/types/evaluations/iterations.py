import datetime
import copy
import json
from typing import Any, Optional, Self

import nanoid  # type: ignore
from pydantic import BaseModel, Field, computed_field, model_validator

from ..._utils.json_schema import clean_schema
from ..._utils.mime import generate_blake2b_hash_from_string
from ..inference_settings import InferenceSettings
from ..predictions import PredictionData
from ..metrics import MetricResult


class Iteration(BaseModel):
    id: str = Field(default_factory=lambda: "eval_iter_" + nanoid.generate())
    updated_at: datetime.datetime = Field(
        default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc),
        description="The last update date of inference settings or json schema",
    )
    inference_settings: InferenceSettings
    json_schema: dict[str, Any]
    predictions: dict[str, PredictionData] = Field(default_factory=dict, description="The predictions of the iteration for all the documents")
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
    from_iteration_id: Optional[str] = Field(
        default=None,
        description="The ID of the iteration to copy the JSON Schema from.",
    )

    # validate that exactly one of from_iteration_id or json_schema is provided
    @model_validator(mode="after")
    def validate_one_of_from_iteration_id_or_json_schema(self) -> Self:
        if (self.from_iteration_id is None) ^ (self.json_schema is None):
            raise ValueError("Exactly one of from_iteration_id or json_schema must be provided")
        return self


class PatchIterationRequest(BaseModel):
    inference_settings: Optional[InferenceSettings] = Field(default=None, description="The new inference settings of the iteration")
    json_schema: Optional[dict[str, Any]] = Field(default=None, description="The new json schema of the iteration")


class AddIterationFromJsonlRequest(BaseModel):
    jsonl_gcs_path: str
