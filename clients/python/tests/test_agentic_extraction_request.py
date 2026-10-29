"""Request-shape tests for the agentic_extraction extraction field.

agentic_extraction asks the server to run agentic extraction on a long document
(it implies deep extraction). Its client-side failure modes are silent: a
dropped field, a renamed key, a value routed into the query string by
``**extra_params``, or an unset default that serializes as ``false`` all give a
request that succeeds and quietly runs something else. So these assert the
prepared request directly.

Pure offline: the prepare_* helpers build the request without a client or
network.
"""

from typing import Any

import pytest

from retab.resources.extractions import Extractions
from retab.types.extractions import ExtractionRequest
from retab.types.mime import MIMEData

# Whole module is unit (pure offline; no server/credentials needed).
pytestmark = pytest.mark.unit


DOCUMENT = MIMEData(filename="statement.pdf", url="data:application/pdf;base64,QUJD")
JSON_SCHEMA: dict[str, Any] = {
    "type": "object",
    "properties": {"transactions": {"type": "array", "items": {"type": "object"}}},
}


def _prepare_create(**kwargs: Any) -> Any:
    return Extractions(client=None).prepare_create(  # type: ignore[arg-type]
        document=DOCUMENT,
        json_schema=JSON_SCHEMA,
        model="retab-large",
        **kwargs,
    )


def _create_body(**kwargs: Any) -> dict[str, Any]:
    """Return the JSON body prepare_create would POST to /v1/extractions."""
    prepared = _prepare_create(**kwargs)
    assert prepared.data is not None
    return prepared.data


def test_agentic_extraction_true_is_sent_in_the_body() -> None:
    prepared = _prepare_create(agentic_extraction=True)
    assert prepared.data is not None
    assert prepared.data["agentic_extraction"] is True
    # Not smuggled into the query string through **extra_params: the server
    # reads the field from the JSON body only.
    assert "agentic_extraction" not in (prepared.params or {})
    assert prepared.url.endswith("/v1/extractions")


def test_explicit_false_agentic_extraction_is_sent() -> None:
    """A deliberate opt-out is distinguishable from an unset default.

    ``exclude_none=True`` drops ``None`` but must keep ``False``.
    """
    body = _create_body(agentic_extraction=False)
    assert body["agentic_extraction"] is False


def test_agentic_extraction_is_omitted_when_unset() -> None:
    """Absent, not ``false``.

    The default request body must stay byte-identical to what it was before
    this field existed.
    """
    body = _create_body()
    assert "agentic_extraction" not in body


def test_agentic_extraction_uses_the_servers_field_name() -> None:
    body = _create_body(agentic_extraction=True)
    assert "agentic_extraction" in body
    assert not any(key.lower() == "agenticextraction" for key in body)


def test_agentic_extraction_does_not_set_deep_extraction_on_the_client() -> None:
    """The server applies "agentic implies deep"; the SDK sends what it was given.

    Adding deep_extraction client-side would make the two indistinguishable on
    the wire and change the request body of every agentic caller.
    """
    body = _create_body(agentic_extraction=True)
    assert "deep_extraction" not in body


def test_agentic_extraction_does_not_displace_sibling_fields() -> None:
    body = _create_body(
        agentic_extraction=True,
        deep_extraction=True,
        n_consensus=1,
        background=True,
        bust_cache=True,
        metadata={"source": "docs"},
    )
    assert body["agentic_extraction"] is True
    assert body["deep_extraction"] is True
    assert body["n_consensus"] == 1
    assert body["background"] is True
    assert body["bust_cache"] is True
    assert body["metadata"] == {"source": "docs"}
    assert body["json_schema"] == JSON_SCHEMA
    assert body["model"] == "retab-large"


def test_agentic_extraction_is_sent_on_the_stream_request() -> None:
    """The streaming create has its own prepare path and payload build.

    The server answers 422 for agentic extraction on the streaming route; the
    SDK must still send the field so the caller gets that error instead of a
    silently ordinary streamed extraction.
    """
    for value in (True, False):
        prepared = Extractions(client=None).prepare_create_stream(  # type: ignore[arg-type]
            document=DOCUMENT,
            json_schema=JSON_SCHEMA,
            model="retab-large",
            agentic_extraction=value,
        )
        assert prepared.data is not None
        assert prepared.data["agentic_extraction"] is value
        assert "agentic_extraction" not in (prepared.params or {})
        assert prepared.url.endswith("/v1/extractions/stream")

    unset = Extractions(client=None).prepare_create_stream(  # type: ignore[arg-type]
        document=DOCUMENT,
        json_schema=JSON_SCHEMA,
        model="retab-large",
    )
    assert unset.data is not None
    assert "agentic_extraction" not in unset.data


def test_agentic_extraction_does_not_send_internal_knobs() -> None:
    body = _create_body(agentic_extraction=True)
    for knob in ("long_list_strategy", "window_size", "holdback_rows", "long_list_target_field"):
        assert knob not in body, f"{knob} must not appear on the public request"


def test_extraction_request_model_types_agentic_extraction_as_optional_bool() -> None:
    """Unset is ``None`` (omitted), and a non-bool is rejected, not coerced."""
    assert "agentic_extraction" in ExtractionRequest.model_fields
    assert ExtractionRequest.model_fields["agentic_extraction"].default is None

    with pytest.raises(Exception):
        ExtractionRequest.model_validate(
            {
                "document": DOCUMENT.model_dump(),
                "json_schema": JSON_SCHEMA,
                "model": "retab-large",
                "agentic_extraction": "yes-please",
            }
        )


def test_agentic_extraction_is_not_echoed_on_the_extraction_response_model() -> None:
    """Request-only: the Extraction object does not carry it."""
    from retab.types.extractions import Extraction

    assert "agentic_extraction" not in Extraction.model_fields
