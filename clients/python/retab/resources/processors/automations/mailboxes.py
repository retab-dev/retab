from io import IOBase
from pathlib import Path
from typing import Any, List, Literal, Optional

from pydantic import EmailStr, HttpUrl
from pydantic_core import PydanticUndefined

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....utils.mime import prepare_mime_document
from ....types.automations.mailboxes import ListMailboxes, Mailbox, UpdateMailboxRequest
from ....types.mime import EmailData, MIMEData
from ....types.standards import PreparedRequest


class MailBoxesMixin:
    mailboxes_base_url: str = "/v1/processors/automations/mailboxes"

    def prepare_create(
        self,
        email: str,
        name: str,
        processor_id: str,
        webhook_url: str,
        authorized_domains: list[str] = PydanticUndefined,  # type: ignore[assignment]
        authorized_emails: list[EmailStr] = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        default_language: str = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
    ) -> PreparedRequest:
        """Create a new email automation configuration."""
        mailbox = Mailbox(
            name=name,
            processor_id=processor_id,
            default_language=default_language,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            email=email,
            authorized_domains=authorized_domains,
            authorized_emails=authorized_emails,
        )
        return PreparedRequest(method="POST", url=self.mailboxes_base_url, data=mailbox.model_dump(mode="json"))

    def prepare_list(
        self,
        processor_id: str,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        email: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
    ) -> PreparedRequest:
        params = {
            "processor_id": processor_id,
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "email": email,
            "name": name,
            "webhook_url": webhook_url,
        }
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}

        return PreparedRequest(method="GET", url=self.mailboxes_base_url, params=params)

    def prepare_get(self, mailbox_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{self.mailboxes_base_url}/{mailbox_id}")

    def prepare_update(
        self,
        mailbox_id: str,
        name: str = PydanticUndefined,  # type: ignore[assignment]
        default_language: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_url: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
        authorized_domains: list[str] = PydanticUndefined,  # type: ignore[assignment]
        authorized_emails: list[str] = PydanticUndefined,  # type: ignore[assignment]
    ) -> PreparedRequest:
        update_mailbox_request = UpdateMailboxRequest(
            name=name,
            default_language=default_language,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            authorized_domains=authorized_domains,
            authorized_emails=authorized_emails,
        )
        return PreparedRequest(method="PUT", url=f"/v1/processors/automations/mailboxes/{mailbox_id}", data=update_mailbox_request.model_dump(mode="json"))

    def prepare_delete(self, mailbox_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/processors/automations/mailboxes/{mailbox_id}", raise_for_status=True)


class Mailboxes(SyncAPIResource, MailBoxesMixin):
    """Emails API wrapper for managing email automation configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.tests = TestMailboxes(client=client)

    def create(
        self,
        email: str,
        name: str,
        webhook_url: str,
        processor_id: str,
        authorized_domains: list[str] = PydanticUndefined,  # type: ignore[assignment]
        authorized_emails: list[EmailStr] = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        default_language: str = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
    ) -> Mailbox:
        """Create a new email automation configuration.

        Args:
            email: Email address for the mailbox
            name: Name of the mailbox
            webhook_url: Webhook URL to receive processed emails
            processor_id: ID of the processor to use for the mailbox
            authorized_domains: List of authorized domains for the mailbox
            authorized_emails: List of authorized emails for the mailbox
            webhook_headers: Webhook headers to send with processed emails

        Returns:
            Mailbox: The created mailbox configuration
        """

        request = self.prepare_create(
            email=email,
            name=name,
            processor_id=processor_id,
            webhook_url=webhook_url,
            authorized_domains=authorized_domains,
            authorized_emails=authorized_emails,
            webhook_headers=webhook_headers,
            default_language=default_language,
            need_validation=need_validation,
        )
        response = self._client._prepared_request(request)

        print(f"Mailbox {response['email']} created. Url: https://www.retab.com/dashboard/processors/automations/{response['id']}")

        return Mailbox.model_validate(response)

    def list(
        self,
        processor_id: str,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        name: Optional[str] = None,
        email: Optional[str] = None,
        webhook_url: Optional[str] = None,
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
        request = self.prepare_list(
            processor_id=processor_id,
            before=before,
            after=after,
            limit=limit,
            order=order,
            email=email,
            name=name,
            webhook_url=webhook_url,
        )
        response = self._client._prepared_request(request)
        return ListMailboxes.model_validate(response)

    def get(self, mailbox_id: str) -> Mailbox:
        """Get a specific email automation configuration.

        Args:
            mailbox_id: ID of the mailbox

        Returns:
            Mailbox: The mailbox configuration
        """
        request = self.prepare_get(mailbox_id)
        response = self._client._prepared_request(request)
        return Mailbox.model_validate(response)

    def update(
        self,
        mailbox_id: str,
        name: str = PydanticUndefined,  # type: ignore[assignment]
        default_language: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_url: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
        authorized_domains: List[str] = PydanticUndefined,  # type: ignore[assignment]
        authorized_emails: List[str] = PydanticUndefined,  # type: ignore[assignment]
    ) -> Mailbox:
        """Update an email automation configuration.

        Args:
            email: Email address of the mailbox to update
            name: New name
            default_language: New default language
            webhook_url: New webhook configuration
            webhook_headers: New webhook configuration
            need_validation: New webhook configuration
            authorized_domains: New webhook configuration
            authorized_emails: New webhook configuration

        Returns:
            Mailbox: The updated mailbox configuration
        """
        request = self.prepare_update(
            mailbox_id=mailbox_id,
            name=name,
            default_language=default_language,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            authorized_domains=authorized_domains,
            authorized_emails=authorized_emails,
        )
        response = self._client._prepared_request(request)
        return Mailbox.model_validate(response)

    def delete(self, mailbox_id: str) -> None:
        """Delete an email automation configuration.

        Args:
            email: Email address of the mailbox to delete
        """
        request = self.prepare_delete(mailbox_id)
        self._client._prepared_request(request)
        return None


class AsyncMailboxes(AsyncAPIResource, MailBoxesMixin):
    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.tests = AsyncTestMailboxes(client=client)

    async def create(
        self,
        email: str,
        name: str,
        webhook_url: str,
        processor_id: str,
        authorized_domains: List[str] = PydanticUndefined,  # type: ignore[assignment]
        authorized_emails: List[EmailStr] = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        default_language: str = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
    ) -> Mailbox:
        request = self.prepare_create(
            email=email,
            name=name,
            processor_id=processor_id,
            webhook_url=webhook_url,
            authorized_domains=authorized_domains,
            authorized_emails=authorized_emails,
            webhook_headers=webhook_headers,
            default_language=default_language,
            need_validation=need_validation,
        )
        response = await self._client._prepared_request(request)

        print(f"Mailbox {response['email']} created. Url: https://www.retab.com/dashboard/processors/automations/{response['id']}")

        return Mailbox.model_validate(response)

    async def list(
        self,
        processor_id: str,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        email: str | None = None,
        name: str | None = None,
        webhook_url: str | None = None,
    ) -> ListMailboxes:
        request = self.prepare_list(
            processor_id=processor_id,
            before=before,
            after=after,
            limit=limit,
            order=order,
            email=email,
            name=name,
            webhook_url=webhook_url,
        )
        response = await self._client._prepared_request(request)
        return ListMailboxes.model_validate(response)

    async def get(self, mailbox_id: str) -> Mailbox:
        request = self.prepare_get(mailbox_id)
        response = await self._client._prepared_request(request)
        return Mailbox.model_validate(response)

    async def update(
        self,
        mailbox_id: str,
        name: str = PydanticUndefined,  # type: ignore[assignment]
        default_language: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_url: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
        authorized_domains: List[str] = PydanticUndefined,  # type: ignore[assignment]
        authorized_emails: List[str] = PydanticUndefined,  # type: ignore[assignment]
    ) -> Mailbox:
        request = self.prepare_update(
            mailbox_id=mailbox_id,
            name=name,
            default_language=default_language,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            authorized_domains=authorized_domains,
            authorized_emails=authorized_emails,
        )
        response = await self._client._prepared_request(request)
        return Mailbox.model_validate(response)

    async def delete(self, mailbox_id: str) -> None:
        request = self.prepare_delete(mailbox_id)
        await self._client._prepared_request(request)
        return None


class TestMailboxesMixin:
    def prepare_forward(
        self,
        mailbox_id: str,
        document: Path | str | IOBase | HttpUrl | MIMEData,
        verbose: bool = True,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)
        return PreparedRequest(method="POST", url=f"/v1/processors/automations/mailboxes/tests/forward/{mailbox_id}", data={"document": mime_document.model_dump()})

    def print_forward_verbose(self, email_data: EmailData) -> None:
        print("\nTEST EMAIL FORWARDING RESULTS:")
        print("\n#########################")
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
        mailbox_id: str,
        document: Path | str | IOBase | HttpUrl | MIMEData,
        verbose: bool = True,
    ) -> EmailData:
        """Mock endpoint that simulates the complete email forwarding process with sample data.

        Args:
            email: Email address of the mailbox to mock

        Returns:
            DocumentExtractResponse: The simulated extraction response
        """
        request = self.prepare_forward(mailbox_id, document, verbose)
        response = self._client._prepared_request(request)

        email_data = EmailData.model_validate(response)

        if verbose:
            self.print_forward_verbose(email_data)
        return email_data


class AsyncTestMailboxes(AsyncAPIResource, TestMailboxesMixin):
    async def forward(
        self,
        mailbox_id: str,
        document: Path | str | IOBase | HttpUrl | MIMEData,
        verbose: bool = True,
    ) -> EmailData:
        request = self.prepare_forward(mailbox_id, document, verbose)
        response = await self._client._prepared_request(request)
        email_data = EmailData.model_validate(response)
        if verbose:
            self.print_forward_verbose(email_data)
        return email_data
