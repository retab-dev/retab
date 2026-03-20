import copy
import warnings
from typing import Any

from retab import Retab
from retab.resources.evals.classify import Classify
from retab.resources.evals.classify.datasets import Datasets as ClassifyDatasets
from retab.resources.evals.classify.iterations import Iterations as ClassifyIterations
from retab.resources.evals.classify.templates import Templates as ClassifyTemplates
from retab.resources.evals.extract import Extract
from retab.resources.evals.extract.datasets import Datasets as ExtractDatasets
from retab.resources.evals.extract.iterations import Iterations as ExtractIterations
from retab.resources.evals.extract.templates import Templates as ExtractTemplates
from retab.resources.evals.split import Split
from retab.resources.evals.split.datasets import Datasets as SplitDatasets
from retab.resources.evals.split.iterations import Iterations as SplitIterations
from retab.resources.evals.split.templates import Templates as SplitTemplates
from retab.types.inference_settings import InferenceSettings
from retab.types.mime import MIMEData
from retab.types.projects.predictions import PredictionData
from retab.types.projects.iterations import SchemaOverrides
from retab.types.standards import PreparedRequest


def _mime_document() -> MIMEData:
    return MIMEData(filename="test.txt", url="data:text/plain;base64,aGVsbG8=")


def _build_inference_settings(**overrides: Any) -> dict[str, Any]:
    settings = {
        "model": "retab-small",
        "image_resolution_dpi": 192,
        "n_consensus": 1,
    }
    settings.update(overrides)
    return settings


def _build_parsed_completion(status: str = "ok") -> dict[str, Any]:
    return {
        "id": "chatcmpl-1",
        "object": "chat.completion",
        "created": 1710000000,
        "model": "retab-small",
        "choices": [
            {
                "index": 0,
                "finish_reason": "stop",
                "message": {
                    "role": "assistant",
                    "content": f'{{"status":"{status}"}}',
                    "parsed": {"status": status},
                },
            }
        ],
        "extraction_id": "ext-1",
    }


def _build_stream_chunk(status: str = "streaming") -> dict[str, Any]:
    return {
        "id": "chatcmpl-1",
        "object": "chat.completion.chunk",
        "created": 1710000000,
        "model": "retab-small",
        "choices": [
            {
                "index": 0,
                "finish_reason": None,
                "delta": {
                    "role": "assistant",
                    "content": f'{{"status":"{status}"}}',
                    "flat_parsed": {"status": status},
                    "flat_likelihoods": {},
                    "flat_deleted_keys": [],
                    "is_valid_json": True,
                },
            }
        ],
        "extraction_id": "ext-1",
    }


def _build_extract_eval(
    name: str = "Extract Eval",
    *,
    json_schema: dict[str, Any] | None = None,
    is_published: bool = False,
) -> dict[str, Any]:
    schema = json_schema or {"type": "object", "properties": {"status": {"type": "string"}}}
    return {
        "id": "eval_123",
        "name": name,
        "updated_at": "2026-03-18T00:00:00Z",
        "published_config": {
            "inference_settings": _build_inference_settings(),
            "json_schema": schema,
            "origin": "manual",
        },
        "draft_config": {
            "inference_settings": _build_inference_settings(),
            "json_schema": schema,
        },
        "is_published": is_published,
        "is_schema_generated": True,
    }


def _build_extract_dataset(name: str = "Dataset") -> dict[str, Any]:
    return {
        "id": "dataset_456",
        "name": name,
        "updated_at": "2026-03-18T00:00:00Z",
        "base_json_schema": {"type": "object", "properties": {"status": {"type": "string"}}},
        "project_id": "eval_123",
    }


def _build_response_mime_data(filename: str = "test.txt") -> dict[str, Any]:
    return {
        "id": "file_123",
        "filename": filename,
        "mime_type": "text/plain",
    }


def _build_dataset_document(document_id: str = "doc_789", extraction_id: str = "ext-1") -> dict[str, Any]:
    return {
        "id": document_id,
        "updated_at": "2026-03-18T00:00:00Z",
        "project_id": "eval_123",
        "dataset_id": "dataset_456",
        "mime_data": _build_response_mime_data(),
        "prediction_data": {"prediction": {"status": "ok"}},
        "extraction_id": extraction_id,
        "validation_flags": {"status": "reviewed"},
    }


def _build_extract_iteration(
    *,
    iteration_id: str = "iter_999",
    status: str = "draft",
    schema_overrides: dict[str, Any] | None = None,
    draft_schema_overrides: dict[str, Any] | None = None,
) -> dict[str, Any]:
    return {
        "id": iteration_id,
        "updated_at": "2026-03-18T00:00:00Z",
        "inference_settings": _build_inference_settings(),
        "schema_overrides": schema_overrides or {},
        "parent_id": None,
        "project_id": "eval_123",
        "dataset_id": "dataset_456",
        "draft": {
            "schema_overrides": draft_schema_overrides if draft_schema_overrides is not None else schema_overrides or {},
            "updated_at": "2026-03-18T00:00:00Z",
            "inference_settings": _build_inference_settings(),
        },
        "status": status,
    }


