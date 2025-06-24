# from typing import Literal, Any
# from pydantic import BaseModel, computed_field
# from ..mime import MIMEData
# from ..._utils.benchmarking import ExtractionAnalysis

# MAX_CONCURRENCY = 15


# class PromptOptimizationObject(BaseModel):
#     mime_document: MIMEData
#     target: dict
#     extracted: dict | None = None

#     @computed_field     # type: ignore
#     @property
#     def analysis(self) -> ExtractionAnalysis | None:
#         if self.extracted is None:
#             return None
#         return ExtractionAnalysis(ground_truth=self.target, prediction=self.extracted)


# Metrics = Literal["levenshtein_similarity_per_field", "accuracy_per_field"]


# # Insert default values for the parameters
# class PromptOptimizationProps(BaseModel):
#     start_hierarchy_level: int = 0
#     threshold: float = 0.9
#     model: str = "gpt-4o-mini"
#     iterations_per_level: int = 1
#     metric: Metrics = "accuracy_per_field"

# class PromptOptimizationJobInputData(BaseModel):
#     json_schema: dict[str, Any]
#     optimization_objects: list[PromptOptimizationObject]
#     schema_optimization_props: PromptOptimizationProps


# class PromptOptimizationJob(BaseModel):
#     job_type: Literal["prompt-optimization"] = "prompt-optimization"
#     input_data: PromptOptimizationJobInputData
