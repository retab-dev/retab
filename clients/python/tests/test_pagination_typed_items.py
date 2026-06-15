"""Regression tests for typed-item re-validation on paginated list endpoints.

When a list endpoint constructs ``PaginatedList(**response)`` without supplying
the generic ``T``, pydantic infers ``T = Any`` and leaves each ``result.data[i]``
as a raw dict. Callers iterating ``for item in client.extractions.list(): ...``
then trip ``AttributeError`` instead of getting a typed model.

The fix pattern (already present on ``workflows.list``) is to re-validate each
item against its concrete type after constructing the envelope:

    result = PaginatedList(**response)
    result.data = [Cls.model_validate(item) if isinstance(item, dict) else item
                   for item in result.data]

These tests pin that contract across all seven resources, sync and async.
"""

from __future__ import annotations

from typing import Any

import pytest

from retab.resources.classifications import (
    AsyncClassifications,
    Classifications,
)
from retab.resources.edits import AsyncEdits, Edits
from retab.resources.edits.templates import (
    AsyncEditTemplates,
    EditTemplates,
)
from retab.resources.extractions import AsyncExtractions, Extractions
from retab.resources.parses import AsyncParses, Parses
from retab.resources.partitions import AsyncPartitions, Partitions
from retab.resources.splits import AsyncSplits, Splits
from retab.types.classifications import Classification
from retab.types.edits import Edit, EditTemplate
from retab.types.extractions import Extraction
from retab.types.pagination import AsyncPaginatedList, PaginatedList
from retab.types.parses import Parse
from retab.types.partitions import Partition
from retab.types.splits import Split
from samples import list_envelope
from mocks import mock_async_client, mock_sync_client

# Whole module is creditless (live list/get/pagination only; no credits).
pytestmark = pytest.mark.creditless


# ---------------------------------------------------------------------------
# Fixture builders — minimal-but-valid records for each model
# ---------------------------------------------------------------------------


_FILE_REF = {"id": "file_abc", "filename": "doc.pdf", "mime_type": "application/pdf"}
_NOW = "2026-05-01T14:30:00Z"


def _extraction_item() -> dict[str, Any]:
    return {
        "id": "ext_1",
        "file": _FILE_REF,
        "model": "retab-small",
        "json_schema": {"type": "object", "properties": {}},
        "output": {"total": 100},
    }


def _parse_item() -> dict[str, Any]:
    return {
        "id": "parse_1",
        "file": _FILE_REF,
        "model": "retab-small",
        "table_parsing_format": "html",
        "image_resolution_dpi": 192,
        "output": {"pages": ["page 1 text"], "text": "page 1 text"},
    }


def _partition_item() -> dict[str, Any]:
    return {
        "id": "part_1",
        "file": _FILE_REF,
        "model": "retab-small",
        "key": "section",
        "output": [{"key": "intro", "pages": [1]}],
    }


def _split_item() -> dict[str, Any]:
    return {
        "id": "split_1",
        "file": _FILE_REF,
        "model": "retab-small",
        "subdocuments": [{"name": "invoice"}],
        "output": [{"name": "invoice", "pages": [1, 2]}],
    }


def _classification_item() -> dict[str, Any]:
    return {
        "id": "class_1",
        "file": _FILE_REF,
        "model": "retab-small",
        "categories": [{"name": "invoice"}],
        "output": {"reasoning": "matches invoice format", "category": "invoice"},
    }


def _edit_item() -> dict[str, Any]:
    return {
        "id": "edit_1",
        "file": _FILE_REF,
        "model": "retab-small",
        "config": {"color": "#000080"},
        "output": {
            "form_data": [],
            "filled_document": {
                "filename": "filled.pdf",
                "url": "data:application/pdf;base64,",
            },
        },
    }


def _edit_template_item() -> dict[str, Any]:
    return {
        "id": "tpl_1",
        "name": "My template",
        "file": _FILE_REF,
        "form_fields": [],
        "field_count": 0,
        "created_at": _NOW,
        "updated_at": _NOW,
    }


# ---------------------------------------------------------------------------
# Parametrized cases — one row per (sync resource, expected model) pair
# ---------------------------------------------------------------------------


_SYNC_CASES = [
    pytest.param(Extractions, _extraction_item, Extraction, id="extractions"),
    pytest.param(Parses, _parse_item, Parse, id="parses"),
    pytest.param(Partitions, _partition_item, Partition, id="partitions"),
    pytest.param(Splits, _split_item, Split, id="splits"),
    pytest.param(Classifications, _classification_item, Classification, id="classifications"),
    pytest.param(Edits, _edit_item, Edit, id="edits"),
    pytest.param(EditTemplates, _edit_template_item, EditTemplate, id="edits.templates"),
]


_ASYNC_CASES = [
    pytest.param(AsyncExtractions, _extraction_item, Extraction, id="extractions"),
    pytest.param(AsyncParses, _parse_item, Parse, id="parses"),
    pytest.param(AsyncPartitions, _partition_item, Partition, id="partitions"),
    pytest.param(AsyncSplits, _split_item, Split, id="splits"),
    pytest.param(AsyncClassifications, _classification_item, Classification, id="classifications"),
    pytest.param(AsyncEdits, _edit_item, Edit, id="edits"),
    pytest.param(AsyncEditTemplates, _edit_template_item, EditTemplate, id="edits.templates"),
]


# ---------------------------------------------------------------------------
# Sync — list endpoints must return typed model instances, not raw dicts
# ---------------------------------------------------------------------------


@pytest.mark.parametrize("resource_cls, item_factory, expected_model", _SYNC_CASES)
def test_sync_list_returns_typed_items(
    resource_cls: type,
    item_factory: Any,
    expected_model: type,
) -> None:
    item = item_factory()
    client = mock_sync_client(list_envelope(item, item))

    result = resource_cls(client=client).list()

    assert isinstance(result, PaginatedList)
    assert len(result) == 2
    # The bug: items come back as plain dicts when the generic isn't supplied
    # and items aren't re-validated. The fix is per-item .model_validate().
    assert isinstance(result.data[0], expected_model), f"{resource_cls.__name__}.list() returned a plain dict at data[0] — expected {expected_model.__name__}"
    assert isinstance(result.data[1], expected_model)


# ---------------------------------------------------------------------------
# Async — same contract on the async resources
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
@pytest.mark.parametrize("resource_cls, item_factory, expected_model", _ASYNC_CASES)
async def test_async_list_returns_typed_items(
    resource_cls: type,
    item_factory: Any,
    expected_model: type,
) -> None:
    item = item_factory()
    client = mock_async_client(list_envelope(item, item))

    result = await resource_cls(client=client).list()

    # Async resources return AsyncPaginatedList; sync return PaginatedList.
    # Both expose the same `.data`/`__len__`/`__getitem__` surface.
    assert isinstance(result, AsyncPaginatedList)
    assert len(result) == 2
    assert isinstance(result.data[0], expected_model), f"{resource_cls.__name__}.list() returned a plain dict at data[0] — expected {expected_model.__name__}"
    assert isinstance(result.data[1], expected_model)
