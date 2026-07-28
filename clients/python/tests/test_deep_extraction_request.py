"""Request-shape tests for the deep_extraction extraction knob.

deep_extraction opts an extraction into the server's conversational long-array
strategy. Every failure mode it has on this side is SILENT: a dropped field, a
renamed key, or a default that serializes as ``false`` all produce a request
that succeeds and quietly returns the truncated single-shot output the flag
exists to fix. So these assert the prepared request body directly rather than
any response behavior.

Pure offline — the prepare_* helpers build the request without a client or
network.
"""

from typing import Any

import pytest

from retab.resources.extractions import Extractions
from retab.types.mime import MIMEData
from retab.types.extractions import ExtractionRequest

# Whole module is unit (pure offline; no server/credentials needed).
pytestmark = pytest.mark.unit


DOCUMENT = MIMEData(filename="ledger.pdf", url="data:application/pdf;base64,QUJD")
JSON_SCHEMA: dict[str, Any] = {
    "type": "object",
    "properties": {"rows": {"type": "array", "items": {"type": "object"}}},
}


def _create_body(**kwargs: Any) -> dict[str, Any]:
    """Return the JSON body prepare_create would POST to /v1/extractions."""
    prepared = Extractions(client=None).prepare_create(  # type: ignore[arg-type]
        document=DOCUMENT,
        json_schema=JSON_SCHEMA,
        model="retab-large",
        **kwargs,
    )
    assert prepared.data is not None
    return prepared.data


def test_deep_extraction_true_is_sent() -> None:
    body = _create_body(deep_extraction=True)
    assert body["deep_extraction"] is True


def test_deep_extraction_is_omitted_by_default() -> None:
    """Absent, not ``false``.

    The default request body must stay byte-identical to what it was before
    this field existed: an unconditional ``deep_extraction: false`` would be a
    wire-shape change for every caller who never asked for the feature.
    """
    body = _create_body()
    assert "deep_extraction" not in body


def test_explicit_false_deep_extraction_is_sent() -> None:
    """A deliberate opt-out must be distinguishable from an unset default.

    ``exclude_none=True`` drops ``None`` but must keep ``False`` — a caller
    overriding a wrapper's ``True`` needs the key to actually reach the server.
    """
    body = _create_body(deep_extraction=False)
    assert body["deep_extraction"] is False


def test_deep_extraction_uses_the_servers_field_name() -> None:
    """The serialized key must be snake_case ``deep_extraction``.

    An alias mismatch would serialize a key the API ignores, so the request
    would succeed and run the default extraction with no error anywhere.
    """
    body = _create_body(deep_extraction=True)
    assert "deep_extraction" in body
    assert not any(key.lower() == "deepextraction" for key in body)


def test_deep_extraction_is_sent_on_the_stream_request() -> None:
    """The streaming create has its own prepare path and its own payload build.

    The server dispatches the conversational strategy on its streaming path
    too, so a field wired into only the non-streaming create is a real hole.
    """
    prepared = Extractions(client=None).prepare_create_stream(  # type: ignore[arg-type]
        document=DOCUMENT,
        json_schema=JSON_SCHEMA,
        model="retab-large",
        deep_extraction=True,
    )
    assert prepared.data is not None
    assert prepared.data["deep_extraction"] is True
    # The streaming create is identified by its URL, not a body flag.
    assert prepared.url.endswith("/v1/extractions/stream")


def test_deep_extraction_does_not_displace_sibling_fields() -> None:
    """The knob is additive: nothing else in the request may shift.

    n_consensus matters most here — the server runs one independent windowed
    conversation per vote, so the two compose rather than conflict.
    """
    body = _create_body(
        deep_extraction=True,
        n_consensus=3,
        bust_cache=True,
        instructions="focus on the line items",
    )
    assert body["deep_extraction"] is True
    assert body["n_consensus"] == 3
    assert body["bust_cache"] is True
    assert body["instructions"] == "focus on the line items"
    assert body["json_schema"] == JSON_SCHEMA


def test_deep_extraction_does_not_send_the_tuning_knobs() -> None:
    """deep_extraction is a single opaque switch.

    The window size and per-window row holdback stay at the server's defaults
    and are deliberately not on the public surface. Sending an explicit
    ``holdback_rows: 0`` would DISABLE the boundary-row holdback rather than
    request the default of 1.
    """
    body = _create_body(deep_extraction=True)
    for knob in ("window_size", "holdback_rows", "long_list_strategy", "long_list_target_field"):
        assert knob not in body, f"{knob} must not appear on the public request"


def test_extraction_request_model_types_deep_extraction_as_optional_bool() -> None:
    """The model must reject a non-bool rather than coerce it.

    A string ``"true"`` silently landing as truthy — or worse, a truthy value
    silently landing as ``False`` — is the worst failure this field has: the
    request looks like it asked for deep extraction and quietly does not.
    """
    assert ExtractionRequest.model_fields["deep_extraction"].default is None

    with pytest.raises(Exception):
        ExtractionRequest.model_validate(
            {
                "document": DOCUMENT.model_dump(),
                "json_schema": JSON_SCHEMA,
                "model": "retab-large",
                "deep_extraction": "yes-please",
            }
        )


def test_deep_extraction_is_not_echoed_on_the_extraction_response_model() -> None:
    """It is a request-only knob.

    The public Extraction object carries no such field; adding one would be a
    silent response-shape change for every consumer.
    """
    from retab.types.extractions import Extraction

    assert "deep_extraction" not in Extraction.model_fields