def _build_iteration_document(document_id: str = "iter_doc_111", extraction_id: str = "ext-2") -> dict[str, Any]:
    return {
        "id": document_id,
        "updated_at": "2026-03-18T00:00:00Z",
        "project_id": "eval_123",
        "iteration_id": "iter_999",
        "dataset_id": "dataset_456",
        "dataset_document_id": "doc_789",
        "mime_data": _build_response_mime_data(),
        "prediction_data": {"prediction": {"status": "iteration"}},
        "extraction_id": extraction_id,
    }


def _build_metrics() -> dict[str, Any]:
    return {
        "overall_metrics": {
            "accuracy": 0.9,
            "similarity": 0.95,
            "total_error_rate": 0.1,
            "true_positive_rate": 0.85,
            "true_negative_rate": 1.0,
            "false_positive_rate": 0.0,
            "false_negative_rate": 0.15,
            "mismatched_value_rate": 0.05,
            "accuracy_per_field": {"status": 0.9},
            "similarity_per_field": {"status": 0.95},
            "total_documents": 1,
            "total_fields_compared": 1,
        },
        "document_metrics": [
            {
                "document_id": "iter_doc_111",
                "filename": "test.txt",
                "true_positives": [],
                "true_negatives": [],
                "false_positives": [],
                "false_negatives": [],
                "mismatched_values": [],
                "field_similarities": {"status": 0.95},
                "key_mappings": {"status": "status"},
            }
        ],
    }


def _build_split_eval(name: str = "Split Eval", is_published: bool = False) -> dict[str, Any]:
    split_config = [{"name": "Invoice", "description": "Invoice pages"}]
    return {
        "id": "split_eval_123",
        "name": name,
        "updated_at": "2026-03-18T00:00:00Z",
        "published_config": {
            "inference_settings": _build_inference_settings(),
            "split_config": split_config,
            "json_schema": {},
            "subdocuments": split_config,
            "origin": "manual",
        },
        "draft_config": {
            "inference_settings": _build_inference_settings(),
            "split_config": split_config,
            "json_schema": {},
            "subdocuments": split_config,
        },
        "is_published": is_published,
    }


def _build_split_dataset(name: str = "Split Dataset") -> dict[str, Any]:
    split_config = [{"name": "Invoice", "description": "Invoice pages"}]
    return {
        "id": "split_dataset_456",
        "name": name,
        "updated_at": "2026-03-18T00:00:00Z",
        "base_split_config": split_config,
        "base_json_schema": {},
        "base_subdocuments": split_config,
        "base_inference_settings": _build_inference_settings(),
        "project_id": "split_eval_123",
    }


def _build_split_iteration(status: str = "draft") -> dict[str, Any]:
    overrides = {"descriptions_override": {"Invoice": "Updated invoice"}}
    return {
        "id": "split_iter_789",
        "updated_at": "2026-03-18T00:00:00Z",
        "inference_settings": _build_inference_settings(),
        "split_config_overrides": overrides,
        "parent_id": None,
        "project_id": "split_eval_123",
        "dataset_id": "split_dataset_456",
        "draft": {
            "split_config_overrides": overrides,
            "updated_at": "2026-03-18T00:00:00Z",
            "inference_settings": _build_inference_settings(),
        },
        "status": status,
    }


def _build_classify_eval(name: str = "Classify Eval", is_published: bool = False) -> dict[str, Any]:
    categories = [{"name": "invoice", "description": "Invoice document"}]
    return {
        "id": "classify_eval_123",
        "name": name,
        "updated_at": "2026-03-18T00:00:00Z",
        "published_config": {
            "inference_settings": _build_inference_settings(),
            "categories": categories,
            "origin": "manual",
        },
        "draft_config": {
            "inference_settings": _build_inference_settings(),
            "categories": categories,
        },
        "is_published": is_published,
    }


def _build_classify_dataset(name: str = "Classify Dataset") -> dict[str, Any]:
    return {
        "id": "classify_dataset_456",
        "name": name,
        "updated_at": "2026-03-18T00:00:00Z",
        "base_categories": [{"name": "invoice", "description": "Invoice document"}],
        "base_inference_settings": _build_inference_settings(),
        "project_id": "classify_eval_123",
    }


def _build_classify_iteration(status: str = "draft") -> dict[str, Any]:
    overrides = {"description_overrides": {"invoice": "Updated description"}}
    return {
        "id": "classify_iter_789",
        "updated_at": "2026-03-18T00:00:00Z",
        "inference_settings": _build_inference_settings(),
        "category_overrides": overrides,
        "parent_id": None,
        "project_id": "classify_eval_123",
        "dataset_id": "classify_dataset_456",
        "draft": {
            "category_overrides": overrides,
            "updated_at": "2026-03-18T00:00:00Z",
            "inference_settings": _build_inference_settings(),
        },
        "status": status,
    }


