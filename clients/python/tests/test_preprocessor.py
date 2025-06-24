from typing import Any

import pytest

from retab import Retab


@pytest.mark.asyncio
async def test_preprocessor_crud(sync_client: Retab, company_json_schema: dict[str, Any]) -> None:
    name = "test_preprocessor"
    model = "gpt-4o-mini"

    # Create a processor
    processor = sync_client.processors.create(
        name=name,
        json_schema=company_json_schema,
        model=model,
    )
    processor_id = processor.id

    try:
        # Get the processor
        retrieved_processor = sync_client.processors.get(processor_id)
        assert retrieved_processor.name == name
        assert retrieved_processor.model == model

        # List processors
        processors = sync_client.processors.list(limit=10)
        assert len(processors.data) > 0
        assert any(p.id == processor_id for p in processors.data)

        # Update processor
        updated_name = "updated_test_preprocessor"
        updated_processor = sync_client.processors.update(
            processor_id,
            name=updated_name,
        )
        assert updated_processor.name == updated_name

        # Delete processor
        sync_client.processors.delete(processor_id)
        with pytest.raises(Exception):
            sync_client.processors.get(processor_id)
    finally:
        # Clean up
        try:
            sync_client.processors.delete(processor_id)
        except Exception:
            pass
