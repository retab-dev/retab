from typing import Any, Optional, Literal, List, Dict
from pydantic import HttpUrl, EmailStr
import json
from PIL.Image import Image
from pathlib import Path
from io import IOBase

from ..._resource import SyncAPIResource, AsyncAPIResource

from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.documents.image_operations import ImageOperations
from ...types.documents.text_operations import TextOperations
from ...types.mime import MIMEData
from ...types.modalities import Modality

from ..._utils.mime import prepare_mime_document


from ...types.automations.automations import MailboxConfig, AutomationConfig, UpdateMailBoxRequest, AutomationLog


class Emails(SyncAPIResource):
    """Emails API wrapper for managing email automation configurations"""

    def create(
        self,
        email: str,

        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,

        # email specific opitonals Fields
        follow_up: bool = False,
        authorized_domains: List[str] = [],
        authorized_emails: List[str] = [],
        # HTTP Config Optional Fields
        webhook_headers: Dict[str, str] = {},
        max_file_size: int = 50 ,
        file_payload: Literal["metadata_only", "file"] = "metadata_only",


        # DocumentExtraction Config
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
            "webhook_url": webhook_url,
            "webhook_headers": webhook_headers,
            "max_file_size": max_file_size,
            "file_payload": file_payload,
            "json_schema": json_schema,
            "follow_up": follow_up,
            "authorized_domains": authorized_domains,
            "authorized_emails": authorized_emails,
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
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        max_file_size: Optional[int] = None,
        file_payload: Optional[Literal["metadata_only", "file"]] = None,
        follow_up: Optional[bool] = None,
        authorized_domains: Optional[List[str]] = None,
        authorized_emails: Optional[List[str]] = None,
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
            webhook_url: New webhook configuration
            webhook_headers: New webhook configuration
            max_file_size: New webhook configuration
            file_payload: New webhook configuration
            follow_up: New webhook configuration
            authorized_domains: New webhook configuration
            authorized_emails: New webhook configuration
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
        if webhook_url is not None:
            data["webhook_url"] = webhook_url
        if webhook_headers is not None:
            data["webhook_headers"] = webhook_headers
        if max_file_size is not None:
            data["max_file_size"] = max_file_size
        if file_payload is not None:
            data["file_payload"] = file_payload
        if follow_up is not None:
            data["follow_up"] = follow_up
        if authorized_domains is not None:
            data["authorized_domains"] = authorized_domains
        if authorized_emails is not None:
            data["authorized_emails"] = authorized_emails
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




    def test_email_forwarding(self, 
                         email: str,
                         document: Path | str | IOBase | HttpUrl | Image | MIMEData,
                         verbose: bool = True
                         ) -> AutomationLog:
        """Mock endpoint that simulates the complete email forwarding process with sample data.
        
        Args:
            email: Email address of the mailbox to mock
            
        Returns:
            DocumentExtractResponse: The simulated extraction response
        """
        mime_document = prepare_mime_document(document)
        response = self._client._request("POST", f"/api/v1/emails/mailbox/test-email-forwarding/{email}", data={"document": mime_document.model_dump()})
        
        log = AutomationLog.model_validate(response)

        if verbose:
            print(f"\nTEST FILE UPLOAD RESULTS:")
            print(f"\n#########################")
            print(f"Status Code: {log.external_request_log.status_code}")
            print(f"Duration: {log.external_request_log.duration_ms:.2f}ms")
            
            if log.external_request_log.error:
                print(f"\nERROR: {log.external_request_log.error}")
            
            if log.external_request_log.response_body:
                print("\n--------------")
                print("RESPONSE BODY:")
                print("--------------")

                print(json.dumps(log.external_request_log.response_body, indent=2))
            if log.external_request_log.response_headers:
                print("\n--------------")
                print("RESPONSE HEADERS:")
                print("--------------")
                print(json.dumps(log.external_request_log.response_headers, indent=2))


        return log
    

    def test_webhook(self, 
                          email: str,
                          verbose: bool = True
                          ) -> AutomationLog:
        """Mock endpoint that simulates the complete webhook process with sample data.
        
        Args:
            email: Email address of the mailbox to mock
            
        Returns:
            AutomationLog: The simulated webhook response
        """

        response = self._client._request("POST", f"/api/v1/emails/mailbox/test-webhook/{email}")

        log = AutomationLog.model_validate(response)

        if verbose:
            print(f"\nTEST WEBHOOK RESULTS:")
            print(f"\n#########################")
            print(f"Status Code: {log.external_request_log.status_code}")
            print(f"Duration: {log.external_request_log.duration_ms:.2f}ms")
            
            if log.external_request_log.error:
                print(f"\nERROR: {log.external_request_log.error}")
            
            if log.external_request_log.response_body:
                print("\n--------------")
                print("RESPONSE BODY:")
                print("--------------")

                print(json.dumps(log.external_request_log.response_body, indent=2))
            if log.external_request_log.response_headers:
                print("\n--------------")
                print("RESPONSE HEADERS:")
                print("--------------")
                print(json.dumps(log.external_request_log.response_headers, indent=2))

        return log