from typing import Any
import os

import nanoid  # type: ignore
import pytest

from retab import Retab


@pytest.mark.asyncio
async def test_mailboxes_crud(sync_client: Retab, company_json_schema: dict[str, Any], booking_confirmation_file_path_1: str) -> None:
    test_idx = nanoid.generate().lower()
    name = f"test_mailbox_{test_idx}"
    email_address = f"bert2_{test_idx}@{os.getenv('EMAIL_DOMAIN', 'mailbox.retab.com')}"
    webhook_url = "http://localhost:4000/product"
    model = "gpt-4o-mini"

    # First create a processor
    processor = sync_client.processors.create(
        name=name,
        json_schema=company_json_schema,
        model=model,
    )
    processor_id = processor.id

    try:
        # Create the mailbox
        mailbox = sync_client.processors.automations.mailboxes.create(
            processor_id=processor_id,
            name=name,
            email=email_address,
            webhook_url=webhook_url,
        )
        try:
            assert mailbox.email == email_address
            assert mailbox.webhook_url == webhook_url
            # Read
            mailbox = sync_client.processors.automations.mailboxes.get(email_address)
            assert mailbox.email == email_address
            # Update
            mailbox = sync_client.processors.automations.mailboxes.update(email_address, webhook_url="http://localhost:4000/product2")
            assert mailbox.webhook_url == "http://localhost:4000/product2"

            # TODO: send and email to email_address (need sendgrid account)

            # Delete the mailbox
            sync_client.processors.automations.mailboxes.delete(email_address)
            with pytest.raises(Exception):
                sync_client.processors.automations.mailboxes.get(email_address)
        finally:
            # Delete the mailbox if it was created
            try:
                sync_client.processors.automations.mailboxes.delete(email_address)
            except Exception:
                pass
    finally:
        # Delete the processor if it was created
        try:
            sync_client.processors.delete(processor_id)
        except Exception:
            pass
