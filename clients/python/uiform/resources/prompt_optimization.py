from typing import Any

from .._resource import SyncAPIResource, AsyncAPIResource


class PromptOptimizationJobs(SyncAPIResource):
    def create(self, training_file: str, model: str) -> Any:
        """Create a new prompt optimization job"""

        # TODO

        raise NotImplementedError("Prompt optimization is not implemented yet")

    def retrieve(self, job_id: str) -> Any:
        """Retrieve status of a prompt optimization job"""

        # TODO

        raise NotImplementedError("Prompt optimization is not implemented yet")

class AsyncOptimizationJobs(AsyncAPIResource):
    async def create(self, training_file: str, model: str) -> Any:
        """Create a new prompt optimization job"""

        # TODO

        raise NotImplementedError("Prompt optimization is not implemented yet")

    async def retrieve(self, job_id: str) -> Any:
        """Retrieve status of a prompt optimization job"""

        # TODO

        raise NotImplementedError("Prompt optimization is not implemented yet")

class PromptOptimization(SyncAPIResource):
    """Prompt optimization jobs API wrapper"""
    _jobs: PromptOptimizationJobs

    def __init__(self, *args: Any, **kwargs: Any):
        super().__init__(*args, **kwargs)
        self._jobs = PromptOptimizationJobs(client=self._client)

class AsyncPromptOptimization(AsyncAPIResource):
    """Prompt optimization jobs Asyncronous API wrapper"""
    _jobs: AsyncOptimizationJobs

    def __init__(self, *args: Any, **kwargs: Any):
        super().__init__(*args, **kwargs)
        self._jobs = AsyncOptimizationJobs(client=self._client)
   