from typing import Any

from retab import Retab
from retab.resources.evals.classify import Classify
from retab.resources.evals.extract import Extract
from retab.resources.evals.split import Split
from retab.types.mime import MIMEData
from retab.types.standards import PreparedRequest


def _mime_document() -> MIMEData:
    return MIMEData(filename="test.txt", url="data:text/plain;base64,aGVsbG8=")


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


def _build_classify_response(label: str = "invoice") -> dict[str, Any]:
    return {
        "result": {
            "reasoning": "Detected invoice keywords",
            "classification": label,
        },
        "usage": {
            "credits": 0.12,
        },
    }


class MockPreparedClient:
    def __init__(
        self,
        responses: dict[str, Any] | None = None,
        stream_responses: dict[str, list[dict[str, Any]]] | None = None,
    ) -> None:
        self.responses = responses or {}
        self.stream_responses = stream_responses or {}
        self.requests: list[PreparedRequest] = []

    def _prepared_request(self, request: PreparedRequest) -> Any:
        self.requests.append(request)
        return self.responses[request.url]

    def _prepared_request_stream(self, request: PreparedRequest):
        self.requests.append(request)
        for item in self.stream_responses.get(request.url, []):
            yield item


def test_client_exposes_only_eval_processing_resources() -> None:
    client = Retab(api_key="test")

    assert hasattr(client.evals, "extract")
    assert hasattr(client.evals, "split")
    assert hasattr(client.evals, "classify")

    assert not hasattr(client.evals.extract, "datasets")
    assert not hasattr(client.evals.extract, "templates")
    assert not hasattr(client.evals.extract, "create")
    assert not hasattr(client.evals.extract, "extract")

    assert not hasattr(client.evals.split, "datasets")
    assert not hasattr(client.evals.split, "templates")
    assert not hasattr(client.evals.split, "create")
    assert not hasattr(client.evals.split, "process_stream")

    assert not hasattr(client.evals.classify, "datasets")
    assert not hasattr(client.evals.classify, "templates")
    assert not hasattr(client.evals.classify, "create")
    assert not hasattr(client.evals.classify, "process_stream")


def test_extract_prepare_process_paths() -> None:
    resource = Extract(client=object())

    base_request = resource.prepare_process(
        eval_id="eval_123",
        document=_mime_document(),
        metadata={"source": "unit-test"},
    )
    iteration_request = resource.prepare_process(
        eval_id="eval_123",
        iteration_id="iter_456",
        document=_mime_document(),
    )
    stream_request = resource.prepare_process_stream(
        eval_id="eval_123",
        iteration_id="iter_456",
        document=_mime_document(),
    )

    assert base_request.url == "/evals/extract/extract/eval_123"
    assert base_request.form_data == {"metadata": '{"source": "unit-test"}'}
    assert iteration_request.url == "/evals/extract/extract/eval_123/iter_456"
    assert stream_request.url == "/evals/extract/extract/eval_123/iter_456/stream"


def test_extract_prepare_process_rejects_documents() -> None:
    resource = Extract(client=object())

    try:
        resource.prepare_process(eval_id="eval_123", documents=[_mime_document()])
    except TypeError as exc:
        assert "accepts only 'document'" in str(exc)
    else:
        raise AssertionError("prepare_process should reject 'documents'")


def test_split_prepare_process_path() -> None:
    resource = Split(client=object())

    request = resource.prepare_process(
        eval_id="split_123",
        iteration_id="iter_456",
        document=_mime_document(),
    )

    assert request.url == "/evals/split/extract/split_123/iter_456"


def test_classify_prepare_process_path() -> None:
    resource = Classify(client=object())

    request = resource.prepare_process(
        eval_id="classify_123",
        document=_mime_document(),
    )

    assert request.url == "/evals/classify/extract/classify_123"


def test_sync_process_methods_parse_typed_models() -> None:
    client = MockPreparedClient(
        responses={
            "/evals/extract/extract/eval_123": _build_parsed_completion(),
            "/evals/split/extract/split_123": _build_parsed_completion("split"),
            "/evals/classify/extract/classify_123": _build_classify_response(),
        },
        stream_responses={
            "/evals/extract/extract/eval_123/stream": [
                _build_stream_chunk("first"),
                _build_stream_chunk("second"),
            ]
        },
    )

    extract = Extract(client=client)
    split = Split(client=client)
    classify = Classify(client=client)

    extract_result = extract.process(eval_id="eval_123", document=_mime_document())
    split_result = split.process(eval_id="split_123", document=_mime_document())
    classify_result = classify.process(eval_id="classify_123", document=_mime_document())

    assert extract_result.choices[0].message.parsed == {"status": "ok"}
    assert split_result.choices[0].message.parsed == {"status": "split"}
    assert classify_result.result.classification == "invoice"

    with extract.process_stream(eval_id="eval_123", document=_mime_document()) as stream:
        chunks = list(stream)

    assert [chunk.choices[0].delta.flat_parsed for chunk in chunks] == [{"status": "first"}, {"status": "second"}]
    assert [request.url for request in client.requests] == [
        "/evals/extract/extract/eval_123",
        "/evals/split/extract/split_123",
        "/evals/classify/extract/classify_123",
        "/evals/extract/extract/eval_123/stream",
    ]
