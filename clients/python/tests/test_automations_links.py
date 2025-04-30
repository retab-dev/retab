import base64
from typing import Any

import httpx
import nanoid  # type: ignore
import pytest
from pydantic import HttpUrl

from uiform import UiForm


@pytest.mark.asyncio
async def test_links_crud(sync_client: UiForm, company_json_schema: dict[str, Any], booking_confirmation_file_path: str) -> None:
    name = nanoid.generate()
    print("name", name)
    model = "gpt-4o-mini"
    webhook_url = HttpUrl('http://localhost:4000/v1/test_ingest_completion')
    # Create
    link = sync_client.deployments.links.create(
        name,
        company_json_schema,
        webhook_url,
        model=model,
    )
    link_id = link.id
    try:
        assert link.name == name
        assert link.webhook_url == webhook_url
        # Read
        link = sync_client.deployments.links.get(link_id)
        assert link.name == name
        # Update
        link = sync_client.deployments.links.update(link_id, password="password")
        assert link.password == "password"
        # Open and read the booking confirmation file

        # Let's upload a file
        with open(booking_confirmation_file_path, "rb") as f:
            async with httpx.AsyncClient(timeout=240) as client:
                usr_pwd_enc = base64.b64encode(f"{name}:password".encode("utf-8")).decode("utf-8")
                headers = {
                    "Authorization": f"Basic {usr_pwd_enc}",
                }
                files = {"file": f}
                response = await client.post(sync_client.base_url + f"/v1/deployments/links/parse/{link_id}", files=files, headers=headers)
                assert response.status_code == 200

        # Delete
        sync_client.deployments.links.delete(link_id)
        with pytest.raises(Exception):
            sync_client.deployments.links.get(link_id)
    finally:
        # Delete the link if it was created.
        try:
            sync_client.deployments.links.delete(link_id)
        except Exception:
            pass