class MockPreparedClient:
    def __init__(
        self,
        *,
        responses: dict[tuple[str, str], Any] | None = None,
        stream_responses: dict[tuple[str, str], list[dict[str, Any]]] | None = None,
    ) -> None:
        self.responses = responses or {}
        self.stream_responses = stream_responses or {}
        self.requests: list[PreparedRequest] = []
        self.stream_requests: list[PreparedRequest] = []

    def _prepared_request(self, request: PreparedRequest) -> Any:
        self.requests.append(request)
        key = (request.method, request.url)
        if key not in self.responses:
            raise AssertionError(f"Unhandled mock request: {request.method} {request.url}")
        return copy.deepcopy(self.responses[key])

    def _prepared_request_stream(self, request: PreparedRequest):
        self.stream_requests.append(request)
        key = (request.method, request.url)
        if key not in self.stream_responses:
            raise AssertionError(f"Unhandled mock stream request: {request.method} {request.url}")
        for item in copy.deepcopy(self.stream_responses[key]):
            yield item


def _find_request(client: MockPreparedClient, method: str, url: str) -> PreparedRequest:
    for request in client.requests:
        if request.method == method and request.url == url:
            return request
    raise AssertionError(f"Missing recorded request: {method} {url}")


def test_client_exposes_eval_resources() -> None:
    client = Retab(api_key="test-key", base_url="https://example.com")
    try:
        assert hasattr(client.evals, "extract")
        assert hasattr(client.evals, "split")
        assert hasattr(client.evals, "classify")
        assert hasattr(client.evals.split, "datasets")
        assert hasattr(client.evals.classify, "datasets")
        assert hasattr(client.evals.split.datasets, "iterations")
        assert hasattr(client.evals.classify.datasets, "iterations")
        assert hasattr(client.evals.extract, "templates")
        assert hasattr(client.evals.split, "templates")
        assert hasattr(client.evals.classify, "templates")
    finally:
        client.close()


def test_extract_prepare_process_uses_eval_paths() -> None:
    resource = Extract(client=object())

    request = resource.prepare_process(
        eval_id="eval_123",
        iteration_id="iter_456",
        document=_mime_document(),
        model="retab-small",
    )

    assert request.method == "POST"
    assert request.url == "/evals/extract/extract/eval_123/iter_456"
    assert isinstance(request.files, dict)
    assert "document" in request.files
    assert request.form_data == {"model": "retab-small"}


def test_split_prepare_process_uses_eval_paths() -> None:
    resource = Split(client=object())

    request = resource.prepare_process(
        eval_id="eval_123",
        document=_mime_document(),
        n_consensus=3,
        metadata={"source": "sdk-test"},
    )

    assert request.method == "POST"
    assert request.url == "/evals/split/extract/eval_123"
    assert request.form_data == {
        "n_consensus": 3,
        "metadata": '{"source": "sdk-test"}',
    }


def test_split_prepare_create_uses_split_config_shape() -> None:
    resource = Split(client=object())

    request = resource.prepare_create(
        name="Split eval",
        split_config=[{"name": "Invoice", "description": "invoice pages"}],
    )

    assert request.url == "/evals/split"
    assert request.data["split_config"][0]["name"] == "Invoice"
    assert "json_schema" not in request.data


def test_split_dataset_prepare_paths() -> None:
    resource = SplitDatasets(client=object())

    create_request = resource.prepare_create(
        eval_id="eval_123",
        name="Invoices",
        base_split_config=[{"name": "Invoice", "description": "Invoice pages"}],
    )
    process_request = resource.prepare_process_document(
        eval_id="eval_123",
        dataset_id="dataset_456",
        document_id="doc_789",
    )

    assert create_request.url == "/evals/split/eval_123/datasets"
    assert create_request.data["base_split_config"][0]["name"] == "Invoice"
    assert process_request.url == "/evals/split/eval_123/datasets/dataset_456/dataset-documents/doc_789/process"


def test_classify_prepare_process_uses_eval_paths() -> None:
    resource = Classify(client=object())

    request = resource.prepare_process(
        eval_id="eval_123",
        iteration_id="iter_456",
        document=_mime_document(),
        model="retab-small",
    )

    assert request.method == "POST"
    assert request.url == "/evals/classify/extract/eval_123/iter_456"
    assert request.form_data == {"model": "retab-small"}


def test_classify_prepare_create_uses_categories_shape() -> None:
    resource = Classify(client=object())

    request = resource.prepare_create(
        name="Classify eval",
        categories=[{"name": "invoice", "description": "invoice docs"}],
    )

    assert request.url == "/evals/classify"
    assert request.data["categories"][0]["name"] == "invoice"
    assert "json_schema" not in request.data


