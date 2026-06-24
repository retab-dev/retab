# pyright: reportAttributeAccessIssue=false, reportArgumentType=false, reportCallIssue=false
import inspect
import importlib
from typing import cast

import pytest

from retab.resources.extractions import AsyncExtractions, Extractions
from retab import Retab
from retab.types.extractions import ExtractionRequest
from retab.types.mime import MIMEData
from retab.types.partitions import Partition
from retab.types.splits import Split
from retab.types.standards import PreparedRequest
from mocks import mock_retab
from samples import sample_document

# Whole module is unit (pure offline; no server/credentials needed).
pytestmark = pytest.mark.unit


def test_removed_surfaces_are_not_exposed() -> None:
    with Retab(api_key="test", base_url="http://example.com") as client:
        assert not hasattr(client, "documents")
        assert not hasattr(client, "models")
        assert not hasattr(client, "projects")
        assert not hasattr(client, "evals")
        assert not hasattr(client, "auth")
        assert not hasattr(client, "environments")


@pytest.mark.parametrize(
    "module_name",
    [
        "retab.resources.documents.client",
        "retab.resources.models",
        "retab.resources.projects",
        "retab.resources.evals",
        "retab.resources.auth",
        "retab.resources.environments",
        "retab.types.projects",
        "retab.types.evals",
        "retab.types.auth",
        "retab.types.environments",
    ],
)
def test_projects_and_evals_modules_are_removed(module_name: str) -> None:
    with pytest.raises(ModuleNotFoundError):
        importlib.import_module(module_name)


def _request_params(request: PreparedRequest) -> dict[str, object]:
    assert request.params is not None
    return cast(dict[str, object], request.params)


def test_extractions_create_exposes_stream_argument() -> None:
    assert "stream" in inspect.signature(Extractions.create).parameters
    assert "stream" in inspect.signature(AsyncExtractions.create).parameters
    assert "stream" in inspect.signature(Extractions.prepare_create).parameters
    assert "stream" in ExtractionRequest.model_fields

    request = Extractions(object()).prepare_create(
        document=sample_document(),
        json_schema={"type": "object"},
        stream=True,
    )

    assert request.data is not None
    assert request.data["stream"] is True


def test_files_resource_exposes_generated_upload_endpoints() -> None:
    def fake_prepared_request(request: object) -> dict[str, object]:
        if getattr(request, "url") == "/v1/files/upload":
            return {
                "fileId": "file_123",
                "uploadUrl": "https://storage.googleapis.com/signed-upload",
                "uploadMethod": "PUT",
                "uploadHeaders": {"Content-Type": "application/pdf"},
                "mimeData": {"filename": "invoice.pdf", "url": "https://storage.retab.com/org_1/file_123.pdf"},
                "expiresAt": "2026-04-24T12:00:00Z",
            }
        return {
            "filename": "invoice.pdf",
            "url": "https://storage.retab.com/org_1/file_123.pdf",
        }

    client, rec = mock_retab(fake_prepared_request)
    with client:
        upload = client.files.create_upload(
            filename="invoice.pdf",
            content_type="application/pdf",
            size_bytes=8,
            sha256="e16fa5d9b51928755db85b917f0297babaf22c7a47e97d9212adab56e61ba04e",
        )
        completed = client.files.complete_upload(
            "file_123",
            sha256="e16fa5d9b51928755db85b917f0297babaf22c7a47e97d9212adab56e61ba04e",
        )

    assert upload.file_id == "file_123"
    assert completed.filename == "invoice.pdf"
    requests = rec.requests
    assert [getattr(request, "url") for request in requests] == [
        "/v1/files/upload",
        "/v1/files/upload/file_123/complete",
    ]
    assert getattr(requests[0], "data") == {
        "filename": "invoice.pdf",
        "content_type": "application/pdf",
        "size_bytes": 8,
        "sha256": "e16fa5d9b51928755db85b917f0297babaf22c7a47e97d9212adab56e61ba04e",
    }
    assert getattr(requests[1], "data") == {
        "sha256": "e16fa5d9b51928755db85b917f0297babaf22c7a47e97d9212adab56e61ba04e",
    }


