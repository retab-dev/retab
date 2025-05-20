import os
from typing import List

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.ai_models import AIProvider
from ...types.secrets.external_api_keys import ExternalAPIKey, ExternalAPIKeyRequest
from ...types.standards import PreparedRequest


class ExternalAPIKeysMixin:
    def prepare_create(self, provider: AIProvider, api_key: str) -> PreparedRequest:
        data = {"provider": provider, "api_key": api_key}
        request = ExternalAPIKeyRequest.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/secrets/external_api_keys", data=request.model_dump(mode="json"))

    def prepare_get(self, provider: AIProvider) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/secrets/external_api_keys/{provider}")

    def prepare_list(self) -> PreparedRequest:
        return PreparedRequest(method="GET", url="/v1/secrets/external_api_keys")

    def prepare_delete(self, provider: AIProvider) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/secrets/external_api_keys/{provider}")


class ExternalAPIKeys(SyncAPIResource, ExternalAPIKeysMixin):
    """External API Keys management wrapper"""

    def create(self, provider: AIProvider, api_key: str) -> dict:
        """Add or update an external API key.

        Args:
            provider: The API provider (openai, gemini, anthropic, xai)
            api_key: The API key to store

        Returns:
            dict: Response indicating success
        """

        request = self.prepare_create(provider, api_key)
        response = self._client._prepared_request(request)
        return response

    def get(
        self,
        provider: AIProvider,
    ) -> ExternalAPIKey:
        """Get an external API key configuration.

        Args:
            provider: The API provider to get the key for

        Returns:
            ExternalAPIKey: The API key configuration
        """
        request = self.prepare_get(provider)
        response = self._client._prepared_request(request)
        return response

        return ExternalAPIKey.model_validate(response)

    def list(self) -> List[ExternalAPIKey]:
        """List all configured external API keys.

        Returns:
            List[ExternalAPIKey]: List of API key configurations
        """
        request = self.prepare_list()
        response = self._client._prepared_request(request)

        return [ExternalAPIKey.model_validate(key) for key in response]

    def delete(self, provider: AIProvider) -> dict:
        """Delete an external API key configuration.

        Args:
            provider: The API provider to delete the key for

        Returns:
            dict: Response indicating success
        """
        request = self.prepare_delete(provider)
        response = self._client._prepared_request(request)

        return response


class AsyncExternalAPIKeys(AsyncAPIResource, ExternalAPIKeysMixin):
    """External API Keys management wrapper"""

    async def create(self, provider: AIProvider, api_key: str) -> dict:
        request = self.prepare_create(provider, api_key)
        response = await self._client._prepared_request(request)
        return response

    async def get(self, provider: AIProvider) -> ExternalAPIKey:
        request = self.prepare_get(provider)
        response = await self._client._prepared_request(request)
        return response

    async def list(self) -> List[ExternalAPIKey]:
        request = self.prepare_list()
        response = await self._client._prepared_request(request)
        return [ExternalAPIKey.model_validate(key) for key in response]

    async def delete(self, provider: AIProvider) -> dict:
        request = self.prepare_delete(provider)
        response = await self._client._prepared_request(request)
        return response