def test_extract_dataset_prepare_paths_use_eval_id() -> None:
    resource = ExtractDatasets(client=object())

    create_request = resource.prepare_create(
        eval_id="eval_123",
        name="Invoices",
        base_json_schema={"type": "object"},
    )
    process_request = resource.prepare_process_document(
        eval_id="eval_123",
        dataset_id="dataset_456",
        document_id="doc_789",
    )

    assert create_request.url == "/evals/extract/eval_123/datasets"
    assert create_request.data["base_json_schema"] == {"type": "object"}
    assert process_request.url == "/evals/extract/eval_123/datasets/dataset_456/dataset-documents/doc_789/process"


def test_extract_iteration_prepare_paths_use_eval_id() -> None:
    resource = ExtractIterations(client=object())

    create_request = resource.prepare_create(
        eval_id="eval_123",
        dataset_id="dataset_456",
    )
    schema_request = resource.prepare_get_schema(
        eval_id="eval_123",
        dataset_id="dataset_456",
        iteration_id="iter_789",
        use_draft=True,
    )
    process_request = resource.prepare_process_document(
        eval_id="eval_123",
        dataset_id="dataset_456",
        iteration_id="iter_789",
        document_id="doc_999",
    )

    assert create_request.url == "/evals/extract/eval_123/datasets/dataset_456/iterations"
    assert schema_request.url == "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_789/schema"
    assert schema_request.params == {"use_draft": True}
    assert process_request.url == "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_789/iteration-documents/doc_999/process"


def test_extract_prepare_process_stream_uses_stream_path() -> None:
    resource = Extract(client=object())

    request = resource.prepare_process_stream(
        eval_id="eval_123",
        iteration_id="iter_456",
        document=_mime_document(),
        model="retab-small",
    )

    assert request.method == "POST"
    assert request.url == "/evals/extract/extract/eval_123/iter_456/stream"


def test_extract_prepare_process_only_accepts_single_document() -> None:
    resource = Extract(client=object())

    try:
        resource.prepare_process(  # type: ignore[call-arg]
            eval_id="eval_123",
            documents=[_mime_document()],
        )
    except TypeError:
        pass
    else:
        assert False, "prepare_process should reject 'documents'"


def test_split_iteration_prepare_paths() -> None:
    resource = SplitIterations(client=object())

    create_request = resource.prepare_create(
        eval_id="eval_123",
        dataset_id="dataset_456",
    )
    schema_request = resource.prepare_get_schema(
        eval_id="eval_123",
        dataset_id="dataset_456",
        iteration_id="iter_789",
        use_draft=True,
    )
    process_request = resource.prepare_process_document(
        eval_id="eval_123",
        dataset_id="dataset_456",
        iteration_id="iter_789",
        document_id="doc_999",
    )

    assert create_request.url == "/evals/split/eval_123/datasets/dataset_456/iterations"
    assert schema_request.url == "/evals/split/eval_123/datasets/dataset_456/iterations/iter_789/schema"
    assert schema_request.params == {"use_draft": True}
    assert process_request.url == "/evals/split/eval_123/datasets/dataset_456/iterations/iter_789/iteration-documents/doc_999/process"


def test_classify_dataset_prepare_paths() -> None:
    resource = ClassifyDatasets(client=object())

    create_request = resource.prepare_create(
        eval_id="eval_123",
        name="Invoices",
        base_categories=[{"name": "invoice", "description": "invoice docs"}],
    )
    process_request = resource.prepare_process_document(
        eval_id="eval_123",
        dataset_id="dataset_456",
        document_id="doc_789",
    )

    assert create_request.url == "/evals/classify/eval_123/datasets"
    assert create_request.data["base_categories"][0]["name"] == "invoice"
    assert process_request.url == "/evals/classify/eval_123/datasets/dataset_456/dataset-documents/doc_789/process"


def test_classify_iteration_prepare_paths() -> None:
    resource = ClassifyIterations(client=object())

    create_request = resource.prepare_create(
        eval_id="eval_123",
        dataset_id="dataset_456",
    )
    categories_request = resource.prepare_get_categories(
        eval_id="eval_123",
        dataset_id="dataset_456",
        iteration_id="iter_789",
    )
    process_request = resource.prepare_process_document(
        eval_id="eval_123",
        dataset_id="dataset_456",
        iteration_id="iter_789",
        document_id="doc_999",
    )

    assert create_request.url == "/evals/classify/eval_123/datasets/dataset_456/iterations"
    assert categories_request.url == "/evals/classify/eval_123/datasets/dataset_456/iterations/iter_789/categories"
    assert process_request.url == "/evals/classify/eval_123/datasets/dataset_456/iterations/iter_789/iteration-documents/doc_999/process"


def test_extract_templates_prepare_paths() -> None:
    resource = ExtractTemplates(client=object())

    list_request = resource.prepare_list(limit=5)
    preview_request = resource.prepare_list_builder_document_previews(["tpl_1", "tpl_2"])
    clone_request = resource.prepare_clone("tpl_1", name="Cloned")

    assert list_request.url == "/evals/extract/templates"
    assert list_request.params == {"limit": 5, "order": "desc"}
    assert preview_request.url == "/evals/extract/templates/builder-documents/previews"
    assert preview_request.params == {"template_ids": "tpl_1,tpl_2"}
    assert clone_request.url == "/evals/extract/templates/tpl_1/clone"
    assert clone_request.data == {"name": "Cloned"}


