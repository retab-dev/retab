"""Regression tests for RFC 2397 data: URL support in prepare_mime_document.

Bug: prepare_mime_document treated `str` inputs as either an https:// URL or a
filesystem path, so a `data:` URL was passed to `Path(...)` / `open(...)` and
raised FileNotFoundError. This file pins the new branch that decodes a data
URL directly into a MIMEData.
"""

import base64

from retab.types.mime import MIMEData
from retab.utils.mime import prepare_mime_document


def test_prepare_mime_document_handles_base64_text_data_url() -> None:
    encoded = base64.b64encode(b"Hello").decode("ascii")
    data_url = f"data:text/plain;base64,{encoded}"

    mime_data = prepare_mime_document(data_url)

    assert isinstance(mime_data, MIMEData)
    assert mime_data.mime_type == "text/plain"
    assert mime_data.content == encoded
    # Payload round-trips back to the original bytes.
    assert base64.b64decode(mime_data.content) == b"Hello"


def test_prepare_mime_document_handles_url_encoded_text_data_url() -> None:
    # No `;base64` marker => payload is urlencoded text.
    data_url = "data:text/plain,hello%20world"

    mime_data = prepare_mime_document(data_url)

    assert isinstance(mime_data, MIMEData)
    assert mime_data.mime_type == "text/plain"
    expected_encoded = base64.b64encode(b"hello world").decode("ascii")
    assert mime_data.content == expected_encoded
    assert base64.b64decode(mime_data.content) == b"hello world"


def test_prepare_mime_document_handles_base64_pdf_data_url() -> None:
    # Minimal PDF magic header bytes — enough to assert mime_type comes from
    # the data URL header, not from a filename guess.
    pdf_bytes = b"%PDF-1.4\n%\xe2\xe3\xcf\xd3\n"
    encoded = base64.b64encode(pdf_bytes).decode("ascii")
    data_url = f"data:application/pdf;base64,{encoded}"

    mime_data = prepare_mime_document(data_url)

    assert isinstance(mime_data, MIMEData)
    assert mime_data.mime_type == "application/pdf"
    assert mime_data.content == encoded
    assert base64.b64decode(mime_data.content) == pdf_bytes
