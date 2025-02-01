from typing import List

from ..._resource import SyncAPIResource
from ...types.secrets.secrets import  ExternalAPIKeyRequest, ExternalAPIKey
from ...types.ai_model import AIProvider

import os
    

class Webhook(SyncAPIResource):
    """Webhook secret management wrapper"""

    def create(
        self,
    ) -> dict:
        """Create a webhook secret.

        Returns:
            dict: Response indicating success
        """
        response = self._client._request(
            "POST",
            "/v1/secrets/webhook",
        )

        return response



    def delete(self) -> dict:
        """Delete a webhook secret.

        Returns:
            dict: Response indicating success
        """
        response = self._client._request(
            "DELETE",
            "/v1/secrets/webhook"
        )

        return response
