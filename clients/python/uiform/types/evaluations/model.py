import datetime
import json
from typing import Any, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, Field, computed_field

from ..._utils.json_schema import compute_schema_data_id
from ..._utils.mime import generate_blake2b_hash_from_string
from ..inference_settings import InferenceSettings
from .documents import EvaluationDocument
from .iterations import Iteration


class Evaluation(BaseModel):
    id: str = Field(default_factory=lambda: "eval_" + nanoid.generate())
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))

    name: str
    old_documents: list[EvaluationDocument] | None = None
    documents: list[EvaluationDocument]
    iterations: list[Iteration]
    json_schema: dict[str, Any]

    project_id: str = Field(description="The ID of the project", default="default_spreadsheets")
    default_inference_settings: Optional[InferenceSettings] = Field(default=None, description="The default inference properties for the evaluation (mostly used in the frontend)")

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


class UpdateEvaluationRequest(BaseModel):
    name: Optional[str] = Field(default=None, description="The name of the document")
    documents: Optional[list[EvaluationDocument]] = Field(default=None, description="The documents of the evaluation")
    iterations: Optional[list[Iteration]] = Field(default=None, description="The iterations of the evaluation")
    json_schema: Optional[dict[str, Any]] = Field(default=None, description="The json schema of the evaluation")

    project_id: Optional[str] = Field(default=None, description="The ID of the project")
    default_inference_settings: Optional[InferenceSettings] = Field(default=None, description="The default inference properties for the evaluation (mostly used in the frontend)")

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


class AddIterationFromJsonlRequest(BaseModel):
    jsonl_gcs_path: str
