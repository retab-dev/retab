from typing import Any

import nanoid  # type: ignore
import pytest
from pydantic import HttpUrl

from uiform import UiForm


@pytest.mark.asyncio
async def test_mailboxes_crud(sync_client: UiForm, company_json_schema: dict[str, Any], booking_confirmation_file_path: str) -> None:
    test_idx = nanoid.generate().lower()
    email_address = f"bert2_{test_idx}@devmail.uiform.com"
    webhook_url = HttpUrl('http://localhost:4000/product')

    # Create
    mailbox = sync_client.deployments.mailboxes.create(email=email_address, json_schema=company_json_schema, webhook_url=webhook_url)
    try:
        assert mailbox.email == email_address
        assert mailbox.webhook_url == webhook_url
        # Read
        mailbox = sync_client.deployments.mailboxes.get(email_address)
        assert mailbox.email == email_address
        # Update
        mailbox = sync_client.deployments.mailboxes.update(email_address, webhook_url=HttpUrl('http://localhost:4000/product2'))
        assert mailbox.webhook_url == HttpUrl('http://localhost:4000/product2')

        # TODO: send and email to email_address (need sendgrid account)

        # Something like we did in test_automations_links.py, but not quite:
        # with open(booking_confirmation_file_path, "rb") as f:
        #     async with httpx.AsyncClient(timeout=240) as client:
        #         usr_pwd_enc = base64.b64encode(f"{name}:password".encode("utf-8")).decode("utf-8")
        #         headers = {
        #             "Authorization": f"Basic {usr_pwd_enc}",
        #         }
        #         files = {
        #             "file": f
        #         }
        #         response = await client.post(
        #             sync_client.base_url + f"/v1/deployments/links/parse/{link_id}",
        #             files=files,
        #             headers=headers
        #         )
        #         assert response.status_code == 200

        # Delete
        sync_client.deployments.mailboxes.delete(email_address)
        with pytest.raises(Exception):
            sync_client.deployments.mailboxes.get(email_address)
    finally:
        # Delete the mailbox
        try:
            sync_client.deployments.mailboxes.delete(email_address)
        except Exception:
            pass
