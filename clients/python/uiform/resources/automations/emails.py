from typing import Any, Optional, Literal, List, Dict

from ..._resource import SyncAPIResource, AsyncAPIResource

from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.modalities import Modality

from typing import Any, Optional, Literal, List, Dict
from datetime import datetime
from pydantic import HttpUrl

from ..._resource import SyncAPIResource, AsyncAPIResource
from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.modalities import Modality

from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.documents.image_operations import ImageOperations
from ...types.documents.parse import DocumentExtractRequest, DocumentExtractResponse
from ...types.documents.text_operations import TextOperations



from .types import MailboxConfig, AutomationConfig, UpdateMailBoxRequest, AutomationLog

from pydantic import BaseModel

class Emails(SyncAPIResource):
    """Emails API wrapper for managing email automation configurations"""

    def create(
        self,
        email: str,
        http_config: Dict[str, Any],
        json_schema: Dict[str, Any],
        text_operations: Optional[Dict[str, Any]] = None,
        image_operations: Optional[Dict[str, Any]] = None,
        modality: Literal["native"] = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        additional_messages: List[ChatCompletionUiformMessage] = []
    ) -> MailboxConfig:
        """Create a new email automation configuration.
        
        Args:
            email: Email address for the mailbox
            http_config: Webhook configuration for forwarding processed emails
            json_schema: JSON schema to validate extracted email data
            text_operations: Optional text preprocessing operations
            image_operations: Optional image preprocessing operations
            modality: Processing modality (currently only "native" supported)
            model: AI model to use for processing
            temperature: Model temperature setting
            additional_messages: Optional additional context messages
            
        Returns:
            MailboxConfig: The created mailbox configuration
        """
        data = {
            "email": email,
            "http_config": http_config,
            "json_schema": json_schema,
            "text_operations": text_operations or TextOperations(),
            "image_operations": image_operations or ImageOperations(
                correct_image_orientation=True,
                dpi=72,
                image_to_text="ocr",
                browser_canvas="A4"
            ),
            "modality": modality,
            "model": model,
            "temperature": temperature,
            "additional_messages": additional_messages
        }

        request = MailboxConfig.model_validate(data)
        response = self._client._request("POST", "/api/v1/emails/mailbox", data=request.model_dump(mode="json"))
        
        return MailboxConfig.model_validate(response)

    def list(self) -> List[MailboxConfig]:
        """List all email automation configurations.
        
        Returns:
            List[MailboxConfig]: List of mailbox configurations
        """
        response = self._client._request("GET", "/api/v1/emails/mailbox")

        return [MailboxConfig.model_validate(mailbox) for mailbox in response]

    def get(self, email: str) -> MailboxConfig:
        """Get a specific email automation configuration.
        
        Args:
            email: Email address of the mailbox
            
        Returns:
            MailboxConfig: The mailbox configuration
        """
        response = self._client._request("GET", f"/api/v1/emails/mailbox/{email}")
        return MailboxConfig.model_validate(response)

    def update(
        self,
        email: str,
        http_config: Optional[Dict[str, Any]] = None,
        text_operations: Optional[Dict[str, Any]] = None,
        image_operations: Optional[Dict[str, Any]] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        additional_messages: Optional[List[ChatCompletionUiformMessage]] = None,
        json_schema: Optional[Dict[str, Any]] = None
    ) -> MailboxConfig:
        """Update an email automation configuration.
        
        Args:
            email: Email address of the mailbox to update
            http_config: New webhook configuration
            text_operations: New text preprocessing operations
            image_operations: New image preprocessing operations
            modality: New processing modality
            model: New AI model
            temperature: New temperature setting
            additional_messages: New context messages
            json_schema: New JSON schema
            
        Returns:
            MailboxConfig: The updated mailbox configuration
        """
        data: dict[str, Any] = {}
        if http_config is not None:
            data["http_config"] = http_config
        if text_operations is not None:
            data["text_operations"] = text_operations
        if image_operations is not None:
            data["image_operations"] = image_operations
        if modality is not None:
            data["modality"] = modality
        if model is not None:
            data["model"] = model
        if temperature is not None:
            data["temperature"] = temperature
        if additional_messages is not None:
            data["additional_messages"] = additional_messages
        if json_schema is not None:
            data["json_schema"] = json_schema

        update_mailbox_request = UpdateMailBoxRequest.model_validate(data)

        response = self._client._request("PUT", f"/api/v1/emails/mailbox/{email}", data=update_mailbox_request.model_dump())

        return MailboxConfig(**response)

    def delete(self, email: str) -> None:
        """Delete an email automation configuration.
        
        Args:
            email: Email address of the mailbox to delete
        """
        response = self._client._request("DELETE", f"/api/v1/emails/mailbox/{email}")


    def list_logs(self, email: str) -> List[AutomationLog]:
        """Get logs for a specific email automation.
        
        Args:
            email: Email address of the mailbox
            
        Returns:
            List[Dict[str, Any]]: List of log entries
        """
        response = self._client._request("GET", f"/api/v1/emails/mailbox/{email}/logs")

        return [AutomationLog.model_validate(log) for log in response]
