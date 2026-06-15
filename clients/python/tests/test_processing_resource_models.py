from retab.types.classifications import Classification
from retab.types.edits import Edit
from retab.types.extractions import Extraction
from retab.types.partitions import Partition
from retab.types.parses import Parse
from retab.types.splits import Split

import pytest

# Whole module is unit (pure offline; no server/credentials needed).
pytestmark = pytest.mark.unit


def test_classification_model_accepts_canonical_consensus_shape() -> None:
    classification = Classification.model_validate(
        {
            "id": "clss_123",
            "file": {
                "id": "file_123",
                "filename": "doc.pdf",
                "mime_type": "application/pdf",
            },
            "model": "retab-small",
            "categories": [{"name": "invoice", "description": "Invoice"}],
            "n_consensus": 1,
            "output": {"reasoning": "", "category": "invoice"},
            "consensus": {"choices": [], "likelihood": None},
            "updated_at": "2026-04-11T12:31:00Z",
        }
    )

    assert classification.consensus is not None
    assert classification.consensus.choices == []
    assert "updated_at" not in Classification.model_fields
    assert "updated_at" not in Classification.model_json_schema()["properties"]
    assert "updated_at" not in classification.model_dump(mode="json")


def test_extraction_model_accepts_canonical_consensus_shape() -> None:
    extraction = Extraction.model_validate(
        {
            "id": "extr_123",
            "file": {
                "id": "file_123",
                "filename": "doc.pdf",
                "mime_type": "application/pdf",
            },
            "model": "retab-small",
            "json_schema": {"type": "object"},
            "n_consensus": 1,
            "image_resolution_dpi": 192,
            "output": {"invoice_number": "INV-001"},
            "consensus": {"choices": [], "likelihoods": None},
            "metadata": {},
            "updated_at": "2026-04-01T13:00:00Z",
        }
    )

    assert extraction.consensus is not None
    assert extraction.consensus.choices == []
    assert "updated_at" not in Extraction.model_fields
    assert "updated_at" not in Extraction.model_json_schema()["properties"]
    assert "updated_at" not in extraction.model_dump(mode="json")


def test_split_model_accepts_single_vote_resource_shape() -> None:
    split = Split.model_validate(
        {
            "id": "splt_123",
            "file": {
                "id": "file_123",
                "filename": "doc.pdf",
                "mime_type": "application/pdf",
            },
            "model": "retab-small",
            "subdocuments": [{"name": "invoice", "description": "Invoice"}],
            "n_consensus": 1,
            "output": [{"name": "invoice", "pages": [1]}],
            "consensus": None,
        }
    )

    assert split.consensus is None


def test_split_model_excludes_updated_at() -> None:
    split = Split.model_validate(
        {
            "id": "splt_123",
            "file": {
                "id": "file_123",
                "filename": "doc.pdf",
                "mime_type": "application/pdf",
            },
            "model": "retab-small",
            "subdocuments": [{"name": "invoice", "description": "Invoice"}],
            "n_consensus": 1,
            "output": [{"name": "invoice", "pages": [1]}],
            "updated_at": "2026-05-20T10:00:00+00:00",
        }
    )

    assert "updated_at" not in Split.model_fields
    assert "updated_at" not in split.model_dump(mode="json")


def test_partition_model_accepts_canonical_consensus_shape() -> None:
    partition = Partition.model_validate(
        {
            "id": "prtn_123",
            "file": {
                "id": "file_123",
                "filename": "doc.pdf",
                "mime_type": "application/pdf",
            },
            "model": "retab-small",
            "key": "invoice_number",
            "instructions": "Group by invoice number.",
            "n_consensus": 1,
            "output": [{"key": "INV-001", "pages": [1]}],
            "consensus": {"choices": [], "likelihoods": None},
        }
    )

    assert partition.consensus is not None
    assert partition.consensus.choices == []


def test_parse_model_excludes_updated_at_from_public_schema_and_output() -> None:
    parse = Parse.model_validate(
        {
            "id": "prs_123",
            "file": {
                "id": "file_123",
                "filename": "doc.pdf",
                "mime_type": "application/pdf",
            },
            "model": "retab-small",
            "table_parsing_format": "html",
            "image_resolution_dpi": 192,
            "output": {"pages": ["page 1"], "text": "page 1"},
            "updated_at": "2026-04-17T10:00:00Z",
        }
    )
    schema = Parse.model_json_schema()

    assert "updated_at" not in Parse.model_fields
    assert "updated_at" not in schema["properties"]
    assert "updated_at" not in parse.model_dump(mode="json")


def test_edit_model_excludes_updated_at_from_public_schema_and_output() -> None:
    edit = Edit.model_validate(
        {
            "id": "edt_123",
            "file": {
                "id": "file_123",
                "filename": "form.pdf",
                "mime_type": "application/pdf",
            },
            "model": "retab-small",
            "instructions": "Fill the form.",
            "config": {"color": "#000080"},
            "output": {
                "form_data": [],
                "filled_document": {
                    "filename": "filled.pdf",
                    "url": "data:application/pdf;base64,AA==",
                },
            },
            "usage": {"credits": 1.0},
            "updated_at": "2026-04-17T10:00:00Z",
        }
    )
    schema = Edit.model_json_schema()

    assert "updated_at" not in Edit.model_fields
    assert "updated_at" not in schema["properties"]
    assert "updated_at" not in edit.model_dump(mode="json")


def test_partition_model_excludes_updated_at() -> None:
    partition = Partition.model_validate(
        {
            "id": "prtn_123",
            "file": {
                "id": "file_123",
                "filename": "doc.pdf",
                "mime_type": "application/pdf",
            },
            "model": "retab-small",
            "key": "invoice_number",
            "instructions": "Group by invoice number.",
            "n_consensus": 1,
            "output": [{"key": "INV-001", "pages": [1]}],
            "consensus": {"choices": [], "likelihoods": None},
            "updated_at": "2026-05-20T10:00:00+00:00",
        }
    )

    assert "updated_at" not in Partition.model_fields
    assert "updated_at" not in partition.model_dump(mode="json")
