"""Regression tests for the second-pass deep fixes:
- tables/secrets resources wired onto the client
- bare file-like objects (no .name) infer a usable type instead of raising
- assert_valid_file_type raises ValueError (survives `python -O`)
"""

import io

import pytest

from retab import AsyncRetab, Retab
from retab.utils.mime import assert_valid_file_type, prepare_mime_document


def _sync() -> Retab:
    return Retab(api_key="sk_test_dummy", base_url="https://api.retab.com")


def test_tables_and_secrets_wired_sync() -> None:
    client = _sync()
    assert hasattr(client, "tables"), "client.tables not wired"
    assert hasattr(client, "secrets"), "client.secrets not wired"


def test_tables_and_secrets_wired_async() -> None:
    client = AsyncRetab(api_key="sk_test_dummy", base_url="https://api.retab.com")
    assert hasattr(client, "tables"), "async client.tables not wired"
    assert hasattr(client, "secrets"), "async client.secrets not wired"


def test_bare_bytesio_without_name_does_not_raise() -> None:
    # A BytesIO has no `.name`; previously this raised
    # "AssertionError: Invalid file type: uploaded_file".
    stream = io.BytesIO(b"%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")
    md = prepare_mime_document(stream)
    assert md.extension and md.extension != "uploaded_file"


def test_assert_valid_file_type_raises_valueerror() -> None:
    # Must be a real raise (not a bare assert) so it survives `python -O`.
    with pytest.raises(ValueError):
        assert_valid_file_type("xyz")
