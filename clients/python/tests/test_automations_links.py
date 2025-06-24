import base64
from typing import Any

import httpx
import nanoid  # type: ignore
import pytest

from retab import Retab


@pytest.mark.asyncio
async def test_links_crud(sync_client: Retab, company_json_schema: dict[str, Any], booking_confirmation_file_path_1: str) -> None:
    name = nanoid.generate()
    print("name", name)
    model = "gpt-4o-mini"
    webhook_url = "http://localhost:4000/v1/test_ingest_completion"

    # First create a processor
    link_id: str | None = None
    processor_id: str | None = None

    processor = sync_client.processors.create(
        name=name,
        json_schema=company_json_schema,
        model=model,
    )
    processor_id = processor.id

    try:
        # Create the link
        link = sync_client.processors.automations.links.create(
            processor_id=processor_id,
            name=name,
            webhook_url=webhook_url,
        )
        link_id = link.id
        assert link.name == name
        assert link.webhook_url == webhook_url
        # Read
        link = sync_client.processors.automations.links.get(link_id)
        assert link.name == name
        # Update
        link = sync_client.processors.automations.links.update(link_id, password="password")
        assert link.password == "password"
        # Open and read the booking confirmation file

        # Let's upload a file
        with open(booking_confirmation_file_path_1, "rb") as f:
            async with httpx.AsyncClient(timeout=240) as client:
                usr_pwd_enc = base64.b64encode(f"{name}:password".encode("utf-8")).decode("utf-8")
                headers = {
                    "Authorization": f"Basic {usr_pwd_enc}",
                }
                files = {"file": f}
                response = await client.post(sync_client.base_url + f"/v1/processors/automations/links/parse/{link_id}", files=files, headers=headers)
                assert response.status_code == 200

        # Delete the link
        sync_client.processors.automations.links.delete(link_id)
        with pytest.raises(Exception):
            sync_client.processors.automations.links.get(link_id)
        link_id = None
    finally:
        # Delete the link if it was created
        if link_id:
            try:
                sync_client.processors.automations.links.delete(link_id)
            except Exception:
                pass
        # Delete the processor if it was created
        if processor_id:
            try:
                sync_client.processors.delete(processor_id)
            except Exception:
                pass
