from typing import List

from ..._resource import SyncAPIResource
from ...types.secrets.external_api_keys import  ExternalAPIKeyRequest, ExternalAPIKey
from ...types.ai_model import AIProvider

import os
    

class ExternalAPIKeys(SyncAPIResource):
    """External API Keys management wrapper"""

    def create(
        self,
        provider: AIProvider,
        api_key: str
    ) -> dict:
        """Add or update an external API key.

        Args:
            provider: The API provider (openai, gemini, anthropic, xai)
            api_key: The API key to store

        Returns:
            dict: Response indicating success
        """
        data = {
            "provider": provider,
            "api_key": api_key
        }

        request = ExternalAPIKeyRequest.model_validate(data)
        response = self._client._request(
            "POST",
            "/v1/secrets",
            data=request.model_dump(mode="json")
        )

        return response

    def update(
        self,
        provider: AIProvider,
        api_key: str
    ) -> dict:
        """Add or update an external API key.

        Args:
            provider: The API provider (openai, gemini, anthropic, xai)
            api_key: The API key to store

        Returns:
            dict: Response indicating success
        """
        data = {
            "provider": provider,
            "api_key": api_key
        }

        request = ExternalAPIKeyRequest.model_validate(data)
        response = self._client._request(
            "PUT",
            "/v1/secrets",
            data=request.model_dump(mode="json")
        )

        return response

    def get(self,
        provider: AIProvider,
    ) -> ExternalAPIKey:
        """Get an external API key configuration.

        Args:
            provider: The API provider to get the key for

        Returns:
            ExternalAPIKey: The API key configuration
        """
        response = self._client._request(
            "GET",
            f"/v1/secrets/{provider}"
        )

        return ExternalAPIKey.model_validate(response)
    
    def list(self) -> List[ExternalAPIKey]:
        """List all configured external API keys.

        Returns:
            List[ExternalAPIKey]: List of API key configurations
        """
        response = self._client._request(
            "GET",
            f"/v1/secrets"
        )

        return [ExternalAPIKey.model_validate(key) for key in response]



    def delete(self, provider: AIProvider) -> dict:
        """Delete an external API key configuration.

        Args:
            provider: The API provider to delete the key for

        Returns:
            dict: Response indicating success
        """
        response = self._client._request(
            "DELETE",
            f"/v1/secrets/{provider}"
        )

        return response
