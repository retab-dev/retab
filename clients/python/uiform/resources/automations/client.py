from typing import Any
import hmac
import hashlib
import json

from ..._resource import SyncAPIResource, AsyncAPIResource

from .emails import Emails
from .extraction_links import ExtractionLinks

class SignatureVerificationError(Exception):
    """Raised when webhook signature verification fails."""
    pass

class Automations(SyncAPIResource): 
    """Automations API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.emails = Emails(client=client)
        self.extraction_links = ExtractionLinks(client=client)



    def verify_event(self, body: bytes, signature: str, secret: str) -> Any:
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
        expected_signature = hmac.new(
            secret.encode(),
            body,
            hashlib.sha256
        ).hexdigest()

        if not hmac.compare_digest(signature, expected_signature):
            raise SignatureVerificationError("Invalid signature")

        return json.loads(body.decode('utf-8'))
