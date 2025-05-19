import hashlib
import hmac
import json
from typing import Any

from ..._resource import AsyncAPIResource, SyncAPIResource
from .endpoints import AsyncEndpoints, Endpoints
from .links import AsyncLinks, Links
from .mailboxes import AsyncMailboxes, Mailboxes
from .outlook import AsyncOutlooks, Outlooks
from .tests import AsyncTests, Tests
from .logs import AsyncLogs, Logs

class SignatureVerificationError(Exception):
    """Raised when webhook signature verification fails."""

    pass


class DeploymentsMixin:
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

        return json.loads(event_body.decode('utf-8'))


class Deployments(SyncAPIResource, DeploymentsMixin):
    """Deployments API wrapper"""

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


class AsyncDeployments(AsyncAPIResource, DeploymentsMixin):
    """Async Deployments API wrapper"""

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
