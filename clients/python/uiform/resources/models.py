from openai.types.model import Model

from .._resource import AsyncAPIResource, SyncAPIResource
from ..types.standards import PreparedRequest


class ModelsMixin:
    def prepare_list(self) -> PreparedRequest:
        return PreparedRequest(method="GET", url="/v1/models")


class Models(SyncAPIResource, ModelsMixin):
    """Models API wrapper"""

    def list(self) -> list[Model]:
        """
        List all available models.

        Returns:
            list[str]: List of available models
        Raises:
            HTTPException if the request fails
        """

        request = self.prepare_list()
        output = self._client._prepared_request(request)
        return [Model(**model) for model in output["data"]]


class AsyncModels(AsyncAPIResource, ModelsMixin):
    """Models Asyncronous API wrapper"""

    async def list(self) -> list[Model]:
        """
        List all available models.

        Returns:
            list[str]: List of available models
        Raises:
            HTTPException if the request fails
        """

        request = self.prepare_list()
        output = await self._client._prepared_request(request)
        return [Model(**model) for model in output["data"]]
