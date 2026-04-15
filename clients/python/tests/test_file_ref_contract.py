from retab.types.mime import FileRef
from retab.types.workflows import FileRef as WorkflowFileRef


def test_file_ref_ignores_internal_storage_fields() -> None:
    file_ref = FileRef.model_validate(
        {
            "id": "file_1",
            "filename": "invoice.pdf",
            "mime_type": "application/pdf",
            "gcs_path_override": "org_1/workflow_node_experiment_inputs/uploads/u1/invoice.pdf",
            "organization_id": "org_1",
        }
    )

    assert file_ref.model_dump() == {
        "id": "file_1",
        "filename": "invoice.pdf",
        "mime_type": "application/pdf",
    }


def test_file_ref_json_schema_exposes_only_public_fields() -> None:
    schema = FileRef.model_json_schema()

    assert set(schema["properties"]) == {"id", "filename", "mime_type"}
    assert set(schema["required"]) == {"id", "filename", "mime_type"}


def test_workflows_module_exports_canonical_file_ref() -> None:
    assert WorkflowFileRef is FileRef
