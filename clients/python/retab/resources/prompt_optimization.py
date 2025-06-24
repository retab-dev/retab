# from typing import Any
# from pathlib import Path
# import json

# from .._resource import SyncAPIResource, AsyncAPIResource
# from .._utils.json_schema import load_json_schema
# from ..types.jobs import JobResponse
# from ..types.jobs.prompt_optimization import PromptOptimizationObject, PromptOptimizationProps, PromptOptimizationJobInputData, PromptOptimizationJob

# MAX_TRAINING_SAMPLES = 10

# class PromptOptimizationJobsMixin:
#     def prepare_create(self, json_schema: dict[str, Any] | Path | str,
#                        training_file: str,
#                        schema_optimization_props: dict[str, Any]) -> PromptOptimizationJob:

#         optimization_objects = []
#         with open(training_file, "r") as f:
#             for line in f:
#                 optimization_objects.append(PromptOptimizationObject(**json.loads(line)))


#         job = PromptOptimizationJob(
#             job_type="prompt-optimization",
#             input_data=PromptOptimizationJobInputData(
#                 json_schema=load_json_schema(json_schema),
#                 optimization_objects=optimization_objects[:MAX_TRAINING_SAMPLES],
#                 schema_optimization_props=PromptOptimizationProps.model_validate(schema_optimization_props)
#             )
#         )
#         return job

# class PromptOptimizationJobs(SyncAPIResource, PromptOptimizationJobsMixin):
#     def create(self, json_schema: dict[str, Any] | Path | str, training_file: str, schema_optimization_props: dict[str, Any]) -> JobResponse:
#         """Create a new prompt optimization job"""

#         request_data = self.prepare_create(json_schema, training_file, schema_optimization_props)
#         response = self._client._request("POST", "/v1/jobs", data=request_data.model_dump(mode="json"))
#         return JobResponse.model_validate(response)


#     def retrieve(self, job_id: str) -> Any:
#         """Retrieve status of a prompt optimization job"""
#         response = self._client._request("GET", f"/v1/jobs/{job_id}")
#         return JobResponse.model_validate(response)

# class AsyncPromptOptimizationJobs(AsyncAPIResource, PromptOptimizationJobsMixin):
#     async def create(self, json_schema: dict[str, Any] | Path | str, training_file: str, schema_optimization_props: dict[str, Any]) -> Any:
#         """Create a new prompt optimization job"""

#         request_data = self.prepare_create(json_schema, training_file, schema_optimization_props)
#         response = await self._client._request("POST", "/v1/jobs/", data=request_data.model_dump(mode="json"))
#         return JobResponse.model_validate(response)

#     async def retrieve(self, job_id: str) -> Any:
#         """Retrieve status of a prompt optimization job"""

#         response = await self._client._request("GET", f"/v1/jobs/{job_id}")
#         return JobResponse.model_validate(response)


# class PromptOptimization(SyncAPIResource):
#     """Prompt optimization jobs API wrapper"""
#     _jobs: PromptOptimizationJobs

#     def __init__(self, *args: Any, **kwargs: Any):
#         super().__init__(*args, **kwargs)
#         self._jobs = PromptOptimizationJobs(client=self._client)

# class AsyncPromptOptimization(AsyncAPIResource):
#     """Prompt optimization jobs Asyncronous API wrapper"""
#     _jobs: AsyncPromptOptimizationJobs

#     def __init__(self, *args: Any, **kwargs: Any):
#         super().__init__(*args, **kwargs)
#         self._jobs = AsyncPromptOptimizationJobs(client=self._client)
