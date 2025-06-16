from typing import Literal, Optional, Self

from pydantic import BaseModel, model_validator
from ..inference_settings import InferenceSettings

SelectionMode = Literal["all", "manual"]


# This is the input data for the prepare_dataset job
class PrepareDatasetInputData(BaseModel):
    dataset_id: Optional[str] = None
    schema_id: Optional[str] = None
    schema_data_id: Optional[str] = None

    selection_model: SelectionMode = "all"

    @model_validator(mode="after")
    def validate_input(self) -> Self:
        # The preference is:
        # 1. dataset_id
        # 2. schema_id
        # 3. schema_data_id
        if self.dataset_id is None and self.schema_id is None and self.schema_data_id is None:
            raise ValueError("At least one of dataset_id, schema_id, or schema_data_id must be provided")

        return self


# This is the input data for the split_dataset job
class DatasetSplitInputData(BaseModel):
    dataset_id: str
    train_size: Optional[int | float] = None
    eval_size: Optional[int | float] = None

    @model_validator(mode="after")
    def validate_input(self) -> Self:
        if self.train_size is not None and self.eval_size is not None:
            raise ValueError("train_size and eval_size cannot both be provided")
        return self


# This is the input data for the batch annotation job
class AnnotationInputData(BaseModel):
    data_file: str
    schema_id: str
    inference_settings: InferenceSettings


# This is the input data for the evaluation job
class EvaluationInputData(BaseModel):
    eval_data_file: str
    schema_id: str
    inference_settings_1: InferenceSettings | None = None
    inference_settings_2: InferenceSettings
