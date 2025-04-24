from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.standards import PreparedRequest


class WebhookMixin:
    def prepare_create(self) -> PreparedRequest:
        return PreparedRequest(method="POST", url="/v1/secrets/webhook")

    def prepare_delete(self) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url="/v1/secrets/webhook")


class Webhook(SyncAPIResource, WebhookMixin):
    """Webhook secret management wrapper"""

    def create(
        self,
    ) -> dict:
        """Create a webhook secret.

        Returns:
            dict: Response indicating success
        """
        request = self.prepare_create()
        response = self._client._prepared_request(request)

        return response

    def delete(self) -> dict:
        """Delete a webhook secret.

        Returns:
            dict: Response indicating success
        """
        request = self.prepare_delete()
        response = self._client._prepared_request(request)

        return response


class AsyncWebhook(AsyncAPIResource, WebhookMixin):
    """Webhook secret management wrapper"""

    async def create(self) -> dict:
        """Create a webhook secret.

        Returns:
            dict: Response indicating success
        """
        request = self.prepare_create()
        response = await self._client._prepared_request(request)
        return response

    async def delete(self) -> dict:
        """Delete a webhook secret.

        Returns:
            dict: Response indicating success
        """
        request = self.prepare_delete()
        response = await self._client._prepared_request(request)
        return response
