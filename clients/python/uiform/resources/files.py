from typing import Any, Literal

from openai import AsyncOpenAI, OpenAI

from .._resource import AsyncAPIResource, SyncAPIResource


class Files(SyncAPIResource):
    """Files API wrapper"""

    # Tell Pydantic that _client is not a "field" for serialization/validation,
    # but rather a private attribute to store arbitrary data.

    def create(self, file: Any, purpose: Literal['assistants', 'batch', 'fine-tune', 'vision']) -> Any:
        openai_client = OpenAI(api_key=self._client.headers["OpenAI-Api-Key"])
        return openai_client.files.create(file=file, purpose=purpose)


class AsyncFiles(AsyncAPIResource):
    """Files Asyncronous API wrapper"""

    async def create(self, file: Any, purpose: Literal['assistants', 'batch', 'fine-tune', 'vision']) -> Any:
        async with AsyncOpenAI(api_key=self._client.headers["OpenAI-Api-Key"]) as openai_client:
            return await openai_client.files.create(file=file, purpose=purpose)
