from typing import Literal, TypedDict, Any
from pydantic import BaseModel, computed_field, field_validator
from .mime import MIMEData
from .schemas.object import Schema
from ..resources.benchmarking import ExtractionAnalysis
from fastapi.encoders import jsonable_encoder
MAX_CONCURRENCY = 15


class PromptOptimizationObject(BaseModel):
    mime_document: MIMEData
    target: dict
    extracted: dict | None = None

    @computed_field     # type: ignore
    @property
    def analysis(self) -> ExtractionAnalysis | None:
        if self.extracted is None:
            return None
        return ExtractionAnalysis(ground_truth=self.target, prediction=self.extracted)
    


Metrics = Literal["levenshtein_similarity_per_field", "accuracy_per_field"]

class PromptOptimizationPropsParams(TypedDict, total=True):
    start_hierarchy_level: int
    threshold: float
    model: str
    iterations_per_level: int
    metric: Metrics

# Insert default values for the parameters
class PromptOptimizationProps(BaseModel):
    start_hierarchy_level: int = 0
    threshold: float = 0.9
    model: str = "gpt-4o-mini"
    iterations_per_level: int = 1
    metric: Metrics = "accuracy_per_field"



class PromptOptimizationJobInputData(BaseModel):
    raw_schema: dict[str, Any]
    optimization_objects: list[PromptOptimizationObject]
    schema_optimization_props: PromptOptimizationProps


class PromptOptimizationJob(BaseModel):
    job_type: Literal["prompt-optimization"]
    input_data: PromptOptimizationJobInputData
