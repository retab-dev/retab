from retab.types.classifications import Classification
from retab.types.extractions import Extraction
from retab.types.partitions import Partition
from retab.types.splits import Split


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
        }
    )

    assert classification.consensus.choices == []


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
        }
    )

    assert extraction.consensus.choices == []


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

    assert partition.consensus.choices == []