def test_extractions_create_accepts_signed_bucket_url() -> None:
    signed_url = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc"

    def fake_prepared_request(request: object) -> dict[str, object]:
        return {
            "id": "extr_123",
            "file": {
                "id": "file_123",
                "filename": "invoice.pdf",
                "mime_type": "application/pdf",
            },
            "model": "retab-small",
            "json_schema": {"type": "object"},
            "n_consensus": 1,
            "output": {},
            "consensus": {"choices": []},
            "metadata": {},
        }

    client, rec = mock_retab(fake_prepared_request)
    with client:
        result = client.extractions.create(
            document=signed_url,
            json_schema={"type": "object"},
            model="retab-small",
        )

    assert result.id == "extr_123"
    assert getattr(rec.request, "url") == "/v1/extractions"
    assert getattr(rec.request, "data")["document"] == {
        "filename": "invoice.pdf",
        "url": signed_url,
    }


def test_mime_data_id_extracts_file_id_from_canonical_retab_storage_url() -> None:
    mime = MIMEData(
        filename="invoice.pdf",
        url="https://storage.retab.com/org_1/file_123.pdf",
    )

    assert mime.id == "file_123"


@pytest.mark.parametrize(
    ("resource_name", "prepare_name", "kwargs"),
    [
        (
            "classifications",
            "prepare_create",
            {
                "categories": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "parses",
            "prepare_create",
            {
                "model": "retab-small",
            },
        ),
        (
            "splits",
            "prepare_create",
            {
                "subdocuments": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "partitions",
            "prepare_create",
            {
                "key": "invoice_number",
                "instructions": "Split the document into one chunk per invoice number.",
                "model": "retab-small",
            },
        ),
        (
            "extractions",
            "prepare_create",
            {
                "json_schema": {"type": "object"},
                "model": "retab-small",
            },
        ),
    ],
)
def test_resource_create_builders_preserve_signed_bucket_urls(
    resource_name: str,
    prepare_name: str,
    kwargs: dict[str, object],
) -> None:
    signed_url = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc"

    with Retab(api_key="test", base_url="http://example.com") as client:
        resource = getattr(client, resource_name)
        request = getattr(resource, prepare_name)(document=signed_url, **kwargs)

    assert request.data["document"] == {
        "filename": "invoice.pdf",
        "url": signed_url,
    }


@pytest.mark.parametrize(
    ("resource_name", "prepare_name", "kwargs"),
    [
        (
            "classifications",
            "prepare_create",
            {
                "categories": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "parses",
            "prepare_create",
            {
                "model": "retab-small",
            },
        ),
        (
            "splits",
            "prepare_create",
            {
                "subdocuments": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "partitions",
            "prepare_create",
            {
                "key": "invoice_number",
                "instructions": "Split the document into one chunk per invoice number.",
                "model": "retab-small",
            },
        ),
        (
            "extractions",
            "prepare_create",
            {
                "json_schema": {"type": "object"},
                "model": "retab-small",
            },
        ),
    ],
)
def test_resource_create_builders_accept_retab_storage_url_string(
    resource_name: str,
    prepare_name: str,
    kwargs: dict[str, object],
) -> None:
    retab_url = "https://storage.retab.com/org_1/file_123.pdf"

    with Retab(api_key="test", base_url="http://example.com") as client:
        resource = getattr(client, resource_name)
        request = getattr(resource, prepare_name)(document=retab_url, **kwargs)

    assert request.data["document"] == {
        "filename": "file_123.pdf",
        "url": retab_url,
    }


@pytest.mark.parametrize(
    ("resource_name", "prepare_name", "kwargs"),
    [
        (
            "classifications",
            "prepare_create",
            {
                "categories": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "parses",
            "prepare_create",
            {
                "model": "retab-small",
            },
        ),
        (
            "splits",
            "prepare_create",
            {
                "subdocuments": [{"name": "invoice", "description": "Invoice documents"}],
                "model": "retab-small",
            },
        ),
        (
            "partitions",
            "prepare_create",
            {
                "key": "invoice_number",
                "instructions": "Split the document into one chunk per invoice number.",
                "model": "retab-small",
            },
        ),
        (
            "extractions",
            "prepare_create",
            {
                "json_schema": {"type": "object"},
                "model": "retab-small",
            },
        ),
    ],
)
def test_resource_create_builders_accept_uploaded_mime_ref(
    resource_name: str,
    prepare_name: str,
    kwargs: dict[str, object],
) -> None:
    mime_ref = MIMEData(
        filename="invoice.pdf",
        url="https://storage.retab.com/org_1/file_123.pdf",
    )

    with Retab(api_key="test", base_url="http://example.com") as client:
        resource = getattr(client, resource_name)
        request = getattr(resource, prepare_name)(document=mime_ref, **kwargs)

    assert request.data["document"] == {
        "filename": "invoice.pdf",
        "url": "https://storage.retab.com/org_1/file_123.pdf",
    }


@pytest.mark.parametrize(
    "resource_name, prepare_name, kwargs",
    [
        ("splits", "prepare_create", {"subdocuments": [{"name": "invoice", "description": "Invoice documents"}]}),
        ("parses", "prepare_create", {}),
        ("classifications", "prepare_create", {"categories": [{"name": "invoice", "description": "Invoice documents"}]}),
        ("extractions", "prepare_create", {"json_schema": {"type": "object"}}),
    ],
)
def test_resource_create_builders_resolve_file_ref_to_mime_data(
    resource_name: str,
    prepare_name: str,
    kwargs: dict[str, object],
) -> None:
    # A stored file id (FileRef) is no longer serialized to the wire — the
    # document routes accept URL-backed MIMEData only, so it is resolved
    # client-side via the Files download-link endpoint before the body is built.
    from unittest.mock import MagicMock

    from retab.types.mime import FileRef

    client = MagicMock()
    link = MagicMock()
    link.download_url = "https://storage.googleapis.com/signed/file_123.pdf?sig=abc"
    link.filename = "server-name.pdf"
    link.mime_data = MIMEData(
        filename="durable.pdf",
        url="https://storage.retab.com/org_1/file_123.pdf",
    )
    client.files.get_download_link.return_value = link

    resource_cls = type(getattr(Retab(api_key="test", base_url="http://example.com"), resource_name))
    resource = resource_cls(client=client)
    request = getattr(resource, prepare_name)(
        document=FileRef(id="file_123", filename="invoice.pdf", mime_type="application/pdf"),
        **kwargs,
    )

    client.files.get_download_link.assert_called_once_with("file_123")
    assert request.data["document"] == {
        "filename": "invoice.pdf",
        "url": "https://storage.retab.com/org_1/file_123.pdf",
    }


def test_resource_create_builders_include_supported_route_fields() -> None:
    signed_url = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc"

    with Retab(api_key="test", base_url="http://example.com") as client:
        classification = client.classifications.prepare_create(
            document=signed_url,
            categories=[{"name": "invoice", "description": "Invoice documents"}],
            model="retab-small",
            first_n_pages=1,
            instructions="Use the cover page.",
        )
        parse = client.parses.prepare_create(
            document=signed_url,
            model="retab-small",
            instructions="Keep invoice sections separate.",
        )
        split = client.splits.prepare_create(
            document=signed_url,
            subdocuments=[{"name": "invoice", "description": "Invoice documents"}],
            model="retab-small",
            instructions="Split attachments from invoices.",
        )
        extraction = client.extractions.prepare_create(
            document=signed_url,
            json_schema={"type": "object"},
            model="retab-small",
            additional_messages=[{"role": "user", "content": "Use totals only."}],
        )

    assert classification.data["first_n_pages"] == 1
    assert classification.data["instructions"] == "Use the cover page."
    assert parse.data["instructions"] == "Keep invoice sections separate."
    assert split.data["instructions"] == "Split attachments from invoices."
    assert extraction.data["additional_messages"] == [{"role": "user", "content": "Use totals only."}]


def test_resource_list_builders_include_filename_filters() -> None:
    with Retab(api_key="test", base_url="http://example.com") as client:
        assert _request_params(client.classifications.prepare_list(filename="invoice.pdf"))["filename"] == "invoice.pdf"
        assert _request_params(client.parses.prepare_list(filename="invoice.pdf"))["filename"] == "invoice.pdf"
        assert _request_params(client.splits.prepare_list(filename="invoice.pdf"))["filename"] == "invoice.pdf"
        assert _request_params(client.partitions.prepare_list(filename="invoice.pdf"))["filename"] == "invoice.pdf"
        assert _request_params(client.extractions.prepare_list(filename="invoice.pdf"))["filename"] == "invoice.pdf"


def test_partitions_list_uses_resource_route() -> None:
    def fake_prepared_request(request: object) -> dict[str, object]:
        return {"data": [], "list_metadata": {"before": None, "after": None}}

    client, rec = mock_retab(fake_prepared_request)
    with client:
        page = client.partitions.list(limit=5, filename="invoice.pdf")

    assert len(page.data) == 0
    assert getattr(rec.request, "url") == "/v1/partitions"
    assert getattr(rec.request, "params") == {
        "limit": 5,
        "order": "desc",
        "filename": "invoice.pdf",
    }


def test_splits_create_uses_new_resource_route() -> None:
    def fake_prepared_request(request: object) -> dict[str, object]:
        return {
            "id": "split_123",
            "file": {
                "id": "file_123",
                "filename": "invoice.txt",
                "mime_type": "text/plain",
            },
            "model": "retab-small",
            "subdocuments": [
                {"name": "invoice", "description": "Invoice documents"},
            ],
            "output": [
                {"name": "invoice", "pages": [1]},
            ],
            "consensus": {
                "choices": [],
                "likelihoods": None,
            },
            "usage": {
                "page_count": 1,
                "credits": 1.0,
            },
        }

    client, rec = mock_retab(fake_prepared_request)
    with client:
        result = client.splits.create(
            document=sample_document(),
            model="retab-small",
            subdocuments=[
                {
                    "name": "invoice",
                    "description": "Invoice documents",
                    "allow_multiple_instances": True,
                }
            ],
        )

    assert isinstance(result, Split)
    assert result.id == "split_123"
    assert getattr(rec.request, "url") == "/v1/splits"
    # Subdocument's canonical schema is {name, description, allow_multiple_instances}.
    # The route rejects legacy ``partition_key`` outright
    # (see backend/main_server/.../splits/unit_tests/test_splits_routes.py::
    # test_create_split_route_rejects_partition_key_subdocuments).
    assert getattr(rec.request, "data")["subdocuments"] == [
        {
            "name": "invoice",
            "description": "Invoice documents",
            "allow_multiple_instances": True,
        }
    ]


def test_partitions_create_uses_new_resource_route() -> None:
    def fake_prepared_request(request: object) -> dict[str, object]:
        return {
            "id": "prtn_123",
            "file": {
                "id": "file_123",
                "filename": "invoice.txt",
                "mime_type": "text/plain",
            },
            "model": "retab-small",
            "key": "invoice_number",
            "instructions": "Split the document into one chunk per invoice number.",
            "n_consensus": 3,
            "output": [
                {
                    "key": "INV-001",
                    "pages": [1, 2],
                }
            ],
            "consensus": {
                "choices": [],
                "likelihoods": None,
            },
            "usage": {
                "credits": 1.0,
            },
        }

    client, rec = mock_retab(fake_prepared_request)
    with client:
        result = client.partitions.create(
            document=sample_document(),
            key="invoice_number",
            instructions="Split the document into one chunk per invoice number.",
            model="retab-small",
            n_consensus=3,
            allow_overlap=True,
        )

    assert isinstance(result, Partition)
    assert result.id == "prtn_123"
    assert result.output is not None
    assert result.output[0].key == "INV-001"
    assert result.consensus is not None
    assert result.consensus.choices == []
    assert result.usage is not None
    assert result.usage.credits == 1.0
    assert getattr(rec.request, "url") == "/v1/partitions"
    assert getattr(rec.request, "data") == {
        "document": {
            "filename": "invoice.txt",
            "url": "data:text/plain;base64,aW52b2ljZQ==",
        },
        "key": "invoice_number",
        "instructions": "Split the document into one chunk per invoice number.",
        "model": "retab-small",
        "n_consensus": 3,
        "allow_overlap": True,
        "bust_cache": False,
        "background": False,
    }


def test_partitions_create_defaults_allow_overlap_to_true() -> None:
    def fake_prepared_request(request: object) -> dict[str, object]:
        return {
            "id": "prtn_123",
            "file": {
                "id": "file_123",
                "filename": "invoice.txt",
                "mime_type": "text/plain",
            },
            "model": "retab-small",
            "key": "invoice_number",
            "instructions": "Split the document into one chunk per invoice number.",
            "n_consensus": 3,
            "output": [],
            "consensus": {
                "choices": [],
                "likelihoods": None,
            },
            "usage": None,
        }

    client, rec = mock_retab(fake_prepared_request)
    with client:
        result = client.partitions.create(
            document=sample_document(),
            key="invoice_number",
            instructions="Split the document into one chunk per invoice number.",
            model="retab-small",
            n_consensus=3,
        )

    assert result.allow_overlap is True
    assert getattr(rec.request, "data")["allow_overlap"] is True
