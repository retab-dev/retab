from typing import Literal, Self

from pydantic import BaseModel, model_validator

from ..jobs.base import AnnotationInputData, EvaluationInputData, InferenceSettings, PrepareDatasetInputData

Workflows = Literal["finetuning-workflow", "annotation-workflow", "evaluation-workflow"]

AnnotationModel = Literal["human"] | str  # If human, then inference_settings is not used


# This is the input data for the standalone annotation workflow (Fully automated)
class StandaloneAnnotationWorkflowInputData(AnnotationInputData):
    pass


# This is the input data for the standalone evaluation workflow (Fully automated)
class StandaloneEvaluationWorkflowInputData(EvaluationInputData):
    pass


# This is the input data for the standalone finetuning workflow (with human in the loop)
class FinetuningWorkflowInputData(BaseModel):
    prepare_dataset_input_data: PrepareDatasetInputData
    annotation_model: AnnotationModel
    inference_settings: InferenceSettings | None = None
    finetuning_props: InferenceSettings

    # Validate the input
    @model_validator(mode="after")
    def validate_input(self) -> Self:
        if self.annotation_model == "human":
            if self.inference_settings is not None:
                raise ValueError("inference_settings must be None if annotation_model is human")
        else:
            self.inference_settings = self.finetuning_props
        return self
