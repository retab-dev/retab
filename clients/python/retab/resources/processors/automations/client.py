import hashlib
import hmac
import json
from typing import Any, Literal, Optional, Union

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.automations.endpoints import Endpoint, UpdateEndpointRequest
from ....types.automations.links import Link, UpdateLinkRequest
from ....types.automations.mailboxes import Mailbox, UpdateMailboxRequest
from ....types.automations.outlook import Outlook, UpdateOutlookRequest
from ....types.standards import PreparedRequest
from .endpoints import AsyncEndpoints, Endpoints
from .links import AsyncLinks, Links
from .logs import AsyncLogs, Logs
from .mailboxes import AsyncMailboxes, Mailboxes
from .outlook import AsyncOutlooks, Outlooks
from .tests import AsyncTests, Tests


class SignatureVerificationError(Exception):
    """Raised when webhook signature verification fails."""

    pass


class AutomationsMixin:
    def _verify_event(self, event_body: bytes, event_signature: str, secret: str) -> Any:
        """
        Verify the signature of a webhook event.

        Args:
            body: The raw request body
            signature: The signature header
            secret: The secret key used for signing

        Returns:
            The parsed event payload

        Raises:
            SignatureVerificationError: If the signature verification fails
        """
        expected_signature = hmac.new(secret.encode(), event_body, hashlib.sha256).hexdigest()

        if not hmac.compare_digest(event_signature, expected_signature):
            raise SignatureVerificationError("Invalid signature")

        return json.loads(event_body.decode("utf-8"))

    def prepare_list(
        self,
        processor_id: str,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        automation_id: Optional[str] = None,
        webhook_url: Optional[str] = None,
        name: Optional[str] = None,
    ) -> PreparedRequest:
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "id": automation_id,
            "webhook_url": webhook_url,
            "name": name,
        }
        params = {k: v for k, v in params.items() if v is not None}
        return PreparedRequest(method="GET", url=f"/v1/processors/{processor_id}/automations", params=params)

    def prepare_get(self, processor_id: str, automation_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/processors/{processor_id}/automations/{automation_id}")

    def prepare_update(
        self,
        processor_id: str,
        automation_id: str,
        automation_data: Union[UpdateLinkRequest, UpdateMailboxRequest, UpdateEndpointRequest, UpdateOutlookRequest],
    ) -> PreparedRequest:
        return PreparedRequest(method="PUT", url=f"/v1/processors/{processor_id}/automations/{automation_id}", data=automation_data.model_dump(mode="json"))

    def prepare_delete(self, processor_id: str, automation_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/processors/{processor_id}/automations/{automation_id}")


class Automations(SyncAPIResource, AutomationsMixin):
    """Automations API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.mailboxes = Mailboxes(client=client)
        self.links = Links(client=client)
        self.outlook = Outlooks(client=client)
        self.endpoints = Endpoints(client=client)
        self.tests = Tests(client=client)
        self.logs = Logs(client=client)

    def verify_event(self, event_body: bytes, event_signature: str, secret: str) -> Any:
        """
        Verify the signature of a webhook event.
        """
        return self._verify_event(event_body, event_signature, secret)

    def list_automations(
        self,
        processor_id: str,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        automation_id: Optional[str] = None,
        webhook_url: Optional[str] = None,
        name: Optional[str] = None,
    ):
        """List automations attached to this processor."""
        request = self.prepare_list(processor_id, before, after, limit, order, automation_id, webhook_url, name)
        response = self._client._prepared_request(request)
        return response

    def get_automation(self, processor_id: str, automation_id: str) -> Union[Link, Mailbox, Endpoint, Outlook]:
        """Get a specific automation attached to this processor."""
        request = self.prepare_get(processor_id, automation_id)
        response = self._client._prepared_request(request)

        # Return the appropriate model based on the automation type
        if response["object"] == "automation.link":
            return Link.model_validate(response)
        elif response["object"] == "automation.mailbox":
            return Mailbox.model_validate(response)
        elif response["object"] == "automation.endpoint":
            return Endpoint.model_validate(response)
        elif response["object"] == "automation.outlook":
            return Outlook.model_validate(response)
        else:
            raise ValueError(f"Unknown automation type: {response.get('object')}")

    def update_automation(
        self,
        processor_id: str,
        automation_id: str,
        automation_data: Union[UpdateLinkRequest, UpdateMailboxRequest, UpdateEndpointRequest, UpdateOutlookRequest],
    ) -> Union[Link, Mailbox, Endpoint, Outlook]:
        """Update an automation attached to this processor."""
        request = self.prepare_update(processor_id, automation_id, automation_data)
        response = self._client._prepared_request(request)

        # Return the appropriate model based on the automation type
        if response["object"] == "automation.link":
            return Link.model_validate(response)
        elif response["object"] == "automation.mailbox":
            return Mailbox.model_validate(response)
        elif response["object"] == "automation.endpoint":
            return Endpoint.model_validate(response)
        elif response["object"] == "automation.outlook":
            return Outlook.model_validate(response)
        else:
            raise ValueError(f"Unknown automation type: {response.get('object')}")

    def delete_automation(self, processor_id: str, automation_id: str) -> None:
        """Delete an automation attached to this processor."""
        request = self.prepare_delete(processor_id, automation_id)
        self._client._prepared_request(request)
        print(f"Automation {automation_id} deleted from processor {processor_id}")


class AsyncAutomations(AsyncAPIResource, AutomationsMixin):
    """Async Automations API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.mailboxes = AsyncMailboxes(client=client)
        self.links = AsyncLinks(client=client)
        self.outlook = AsyncOutlooks(client=client)
        self.endpoints = AsyncEndpoints(client=client)
        self.tests = AsyncTests(client=client)
        self.logs = AsyncLogs(client=client)

    async def verify_event(self, event_body: bytes, event_signature: str, secret: str) -> Any:
        """
        Verify the signature of a webhook event.
        """
        return self._verify_event(event_body, event_signature, secret)

    async def list(
        self,
        processor_id: str,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        automation_id: Optional[str] = None,
        webhook_url: Optional[str] = None,
        name: Optional[str] = None,
    ):
        """List automations attached to this processor."""
        request = self.prepare_list(processor_id, before, after, limit, order, automation_id, webhook_url, name)
        response = await self._client._prepared_request(request)
        return response

    async def get(self, processor_id: str, automation_id: str) -> Union[Link, Mailbox, Endpoint, Outlook]:
        """Get a specific automation attached to this processor."""
        request = self.prepare_get(processor_id, automation_id)
        response = await self._client._prepared_request(request)

        # Return the appropriate model based on the automation type
        if response["object"] == "automation.link":
            return Link.model_validate(response)
        elif response["object"] == "automation.mailbox":
            return Mailbox.model_validate(response)
        elif response["object"] == "automation.endpoint":
            return Endpoint.model_validate(response)
        elif response["object"] == "automation.outlook":
            return Outlook.model_validate(response)
        else:
            raise ValueError(f"Unknown automation type: {response.get('object')}")

    async def update(
        self,
        processor_id: str,
        automation_id: str,
        automation_data: Union[UpdateLinkRequest, UpdateMailboxRequest, UpdateEndpointRequest, UpdateOutlookRequest],
    ) -> Union[Link, Mailbox, Endpoint, Outlook]:
        """Update an automation attached to this processor."""
        request = self.prepare_update(processor_id, automation_id, automation_data)
        response = await self._client._prepared_request(request)

        # Return the appropriate model based on the automation type
        if response["object"] == "automation.link":
            return Link.model_validate(response)
        elif response["object"] == "automation.mailbox":
            return Mailbox.model_validate(response)
        elif response["object"] == "automation.endpoint":
            return Endpoint.model_validate(response)
        elif response["object"] == "automation.outlook":
            return Outlook.model_validate(response)
        else:
            raise ValueError(f"Unknown automation type: {response.get('object')}")

    async def delete(self, processor_id: str, automation_id: str) -> None:
        """Delete an automation attached to this processor."""
        request = self.prepare_delete(processor_id, automation_id)
        await self._client._prepared_request(request)
        print(f"Automation {automation_id} deleted from processor {processor_id}")
