import base64

import pytest

from retab.types.mime import MIMEData
from retab.utils.hashing import generate_blake2b_hash_from_base64


def test_mime_data_id_hashes_inline_content() -> None:
    encoded = base64.b64encode(b"hello").decode("ascii")
    mime_data = MIMEData(filename="hello.txt", url=f"data:text/plain;base64,{encoded}")

    assert mime_data.id == f"file_{generate_blake2b_hash_from_base64(encoded)}"


def test_mime_data_id_uses_retab_storage_file_id() -> None:
    mime_data = MIMEData(filename="invoice.pdf", url="https://storage.retab.com/file_123")

    assert mime_data.id == "file_123"
    assert mime_data.unique_filename == "file_123.pdf"


def test_mime_data_repr_is_safe_for_retab_storage_url() -> None:
    mime_data = MIMEData(filename="invoice.pdf", url="https://storage.retab.com/file_123")

    assert "size='unavailable'" in str(mime_data)


@pytest.mark.parametrize(
    "url",
    [
        "http://storage.retab.com/file_123",
        "https://api.retab.com/file_123",
        "https://storage.retab.com/org_1/file_123",
        "https://storage.retab.com/file_123?x=1",
        "https://storage.retab.com/file_123#fragment",
    ],
)
def test_mime_data_id_does_not_treat_other_remote_urls_as_file_ids(url: str) -> None:
    mime_data = MIMEData(filename="invoice.pdf", url=url)

    with pytest.raises(ValueError, match="Content is not available"):
        _ = mime_data.id