def test_split_templates_prepare_paths() -> None:
    resource = SplitTemplates(client=object())

    get_request = resource.prepare_get("tpl_1")
    docs_request = resource.prepare_list_builder_documents("tpl_1")

    assert get_request.url == "/evals/split/templates/tpl_1"
    assert docs_request.url == "/evals/split/templates/tpl_1/builder-documents"


def test_classify_templates_prepare_paths() -> None:
    resource = ClassifyTemplates(client=object())

    get_request = resource.prepare_get("tpl_1")
    docs_request = resource.prepare_list_builder_documents("tpl_1")

    assert get_request.url == "/evals/classify/templates/tpl_1"
    assert docs_request.url == "/evals/classify/templates/tpl_1/builder-documents"


def test_extract_alias_warns_and_forwards_to_process() -> None:
    resource = Extract(client=object())
    captured: dict[str, object] = {}

    def fake_process(**kwargs: object) -> str:
        captured.update(kwargs)
        return "ok"

    resource.process = fake_process  # type: ignore[method-assign]

    with warnings.catch_warnings(record=True) as caught:
        warnings.simplefilter("always")
        result = resource.extract(project_id="eval_123", document=_mime_document())

    assert result == "ok"
    assert captured["eval_id"] == "eval_123"
    assert any(item.category is DeprecationWarning for item in caught)


