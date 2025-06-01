import datetime
import json
from io import IOBase
from pathlib import Path
from typing import Any, Dict, List, Literal, Optional

import httpx
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from PIL.Image import Image
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.ai_models import assert_valid_model_extraction
from ..._utils.mime import prepare_mime_document
from ...types.deployments.mailboxes import ListMailboxes, Mailbox, UpdateMailboxRequest
from ...types.documents.extractions import UiParsedChatCompletion
from ...types.logs import DeploymentLog, ExternalRequestLog
from ...types.mime import BaseMIMEData, EmailData, MIMEData
from ...types.modalities import Modality
from ...types.standards import PreparedRequest


class MailBoxesMixin:
    def prepare_create(
        self,
        email: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        # email specific opitonals Fields
        authorized_domains: List[str] = [],
        authorized_emails: List[str] = [],
        # HTTP Config Optional Fields
        webhook_headers: Dict[str, str] = {},
        # DocumentExtraction Config
        image_resolution_dpi: int = 96,
        browser_canvas: Literal['A3', 'A4', 'A5'] = 'A4',
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
    ) -> PreparedRequest:
        assert_valid_model_extraction(model)

        data = {
            "email": email,
            "webhook_url": webhook_url,
            "webhook_headers": webhook_headers,
            "json_schema": json_schema,
            "authorized_domains": authorized_domains,
            "authorized_emails": authorized_emails,
            "image_resolution_dpi": image_resolution_dpi,
            "browser_canvas": browser_canvas,
            "modality": modality,
            "model": model,
            "temperature": temperature,
            "reasoning_effort": reasoning_effort,
        }

        # Validate the data
        mailbox_data = Mailbox.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/deployments/mailboxes", data=mailbox_data.model_dump(mode="json"))

    def prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        email: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> PreparedRequest:
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "email": email,
            "webhook_url": webhook_url,
            "schema_id": schema_id,
            "schema_data_id": schema_data_id,
        }
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}

        return PreparedRequest(method="GET", url="/v1/deployments/mailboxes", params=params)

    def prepare_get(self, email: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/deployments/mailboxes/{email}")

    def prepare_update(
        self,
        email: str,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        authorized_domains: Optional[List[str]] = None,
        authorized_emails: Optional[List[str]] = None,
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
        json_schema: Optional[Dict[str, Any]] = None,
    ) -> PreparedRequest:
        data: dict[str, Any] = {}
        if webhook_url is not None:
            data["webhook_url"] = webhook_url
        if webhook_headers is not None:
            data["webhook_headers"] = webhook_headers
        if authorized_domains is not None:
            data["authorized_domains"] = authorized_domains
        if authorized_emails is not None:
            data["authorized_emails"] = authorized_emails
        if image_resolution_dpi is not None:
            data["image_resolution_dpi"] = image_resolution_dpi
        if browser_canvas is not None:
            data["browser_canvas"] = browser_canvas
        if modality is not None:
            data["modality"] = modality
        if model is not None:
            assert_valid_model_extraction(model)
            data["model"] = model
        if temperature is not None:
            data["temperature"] = temperature
        if reasoning_effort is not None:
            data["reasoning_effort"] = reasoning_effort
        if json_schema is not None:
            data["json_schema"] = json_schema

        update_mailbox_request = UpdateMailboxRequest.model_validate(data)

        return PreparedRequest(method="PUT", url=f"/v1/deployments/mailboxes/{email}", data=update_mailbox_request.model_dump(mode="json"))

    def prepare_delete(self, email: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/deployments/mailboxes/{email}", raise_for_status=True)

    def prepare_logs(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        email: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> PreparedRequest:
        params = {
            "email": email,
            "webhook_url": webhook_url,
            "schema_id": schema_id,
            "schema_data_id": schema_data_id,
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
        }
        return PreparedRequest(method="GET", url=f"/v1/deployments/mailboxes/{email}/logs", params=params)


class Mailboxes(SyncAPIResource, MailBoxesMixin):
    """Emails API wrapper for managing email automation configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.tests = TestMailboxes(client=client)

    def create(
        self,
        email: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        # email specific opitonals Fields
        authorized_domains: List[str] = [],
        authorized_emails: List[str] = [],
        # HTTP Config Optional Fields
        webhook_headers: Dict[str, str] = {},
        # DocumentExtraction Config
        image_resolution_dpi: int = 96,
        browser_canvas: Literal['A3', 'A4', 'A5'] = 'A4',
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
    ) -> Mailbox:
        """Create a new email automation configuration.

        Args:
            email: Email address for the mailbox
            json_schema: JSON schema to validate extracted email data
            webhook_url: Webhook URL to receive processed emails
            webhook_headers: Webhook headers to send with processed emails
            authorized_domains: List of authorized domains for the mailbox
            authorized_emails: List of authorized emails for the mailbox
            image_resolution_dpi: Image resolution DPI
            browser_canvas: Browser canvas size
            modality: Processing modality (currently only "native" supported)
            model: AI model to use for processing
            temperature: Model temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.

        Returns:
            Mailbox: The created mailbox configuration
        """

        request = self.prepare_create(
            email, json_schema, webhook_url, authorized_domains, authorized_emails, webhook_headers, image_resolution_dpi, browser_canvas, modality, model, temperature, reasoning_effort
        )
        response = self._client._prepared_request(request)

        print(f"Email automation created. Mailbox available at https://www.uiform.com/dashboard/deployments/{response['id']}")

        return Mailbox.model_validate(response)

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        email: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> ListMailboxes:
        """List all email automation configurations.

        Args:
            before: Optional cursor for pagination - get results before this log ID
            after: Optional cursor for pagination - get results after this log ID
            limit: Maximum number of logs to return (1-100, default 10)
            order: Sort order by creation time - "asc" or "desc" (default "desc")
            email: Optional email address filter
            webhook_url: Optional webhook URL filter
            schema_id: Optional schema ID filter
            schema_data_id: Optional schema data ID filter

        Returns:
            ListMailboxes: List of mailbox configurations
        """
        request = self.prepare_list(before, after, limit, order, email, webhook_url, schema_id, schema_data_id)
        response = self._client._prepared_request(request)
        return ListMailboxes.model_validate(response)

    def get(self, email: str) -> Mailbox:
        """Get a specific email automation configuration.

        Args:
            email: Email address of the mailbox

        Returns:
            Mailbox: The mailbox configuration
        """
        request = self.prepare_get(email)
        response = self._client._prepared_request(request)
        return Mailbox.model_validate(response)

    def update(
        self,
        email: str,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        authorized_domains: Optional[List[str]] = None,
        authorized_emails: Optional[List[str]] = None,
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
        json_schema: Optional[Dict[str, Any]] = None,
    ) -> Mailbox:
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
            image_resolution_dpi: New image resolution DPI
            browser_canvas: New browser canvas size
            modality: New processing modality
            model: New AI model
            temperature: New temperature setting
            reasoning_effort: New reasoning effort
            json_schema: New JSON schema

        Returns:
            Mailbox: The updated mailbox configuration
        """
        request = self.prepare_update(
            email, webhook_url, webhook_headers, authorized_domains, authorized_emails, image_resolution_dpi, browser_canvas, modality, model, temperature, reasoning_effort, json_schema
        )
        response = self._client._prepared_request(request)
        return Mailbox.model_validate(response)

    def delete(self, email: str) -> None:
        """Delete an email automation configuration.

        Args:
            email: Email address of the mailbox to delete
        """
        request = self.prepare_delete(email)
        response = self._client._prepared_request(request)
        return None

    def logs(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        email: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> List[DeploymentLog]:
        """Get logs for a specific email automation.

        Args:
            before: Optional cursor for pagination - get results before this log ID
            after: Optional cursor for pagination - get results after this log ID
            limit: Maximum number of logs to return (1-100, default 10)
            order: Sort order by creation time - "asc" or "desc" (default "desc")
            email: Optional email address filter
            webhook_url: Optional webhook URL filter
            schema_id: Optional schema ID filter
            schema_data_id: Optional schema data ID filter

        Returns:
            List[Dict[str, Any]]: List of log entries
        """
        request = self.prepare_logs(before, after, limit, order, email, webhook_url, schema_id, schema_data_id)
        response = self._client._prepared_request(request)
        return [DeploymentLog.model_validate(log) for log in response]


class AsyncMailboxes(AsyncAPIResource, MailBoxesMixin):
    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.tests = AsyncTestMailboxes(client=client)

    async def create(
        self,
        email: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        authorized_domains: List[str] = [],
        authorized_emails: List[str] = [],
        webhook_headers: Dict[str, str] = {},
        image_resolution_dpi: int = 96,
        browser_canvas: Literal['A3', 'A4', 'A5'] = 'A4',
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
    ) -> Mailbox:
        request = self.prepare_create(
                email, json_schema, webhook_url, authorized_domains, authorized_emails, webhook_headers, image_resolution_dpi, browser_canvas, modality, model, temperature, reasoning_effort
        )
        response = await self._client._prepared_request(request)

        print(f"Email automation created. Mailbox available at https://www.uiform.com/dashboard/deployments/{response['id']}")

        return Mailbox.model_validate(response)

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        email: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> ListMailboxes:
        request = self.prepare_list(before, after, limit, order, email, webhook_url, schema_id, schema_data_id)
        response = await self._client._prepared_request(request)
        return ListMailboxes.model_validate(response)

    async def get(self, email: str) -> Mailbox:
        request = self.prepare_get(email)
        response = await self._client._prepared_request(request)
        return Mailbox.model_validate(response)

    async def update(
        self,
        email: str,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        authorized_domains: Optional[List[str]] = None,
        authorized_emails: Optional[List[str]] = None,
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
        json_schema: Optional[Dict[str, Any]] = None,
    ) -> Mailbox:
        request = self.prepare_update(
            email, webhook_url, webhook_headers, authorized_domains, authorized_emails, image_resolution_dpi, browser_canvas, modality, model, temperature, reasoning_effort, json_schema
        )
        response = await self._client._prepared_request(request)
        return Mailbox.model_validate(response)

    async def delete(self, email: str) -> None:
        request = self.prepare_delete(email)
        await self._client._prepared_request(request)
        return None

    async def logs(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        email: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> List[DeploymentLog]:
        request = self.prepare_logs(before, after, limit, order, email, webhook_url, schema_id, schema_data_id)
        response = await self._client._prepared_request(request)
        return [DeploymentLog.model_validate(log) for log in response]


class TestMailboxesMixin:
    def prepare_forward(
        self,
        email: str,
        document: Path | str | IOBase | HttpUrl | MIMEData,
        verbose: bool = True,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)
        return PreparedRequest(method="POST", url=f"/v1/deployments/mailboxes/tests/forward/{email}", data={"document": mime_document.model_dump()})

    def print_forward_verbose(self, email_data: EmailData) -> None:
        print(f"\nTEST EMAIL FORWARDING RESULTS:")
        print(f"\n#########################")
        print(f"Email ID: {email_data.id}")
        print(f"Subject: {email_data.subject}")
        print(f"From: {email_data.sender}")
        print(f"To: {', '.join(str(r) for r in email_data.recipients_to)}")
        if email_data.recipients_cc:
            print(f"CC: {', '.join(str(r) for r in email_data.recipients_cc)}")
        print(f"Sent at: {email_data.sent_at}")
        print(f"Attachments: {len(email_data.attachments)}")
        if email_data.body_plain:
            print("\nBody Preview:")
            print(email_data.body_plain[:500] + "..." if len(email_data.body_plain) > 500 else email_data.body_plain)

class TestMailboxes(SyncAPIResource, TestMailboxesMixin):
    def forward(
        self,
        email: str,
        document: Path | str | IOBase | HttpUrl | MIMEData,
        verbose: bool = True,
    ) -> EmailData:
        """Mock endpoint that simulates the complete email forwarding process with sample data.

        Args:
            email: Email address of the mailbox to mock

        Returns:
            DocumentExtractResponse: The simulated extraction response
        """
        request = self.prepare_forward(email, document, verbose)
        response = self._client._prepared_request(request)

        email_data = EmailData.model_validate(response)

        if verbose:
            self.print_forward_verbose(email_data)
        return email_data

class AsyncTestMailboxes(AsyncAPIResource, TestMailboxesMixin):
    async def forward(
        self,
        email: str,
        document: Path | str | IOBase | HttpUrl | MIMEData,
        verbose: bool = True,
    ) -> EmailData:
        request = self.prepare_forward(email, document, verbose)
        response = await self._client._prepared_request(request)
        email_data = EmailData.model_validate(response)
        if verbose:
            self.print_forward_verbose(email_data)
        return email_data
