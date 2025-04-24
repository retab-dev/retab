from typing import Any, get_args

from openai import AsyncOpenAI, OpenAI

from .._resource import AsyncAPIResource, SyncAPIResource
from ..types.ai_models import OpenAIModel


class FineTuningJobs(SyncAPIResource):
    def create(self, training_file: str, model: str) -> Any:
        """Create a new fine-tuning job"""
        openai_client = OpenAI(api_key=self._client.headers["OpenAI-Api-Key"])
        assert model in get_args(OpenAIModel), f"Model {model} is not supported. Supported models are: {get_args(OpenAIModel)}"
        return openai_client.fine_tuning.jobs.create(training_file=training_file, model=model)

    def retrieve(self, fine_tuning_job_id: str) -> Any:
        """Retrieve status of a fine-tuning job"""
        openai_client = OpenAI(api_key=self._client.headers["OpenAI-Api-Key"])
        return openai_client.fine_tuning.jobs.retrieve(fine_tuning_job_id)


class AsyncFineTuningJobs(AsyncAPIResource):
    async def create(self, training_file: str, model: str) -> Any:
        """Create a new fine-tuning job"""
        assert model in get_args(OpenAIModel), f"Model {model} is not supported. Supported models are: {get_args(OpenAIModel)}"
        async with AsyncOpenAI(api_key=self._client.headers["OpenAI-Api-Key"]) as openai_client:
            return await openai_client.fine_tuning.jobs.create(training_file=training_file, model=model)

    async def retrieve(self, fine_tuning_job_id: str) -> Any:
        """Retrieve status of a fine-tuning job"""
        async with AsyncOpenAI(api_key=self._client.headers["OpenAI-Api-Key"]) as openai_client:
            return await openai_client.fine_tuning.jobs.retrieve(fine_tuning_job_id)


class FineTuning(SyncAPIResource):
    """Fine-tuning jobs API wrapper"""

    _jobs: FineTuningJobs

    def __init__(self, *args: Any, **kwargs: Any):
        super().__init__(*args, **kwargs)
        self._jobs = FineTuningJobs(self._client)  # Initialize the Jobs instance with the client

    @property
    def jobs(self) -> FineTuningJobs:
        """Expose your private _jobs attribute through a property."""
        return self._jobs


class AsyncFineTuning(AsyncAPIResource):
    """Fine-tuning jobs Asyncronous API wrapper"""

    _jobs: AsyncFineTuningJobs

    def __init__(self, *args: Any, **kwargs: Any):
        super().__init__(*args, **kwargs)
        self._jobs = AsyncFineTuningJobs(self._client)  # Initialize the Jobs instance with the client

    @property
    def jobs(self) -> AsyncFineTuningJobs:
        """Expose your private _jobs attribute through a property."""
        return self._jobs
