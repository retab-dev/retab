from ..._resource import SyncAPIResource

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