def test_extract_lifecycle_sync_methods_parse_models_and_requests() -> None:
    client = MockPreparedClient(
        responses={
            ("POST", "/evals/extract"): _build_extract_eval(),
            ("GET", "/evals/extract/eval_123"): _build_extract_eval(),
            ("GET", "/evals/extract"): {"data": [_build_extract_eval()]},
            ("PATCH", "/evals/extract/eval_123"): _build_extract_eval(name="Extract Eval v2"),
            ("POST", "/evals/extract/eval_123/publish"): _build_extract_eval(is_published=True),
            ("POST", "/evals/extract/extract/eval_123"): _build_parsed_completion(),
            ("POST", "/evals/extract/eval_123/datasets"): _build_extract_dataset(),
            ("GET", "/evals/extract/eval_123/datasets/dataset_456"): _build_extract_dataset(),
            ("GET", "/evals/extract/eval_123/datasets"): {"data": [_build_extract_dataset()]},
            ("PATCH", "/evals/extract/eval_123/datasets/dataset_456"): _build_extract_dataset(name="Dataset v2"),
            ("POST", "/evals/extract/eval_123/datasets/dataset_456/duplicate"): _build_extract_dataset(name="Dataset Copy"),
            ("POST", "/evals/extract/eval_123/datasets/dataset_456/dataset-documents"): _build_dataset_document(),
            ("GET", "/evals/extract/eval_123/datasets/dataset_456/dataset-documents/doc_789"): _build_dataset_document(),
            ("GET", "/evals/extract/eval_123/datasets/dataset_456/dataset-documents"): [_build_dataset_document()],
            ("PATCH", "/evals/extract/eval_123/datasets/dataset_456/dataset-documents/doc_789"): _build_dataset_document(extraction_id="ext-updated"),
            ("POST", "/evals/extract/eval_123/datasets/dataset_456/dataset-documents/doc_789/process"): _build_parsed_completion(),
            ("DELETE", "/evals/extract/eval_123/datasets/dataset_456/dataset-documents/doc_789"): {"success": True, "id": "doc_789"},
            ("POST", "/evals/extract/eval_123/datasets/dataset_456/iterations"): _build_extract_iteration(),
            ("GET", "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999"): _build_extract_iteration(),
            ("GET", "/evals/extract/eval_123/datasets/dataset_456/iterations"): {"data": [_build_extract_iteration()]},
            (
                "PATCH",
                "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999",
            ): _build_extract_iteration(
                draft_schema_overrides={"descriptionsOverride": {"status": "Review status"}}
            ),
            ("GET", "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/schema"): {
                "type": "object",
                "properties": {"status": {"type": "string", "description": "Review status"}},
            },
            (
                "POST",
                "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/iteration-documents/processDocumentsFromDatasetId",
            ): {"success": True, "id": "iter_doc_111"},
            (
                "GET",
                "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/iteration-documents/iter_doc_111",
            ): _build_iteration_document(),
            (
                "GET",
                "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/iteration-documents",
            ): [_build_iteration_document()],
            (
                "PATCH",
                "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/iteration-documents/iter_doc_111",
            ): _build_iteration_document(extraction_id="ext-doc-updated"),
            ("GET", "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/metrics"): _build_metrics(),
            (
                "POST",
                "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/iteration-documents/iter_doc_111/process",
            ): _build_parsed_completion(status="iteration"),
            ("POST", "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/finalize"): _build_extract_iteration(status="completed"),
            (
                "DELETE",
                "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/iteration-documents/iter_doc_111",
            ): {"success": True, "id": "iter_doc_111"},
            ("DELETE", "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999"): {"success": True, "id": "iter_999"},
            ("DELETE", "/evals/extract/eval_123/datasets/dataset_456"): {"success": True, "id": "dataset_456"},
            ("DELETE", "/evals/extract/eval_123"): {"success": True, "id": "eval_123"},
        }
    )
    resource = Extract(client=client)

    created_eval = resource.create(name="Extract Eval", json_schema={"type": "object", "properties": {"status": {"type": "string"}}})
    fetched_eval = resource.get("eval_123")
    listed_evals = resource.list(limit=1)
    updated_eval = resource.update("eval_123", name="Extract Eval v2", json_schema={"type": "object"})
    published_eval = resource.publish("eval_123", origin="iter_999")
    processed_eval = resource.process(
        eval_id="eval_123",
        document=_mime_document(),
        model="retab-small",
        metadata={"source": "lifecycle"},
    )

    created_dataset = resource.datasets.create("eval_123", name="Dataset", base_json_schema={"type": "object"})
    fetched_dataset = resource.datasets.get("eval_123", "dataset_456")
    listed_datasets = resource.datasets.list("eval_123", limit=1)
    updated_dataset = resource.datasets.update("eval_123", "dataset_456", name="Dataset v2")
    duplicated_dataset = resource.datasets.duplicate("eval_123", "dataset_456", name="Dataset Copy")
    created_dataset_document = resource.datasets.add_document(
        "eval_123",
        "dataset_456",
        mime_data=_mime_document(),
        prediction_data=PredictionData(prediction={"status": "ok"}),
    )
    fetched_dataset_document = resource.datasets.get_document("eval_123", "dataset_456", "doc_789")
    listed_dataset_documents = resource.datasets.list_documents("eval_123", "dataset_456")
    updated_dataset_document = resource.datasets.update_document(
        "eval_123",
        "dataset_456",
        "doc_789",
        validation_flags={"status": "reviewed"},
        extraction_id="ext-updated",
    )
    processed_dataset_document = resource.datasets.process_document("eval_123", "dataset_456", "doc_789")
    deleted_dataset_document = resource.datasets.delete_document("eval_123", "dataset_456", "doc_789")

    created_iteration = resource.datasets.iterations.create("eval_123", "dataset_456")
    fetched_iteration = resource.datasets.iterations.get("eval_123", "dataset_456", "iter_999")
    listed_iterations = resource.datasets.iterations.list("eval_123", "dataset_456", limit=1)
    updated_iteration = resource.datasets.iterations.update_draft(
        "eval_123",
        "dataset_456",
        "iter_999",
        inference_settings=InferenceSettings(model="retab-small", image_resolution_dpi=192, n_consensus=1),
        schema_overrides=SchemaOverrides(descriptionsOverride={"status": "Review status"}),
    )
    materialized_schema = resource.datasets.iterations.get_schema("eval_123", "dataset_456", "iter_999", use_draft=True)
    process_documents_result = resource.datasets.iterations.process_documents("eval_123", "dataset_456", "iter_999", "doc_789")
    fetched_iteration_document = resource.datasets.iterations.get_document("eval_123", "dataset_456", "iter_999", "iter_doc_111")
    listed_iteration_documents = resource.datasets.iterations.list_documents("eval_123", "dataset_456", "iter_999")
    updated_iteration_document = resource.datasets.iterations.update_document(
        "eval_123",
        "dataset_456",
        "iter_999",
        "iter_doc_111",
        prediction_data=PredictionData(prediction={"status": "iteration"}),
        extraction_id="ext-doc-updated",
    )
    metrics = resource.datasets.iterations.get_metrics("eval_123", "dataset_456", "iter_999", force_refresh=True)
    processed_iteration_document = resource.datasets.iterations.process_document("eval_123", "dataset_456", "iter_999", "iter_doc_111")
    finalized_iteration = resource.datasets.iterations.finalize("eval_123", "dataset_456", "iter_999")
    deleted_iteration_document = resource.datasets.iterations.delete_document("eval_123", "dataset_456", "iter_999", "iter_doc_111")
    deleted_iteration = resource.datasets.iterations.delete("eval_123", "dataset_456", "iter_999")
    deleted_dataset = resource.datasets.delete("eval_123", "dataset_456")
    deleted_eval = resource.delete("eval_123")

    assert created_eval.id == "eval_123"
    assert fetched_eval.name == "Extract Eval"
    assert listed_evals[0].draft_config.json_schema["type"] == "object"
    assert updated_eval.name == "Extract Eval v2"
    assert published_eval.is_published is True
    assert processed_eval.data == {"status": "ok"}

    assert created_dataset.project_id == "eval_123"
    assert fetched_dataset.id == "dataset_456"
    assert listed_datasets[0].name == "Dataset"
    assert updated_dataset.name == "Dataset v2"
    assert duplicated_dataset.name == "Dataset Copy"
    assert created_dataset_document.extraction_id == "ext-1"
    assert fetched_dataset_document.id == "doc_789"
    assert listed_dataset_documents[0].validation_flags["status"] == "reviewed"
    assert updated_dataset_document.extraction_id == "ext-updated"
    assert processed_dataset_document.data == {"status": "ok"}
    assert deleted_dataset_document["id"] == "doc_789"

    assert created_iteration.project_id == "eval_123"
    assert fetched_iteration.id == "iter_999"
    assert listed_iterations[0].status == "draft"
    assert updated_iteration.draft.schema_overrides.descriptionsOverride == {"status": "Review status"}
    assert materialized_schema["properties"]["status"]["description"] == "Review status"
    assert process_documents_result["success"] is True
    assert fetched_iteration_document.dataset_document_id == "doc_789"
    assert listed_iteration_documents[0].iteration_id == "iter_999"
    assert updated_iteration_document.extraction_id == "ext-doc-updated"
    assert metrics.overall_metrics.total_documents == 1
    assert processed_iteration_document.data == {"status": "iteration"}
    assert finalized_iteration.status == "finalized"
    assert deleted_iteration_document["id"] == "iter_doc_111"
    assert deleted_iteration["id"] == "iter_999"
    assert deleted_dataset["id"] == "dataset_456"
    assert deleted_eval["id"] == "eval_123"

    assert _find_request(client, "POST", "/evals/extract").data == {
        "name": "Extract Eval",
        "json_schema": {"type": "object", "properties": {"status": {"type": "string"}}},
    }
    assert _find_request(client, "GET", "/evals/extract").params == {"limit": 1}
    assert _find_request(client, "PATCH", "/evals/extract/eval_123").data == {
        "name": "Extract Eval v2",
        "json_schema": {"type": "object"},
    }
    assert _find_request(client, "POST", "/evals/extract/eval_123/publish").params == {"origin": "iter_999"}
    assert _find_request(client, "POST", "/evals/extract/extract/eval_123").form_data == {
        "model": "retab-small",
        "metadata": '{"source": "lifecycle"}',
    }
    assert _find_request(client, "POST", "/evals/extract/eval_123/datasets/dataset_456/dataset-documents").data == {
        "mime_data": _mime_document().model_dump(mode="json"),
        "project_id": "eval_123",
        "dataset_id": "dataset_456",
        "prediction_data": {"prediction": {"status": "ok"}, "metadata": None, "updated_at": None},
    }
    assert _find_request(client, "PATCH", "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999").data["draft"]["schema_overrides"] == {
        "descriptionsOverride": {"status": "Review status"}
    }
    assert _find_request(client, "GET", "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/schema").params == {"use_draft": True}
    assert _find_request(
        client,
        "POST",
        "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/iteration-documents/processDocumentsFromDatasetId",
    ).data == {"dataset_document_id": "doc_789"}
    assert _find_request(client, "GET", "/evals/extract/eval_123/datasets/dataset_456/iterations/iter_999/metrics").params == {"force_refresh": True}


