from openai.types.model import Model

from .._resource import AsyncAPIResource, SyncAPIResource
from ..types.standards import PreparedRequest


class ModelsMixin:
    def prepare_list(
        self,
        supports_finetuning: bool = False,
        supports_image: bool = False,
        include_finetuned_models: bool = True,
    ) -> PreparedRequest:
        params = {
            "supports_finetuning": supports_finetuning,
            "supports_image": supports_image,
            "include_finetuned_models": include_finetuned_models,
        }
        return PreparedRequest(method="GET", url="/v1/models", params=params)


class Models(SyncAPIResource, ModelsMixin):
    """Models API wrapper"""

    def list(
        self,
        supports_finetuning: bool = False,
        supports_image: bool = False,
        include_finetuned_models: bool = True,
    ) -> list[Model]:
        """
        List all available models.

        Args:
            supports_finetuning: Filter for models that support fine-tuning
            supports_image: Filter for models that support image inputs
            include_finetuned_models: Include fine-tuned models in results

        Returns:
            list[Model]: List of available models
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(supports_finetuning, supports_image, include_finetuned_models)
        output = self._client._prepared_request(request)
        return [Model(**model) for model in output["data"]]


class AsyncModels(AsyncAPIResource, ModelsMixin):
    """Models Asyncronous API wrapper"""

    async def list(
        self,
        supports_finetuning: bool = False,
        supports_image: bool = False,
        include_finetuned_models: bool = True,
    ) -> list[Model]:
        """
        List all available models asynchronously.

        Args:
            supports_finetuning: Filter for models that support fine-tuning
            supports_image: Filter for models that support image inputs
            include_finetuned_models: Include fine-tuned models in results

        Returns:
            list[Model]: List of available models
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(supports_finetuning, supports_image, include_finetuned_models)
        output = await self._client._prepared_request(request)
        return [Model(**model) for model in output["data"]]
