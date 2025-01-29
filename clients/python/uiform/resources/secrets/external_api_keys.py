from typing import List

from ..._resource import SyncAPIResource
from ...types.secrets.secrets import ApiProvider, ExternalAPIKey, ExternalAPIKeyResponse

class ExternalAPIKeys(SyncAPIResource):
    """External API Keys management wrapper"""

    def create(
        self,
        provider: ApiProvider,
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

        request = ExternalAPIKey.model_validate(data)
        response = self._client._request(
            "POST",
            "/api/v1/iam/external_api_keys/",
            data=request.model_dump(mode="json")
        )

        return response

    def update(
        self,
        provider: ApiProvider,
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

        request = ExternalAPIKey.model_validate(data)
        response = self._client._request(
            "PUT",
            "/api/v1/iam/external_api_keys/",
            data=request.model_dump(mode="json")
        )

        return response

    def get(self,
        provider: ApiProvider,
        display_key:bool = False ) -> ExternalAPIKeyResponse:
        """Get an external API key configuration.

        Args:
            provider: The API provider to get the key for

        Returns:
            ExternalAPIKeyResponse: The API key configuration
        """
        response = self._client._request(
            "GET",
            f"/api/v1/iam/external_api_keys/{provider}?display_key={display_key}"
        )

        return ExternalAPIKeyResponse.model_validate(response)
    def list(self, display_keys:bool = False) -> List[ExternalAPIKeyResponse]:
        """List all configured external API keys.

        Returns:
            List[ExternalAPIKeyResponse]: List of API key configurations
        """
        response = self._client._request(
            "GET",
            f"/api/v1/iam/external_api_keys/list?display_keys={display_keys}"
        )

        return [ExternalAPIKeyResponse.model_validate(key) for key in response]



    def delete(self, provider: ApiProvider) -> dict:
        """Delete an external API key configuration.

        Args:
            provider: The API provider to delete the key for

        Returns:
            dict: Response indicating success
        """
        response = self._client._request(
            "DELETE",
            f"/api/v1/iam/external_api_keys/{provider}"
        )

        return response
