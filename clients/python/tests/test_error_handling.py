"""Unit tests for SDK error handling: APIError formatting, structured parsing, and retry behavior."""

import json

import httpx
import pytest

from retab.exceptions import (
    APIError,
    AuthenticationError,
    InternalServerError,
    NotFoundError,
    PermissionDeniedError,
    RateLimitError,
    ValidationError,
)


# ---------------------------------------------------------------------------
# APIError construction & fields
# ---------------------------------------------------------------------------


def test_api_error_basic_fields():
    err = APIError("something broke", status_code=500)
    assert err.status_code == 500
    assert err.message == "something broke"
    assert err.code is None
    assert err.details is None
    assert err.body == ""
    assert err.request_id is None
    assert err.method is None
    assert err.url is None
    assert err.retries == 0


def test_api_error_all_fields():
    err = APIError(
        "bad request",
        status_code=400,
        code="INVALID_SCHEMA",
        details={"field": "json_schema", "reason": "missing required property"},
        body='{"detail": {"code": "INVALID_SCHEMA"}}',
        request_id="req_abc123",
        method="POST",
        url="http://localhost:4000/v1/documents/extract",
    )
    assert err.status_code == 400
    assert err.code == "INVALID_SCHEMA"
    assert err.details == {"field": "json_schema", "reason": "missing required property"}
    assert err.request_id == "req_abc123"
    assert err.method == "POST"
    assert err.url == "http://localhost:4000/v1/documents/extract"
    assert err.retries == 0


def test_api_error_is_runtime_error():
    err = APIError("fail", status_code=500)
    assert isinstance(err, RuntimeError)
    assert isinstance(err, Exception)


# ---------------------------------------------------------------------------
# APIError.__str__  (the main user-facing output)
# ---------------------------------------------------------------------------


def test_str_minimal():
    err = APIError("Request failed (502)", status_code=502)
    s = str(err)
    assert "502" in s
    assert "Request failed (502)" in s
    # No extra lines for empty optional fields
    assert "URL:" not in s
    assert "Request-ID:" not in s
    assert "Retries:" not in s


def test_str_with_url():
    err = APIError(
        "Request failed (502)",
        status_code=502,
        method="POST",
        url="http://localhost:4000/v1/documents/extract",
    )
    s = str(err)
    assert "POST http://localhost:4000/v1/documents/extract" in s


def test_str_with_request_id():
    err = APIError("fail", status_code=500, request_id="req_xyz")
    s = str(err)
    assert "req_xyz" in s
    assert "Request-ID:" in s


def test_str_with_code():
    err = APIError("fail", status_code=422, code="INVALID_SCHEMA")
    s = str(err)
    assert "INVALID_SCHEMA" in s
    assert "Code:" in s


def test_str_with_details():
    err = APIError("fail", status_code=422, details={"field": "name"})
    s = str(err)
    assert "field" in s
    assert "Details:" in s


def test_str_with_body():
    err = APIError("fail", status_code=502, body='{"error": "upstream timeout"}')
    s = str(err)
    assert "upstream timeout" in s
    assert "Body:" in s


def test_str_body_truncated():
    long_body = "x" * 1000
    err = APIError("fail", status_code=502, body=long_body)
    s = str(err)
    assert "..." in s
    # Should contain at most 500 chars of body + "..."
    body_line = [line for line in s.split("\n") if "Body:" in line][0]
    # 500 chars of body + "..." + "  Body:       " prefix
    assert len(body_line) < 520


def test_str_body_not_truncated_when_short():
    short_body = "x" * 499
    err = APIError("fail", status_code=502, body=short_body)
    s = str(err)
    assert "..." not in s


def test_str_with_retries():
    err = APIError("fail", status_code=502)
    err.retries = 4
    s = str(err)
    assert "Retries:" in s
    assert "4" in s


def test_str_no_retries_line_when_zero():
    err = APIError("fail", status_code=502)
    s = str(err)
    assert "Retries:" not in s


