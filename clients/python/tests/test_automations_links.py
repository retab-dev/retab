
import pytest
from pydantic import HttpUrl
from typing import Any
from uiform import UiForm

@pytest.mark.asyncio
async def test_mailboxes_crud(sync_client: UiForm, company_json_schema: dict[str, Any]) -> None:
    name = "invoices15"
    model = "gpt-4o-mini"
    webhook_url = HttpUrl('http://localhost:4000/product')
    # Create
    link = sync_client.automations.links.create(
        name, 
        company_json_schema, 
        webhook_url,
        model=model,
    )
    link_id = link.id
    assert link.name == name
    assert link.webhook_url == webhook_url
    # Read
    link = sync_client.automations.links.get(link_id)
    assert link.name == name
    # Update
    link = sync_client.automations.links.update(link_id, webhook_url=HttpUrl('http://localhost:4000/product2'))
    assert link.webhook_url == HttpUrl('http://localhost:4000/product2')
    # Delete
    sync_client.automations.links.delete(link_id)
    with pytest.raises(Exception):
        sync_client.automations.links.get(link_id)