def test_extract_process_stream_parses_chunks() -> None:
    client = MockPreparedClient(
        stream_responses={
            ("POST", "/evals/extract/extract/eval_123/stream"): [
                _build_stream_chunk(),
                {},
            ]
        }
    )
    resource = Extract(client=client)

    with resource.process_stream(eval_id="eval_123", document=_mime_document()) as stream:
        chunks = list(stream)

    assert len(chunks) == 1
    assert chunks[0].choices[0].delta.flat_parsed == {"status": "streaming"}
    assert client.stream_requests[0].url == "/evals/extract/extract/eval_123/stream"


def test_split_and_classify_lifecycle_methods_parse_typed_models() -> None:
    client = MockPreparedClient(
        responses={
            ("POST", "/evals/split"): _build_split_eval(),
            ("GET", "/evals/split/split_eval_123"): _build_split_eval(),
            ("GET", "/evals/split"): {"data": [_build_split_eval()]},
            ("PATCH", "/evals/split/split_eval_123"): _build_split_eval(name="Split Eval v2"),
            ("POST", "/evals/split/split_eval_123/publish"): _build_split_eval(is_published=True),
            ("POST", "/evals/split/split_eval_123/datasets"): _build_split_dataset(),
            ("POST", "/evals/split/split_eval_123/datasets/split_dataset_456/iterations"): _build_split_iteration(),
            ("GET", "/evals/split/split_eval_123/datasets/split_dataset_456/iterations/split_iter_789/schema"): {
                "subdocuments": [{"name": "Invoice", "description": "Updated invoice"}]
            },
            ("POST", "/evals/split/split_eval_123/datasets/split_dataset_456/iterations/split_iter_789/finalize"): _build_split_iteration(status="completed"),
            ("POST", "/evals/classify"): _build_classify_eval(),
            ("GET", "/evals/classify/classify_eval_123"): _build_classify_eval(),
            ("GET", "/evals/classify"): {"data": [_build_classify_eval()]},
            ("PATCH", "/evals/classify/classify_eval_123"): _build_classify_eval(name="Classify Eval v2"),
            ("POST", "/evals/classify/classify_eval_123/publish"): _build_classify_eval(is_published=True),
            ("POST", "/evals/classify/classify_eval_123/datasets"): _build_classify_dataset(),
            ("POST", "/evals/classify/classify_eval_123/datasets/classify_dataset_456/iterations"): _build_classify_iteration(),
            (
                "GET",
                "/evals/classify/classify_eval_123/datasets/classify_dataset_456/iterations/classify_iter_789/categories",
            ): [{"name": "invoice", "description": "Updated description"}],
            (
                "POST",
                "/evals/classify/classify_eval_123/datasets/classify_dataset_456/iterations/classify_iter_789/finalize",
            ): _build_classify_iteration(status="completed"),
        }
    )
    split = Split(client=client)
    classify = Classify(client=client)

    split_eval = split.create(name="Split Eval", split_config=[{"name": "Invoice", "description": "Invoice pages"}])
    split_fetched = split.get("split_eval_123")
    split_listed = split.list(limit=1)
    split_updated = split.update("split_eval_123", name="Split Eval v2")
    split_published = split.publish("split_eval_123", origin="split_iter_789")
    split_dataset = split.datasets.create(
        "split_eval_123",
        name="Split Dataset",
        base_split_config=[{"name": "Invoice", "description": "Invoice pages"}],
    )
    split_iteration = split.datasets.iterations.create("split_eval_123", "split_dataset_456")
    split_schema = split.datasets.iterations.get_schema("split_eval_123", "split_dataset_456", "split_iter_789", use_draft=True)
    split_finalized = split.datasets.iterations.finalize("split_eval_123", "split_dataset_456", "split_iter_789")

    classify_eval = classify.create(name="Classify Eval", categories=[{"name": "invoice", "description": "Invoice document"}])
    classify_fetched = classify.get("classify_eval_123")
    classify_listed = classify.list(limit=1)
    classify_updated = classify.update("classify_eval_123", name="Classify Eval v2")
    classify_published = classify.publish("classify_eval_123", origin="classify_iter_789")
    classify_dataset = classify.datasets.create(
        "classify_eval_123",
        name="Classify Dataset",
        base_categories=[{"name": "invoice", "description": "Invoice document"}],
    )
    classify_iteration = classify.datasets.iterations.create("classify_eval_123", "classify_dataset_456")
    classify_categories = classify.datasets.iterations.get_categories(
        "classify_eval_123",
        "classify_dataset_456",
        "classify_iter_789",
    )
    classify_finalized = classify.datasets.iterations.finalize(
        "classify_eval_123",
        "classify_dataset_456",
        "classify_iter_789",
    )

    assert split_eval.draft_config.split_config[0].name == "Invoice"
    assert split_fetched.id == "split_eval_123"
    assert split_listed[0].published_config.subdocuments[0].name == "Invoice"
    assert split_updated.name == "Split Eval v2"
    assert split_published.is_published is True
    assert split_dataset.base_split_config[0].name == "Invoice"
    assert split_iteration.project_id == "split_eval_123"
    assert split_schema["subdocuments"][0]["description"] == "Updated invoice"
    assert split_finalized.status == "finalized"

    assert classify_eval.draft_config.categories[0].name == "invoice"
    assert classify_fetched.id == "classify_eval_123"
    assert classify_listed[0].published_config.categories[0].name == "invoice"
    assert classify_updated.name == "Classify Eval v2"
    assert classify_published.is_published is True
    assert classify_dataset.base_categories[0].name == "invoice"
    assert classify_iteration.project_id == "classify_eval_123"
    assert classify_categories[0]["description"] == "Updated description"
    assert classify_finalized.status == "finalized"

    assert _find_request(client, "POST", "/evals/split").data == {
        "name": "Split Eval",
        "split_config": [{"name": "Invoice", "description": "Invoice pages"}],
    }
    assert "json_schema" not in _find_request(client, "POST", "/evals/split").data
    assert _find_request(client, "PATCH", "/evals/split/split_eval_123").data == {"name": "Split Eval v2"}
    assert _find_request(client, "POST", "/evals/split/split_eval_123/publish").params == {"origin": "split_iter_789"}
    assert _find_request(client, "POST", "/evals/classify").data == {
        "name": "Classify Eval",
        "categories": [{"name": "invoice", "description": "Invoice document"}],
    }
    assert "json_schema" not in _find_request(client, "POST", "/evals/classify").data
    assert _find_request(client, "PATCH", "/evals/classify/classify_eval_123").data == {"name": "Classify Eval v2"}
    assert _find_request(client, "POST", "/evals/classify/classify_eval_123/publish").params == {"origin": "classify_iter_789"}