def test_str_full_output():
    err = APIError(
        "upstream timeout",
        status_code=502,
        code="GATEWAY_ERROR",
        details={"service": "preprocessing_server"},
        body='{"error": "upstream timeout from preprocessing_server"}',
        request_id="req_abc123",
        method="POST",
        url="http://localhost:4000/v1/documents/extract",
    )
    err.retries = 3
    s = str(err)
    lines = s.split("\n")
    assert "502" in lines[0]
    assert "upstream timeout" in lines[0]
    assert any("POST http://localhost:4000/v1/documents/extract" in l for l in lines)
    assert any("req_abc123" in l for l in lines)
    assert any("GATEWAY_ERROR" in l for l in lines)
    assert any("preprocessing_server" in l for l in lines)
    assert any("Body:" in l for l in lines)
    assert any("3" in l and "Retries:" in l for l in lines)


# ---------------------------------------------------------------------------
# APIError.__repr__
# ---------------------------------------------------------------------------


def test_repr_includes_class_name():
    err = InternalServerError("fail", status_code=502)
    r = repr(err)
    assert r.startswith("InternalServerError(")


def test_repr_includes_body_when_present():
    err = APIError("fail", status_code=500, body="some body")
    r = repr(err)
    assert "body=" in r


def test_repr_excludes_body_when_empty():
    err = APIError("fail", status_code=500, body="")
    r = repr(err)
    assert "body=" not in r


def test_repr_includes_retries_when_nonzero():
    err = APIError("fail", status_code=500)
    err.retries = 2
    r = repr(err)
    assert "retries=2" in r


def test_repr_excludes_retries_when_zero():
    err = APIError("fail", status_code=500)
    r = repr(err)
    assert "retries=" not in r


# ---------------------------------------------------------------------------
# Exception hierarchy
# ---------------------------------------------------------------------------


def test_subclass_hierarchy():
    assert issubclass(AuthenticationError, APIError)
    assert issubclass(PermissionDeniedError, APIError)
    assert issubclass(NotFoundError, APIError)
    assert issubclass(ValidationError, APIError)
    assert issubclass(RateLimitError, APIError)
    assert issubclass(InternalServerError, APIError)


def test_subclass_inherits_all_fields():
    err = InternalServerError(
        "bad gateway",
        status_code=502,
        code="GATEWAY",
        body="raw body",
        method="GET",
        url="/v1/models",
        request_id="req_1",
    )
    assert err.status_code == 502
    assert err.code == "GATEWAY"
    assert err.body == "raw body"
    assert err.method == "GET"
    assert err.url == "/v1/models"
    assert err.request_id == "req_1"
    assert err.retries == 0
    assert isinstance(err, APIError)
    assert isinstance(err, RuntimeError)


def test_catch_by_parent_class():
    with pytest.raises(APIError):
        raise InternalServerError("fail", status_code=500)


# ---------------------------------------------------------------------------
# _validate_response — structured error parsing
# ---------------------------------------------------------------------------


def _make_mock_response(
    status_code: int,
    body: str,
    headers: dict | None = None,
    method: str = "POST",
    url: str = "http://localhost:4000/v1/test",
) -> httpx.Response:
    """Build a mock httpx.Response for _validate_response tests."""
    request = httpx.Request(method=method, url=url)
    response = httpx.Response(
        status_code=status_code,
        text=body,
        headers=headers or {},
        request=request,
    )
    return response


def _validate(response: httpx.Response) -> APIError:
    """Call _validate_response via a minimal client and capture the raised error."""
    from retab.client import BaseRetab

    client = object.__new__(BaseRetab)
    with pytest.raises(APIError) as exc_info:
        client._validate_response(response)
    return exc_info.value


def test_validate_structured_json_detail_dict():
    body = json.dumps({
        "detail": {
            "code": "INVALID_SCHEMA",
            "message": "Schema validation failed",
            "details": {"field": "json_schema"},
        }
    })
    resp = _make_mock_response(422, body)
    err = _validate(resp)
    assert isinstance(err, ValidationError)
    assert err.code == "INVALID_SCHEMA"
    assert err.message == "Schema validation failed"
    assert err.details == {"field": "json_schema"}
    assert err.body == body
    assert err.method == "POST"
    assert err.url == "http://localhost:4000/v1/test"


