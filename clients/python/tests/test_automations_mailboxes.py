
import pytest
from pydantic import HttpUrl
from typing import Any
from uiform import UiForm

@pytest.mark.asyncio
async def test_mailboxes_crud(sync_client: UiForm, company_json_schema: dict[str, Any]) -> None:
    email_address = "bert2@devmail.uiform.com"
    webhook_url = HttpUrl('http://localhost:4000/product')
    # Create
    mailbox = sync_client.automations.mailboxes.create(
        email=email_address, 
        json_schema=company_json_schema, 
        webhook_url=webhook_url
    )
    assert mailbox.email == email_address
    assert mailbox.webhook_url == webhook_url
    # Read
    mailbox = sync_client.automations.mailboxes.get(email_address)
    assert mailbox.email == email_address
    # Update
    mailbox = sync_client.automations.mailboxes.update(email_address, webhook_url=HttpUrl('http://localhost:4000/product2'))
    assert mailbox.webhook_url == HttpUrl('http://localhost:4000/product2')
    # Delete
    sync_client.automations.mailboxes.delete(email_address)
    with pytest.raises(Exception):
        sync_client.automations.mailboxes.get(email_address)