def test_validate_structured_json_detail_string():
    body = json.dumps({"detail": "Not authenticated"})
    resp = _make_mock_response(401, body)
    err = _validate(resp)
    assert isinstance(err, AuthenticationError)
    assert err.message == "Not authenticated"
    assert err.body == body


def test_validate_unstructured_json():
    body = json.dumps({"error": "something unexpected"})
    resp = _make_mock_response(500, body)
    err = _validate(resp)
    assert isinstance(err, InternalServerError)
    # No 'detail' key, so message stays as fallback
    assert "Request failed (500)" in err.message


def test_validate_plain_text_body():
    body = "Bad Gateway: upstream timeout"
    resp = _make_mock_response(502, body)
    err = _validate(resp)
    assert isinstance(err, InternalServerError)
    assert err.message == "Bad Gateway: upstream timeout"
    assert err.body == body


def test_validate_empty_body():
    resp = _make_mock_response(500, "")
    err = _validate(resp)
    assert isinstance(err, InternalServerError)
    assert "Request failed (500)" in err.message


def test_validate_request_id_header():
    resp = _make_mock_response(
        500,
        "fail",
        headers={"x-request-id": "req_test123"},
    )
    err = _validate(resp)
    assert err.request_id == "req_test123"


def test_validate_method_and_url():
    resp = _make_mock_response(
        404,
        json.dumps({"detail": "Not found"}),
        method="GET",
        url="http://localhost:4000/v1/files/abc",
    )
    err = _validate(resp)
    assert isinstance(err, NotFoundError)
    assert err.method == "GET"
    assert err.url == "http://localhost:4000/v1/files/abc"


def test_validate_status_code_mapping():
    cases = [
        (401, AuthenticationError),
        (403, PermissionDeniedError),
        (404, NotFoundError),
        (422, ValidationError),
        (429, RateLimitError),
        (500, InternalServerError),
        (502, InternalServerError),
        (503, InternalServerError),
    ]
    for status, expected_cls in cases:
        resp = _make_mock_response(status, "{}")
        err = _validate(resp)
        assert isinstance(err, expected_cls), f"Expected {expected_cls.__name__} for status {status}, got {type(err).__name__}"


# ---------------------------------------------------------------------------
# raise_max_tries_exceeded — retry behavior
# ---------------------------------------------------------------------------


def test_raise_max_tries_preserves_api_error():
    from retab.client import raise_max_tries_exceeded

    original = InternalServerError(
        "bad gateway",
        status_code=502,
        body="raw body",
        method="POST",
        url="/v1/extract",
        request_id="req_1",
    )
    details = {"exception": original, "tries": 4}

    with pytest.raises(InternalServerError) as exc_info:
        raise_max_tries_exceeded(details)

    err = exc_info.value
    assert err is original
    assert err.retries == 4
    assert err.status_code == 502
    assert err.body == "raw body"
    assert err.method == "POST"


def test_raise_max_tries_wraps_non_api_error():
    from retab.client import raise_max_tries_exceeded

    original = ConnectionError("refused")
    details = {"exception": original, "tries": 3}

    with pytest.raises(Exception, match="Max tries exceeded after 3 tries"):
        raise_max_tries_exceeded(details)


def test_raise_max_tries_no_exception():
    from retab.client import raise_max_tries_exceeded

    details = {"tries": 2}

    with pytest.raises(Exception, match="Max tries exceeded after 2 tries"):
        raise_max_tries_exceeded(details)


# ---------------------------------------------------------------------------
# End-to-end: str output in traceback-like scenario
# ---------------------------------------------------------------------------


def test_error_str_in_except_block():
    """Verify that catching and printing an APIError shows all context."""
    err = InternalServerError(
        "upstream timeout",
        status_code=502,
        body='{"error": "preprocessing_server unreachable"}',
        method="POST",
        url="http://localhost:4000/v1/documents/extract",
        request_id="req_999",
    )
    err.retries = 3

    output = str(err)
    assert "502" in output
    assert "upstream timeout" in output
    assert "POST http://localhost:4000/v1/documents/extract" in output
    assert "req_999" in output
    assert "preprocessing_server unreachable" in output
    assert "3" in output
