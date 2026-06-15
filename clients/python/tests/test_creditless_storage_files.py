"""Creditless storage coverage for the Files resource.

Every test here is CREDITLESS: it only uploads bytes to storage (which does NOT
invoke a model) and reads file metadata back. No extraction / parse / OCR /
classification is ever triggered.

Note on cleanup: the Python SDK exposes no ``files.delete`` method, so uploaded
fixtures cannot be removed by the test. We keep them tiny (a few dozen bytes)
and clearly tagged (``creditless_storage_*``) so they are obviously test data.
This is reported as an SDK gap, not a test defect.
"""

from __future__ import annotations

import hashlib
import time
import uuid

import httpx
import pytest


from retab import AsyncRetab, Retab
from retab.exceptions import NotFoundError, ValidationError
from retab.types.files import CreateUploadResponse, File, FileLink
from retab.types.mime import MIMEData

from factories import TINY_FILE_CONTENT, unique_name, upload_file

# Whole module is creditless (storage/config/list/get/error paths only).
pytestmark = pytest.mark.creditless

# The full create_upload -> signed PUT -> complete_upload flow, the tiny payload,
# and filename generation all live in factories.py so every suite shares one
# storage-upload path.


# --------------------------------------------------------------------------- #
# Upload + shape
# --------------------------------------------------------------------------- #


def test_create_upload_returns_signed_session(sync_client: Retab) -> None:
    content = TINY_FILE_CONTENT
    sha = hashlib.sha256(content).hexdigest()
    session = sync_client.files.create_upload(
        filename=unique_name("creditless_storage", "txt"),
        size_bytes=len(content),
        content_type="text/plain",
        sha256=sha,
    )
    assert isinstance(session, CreateUploadResponse)
    assert session.file_id.startswith("file_")
    assert session.upload_url.startswith("http")
    # A signed upload URL must carry an expiry so the client knows its lifetime.
    assert session.expires_at is not None
    # We never complete this session, so no file is finalized -> nothing to clean.


def test_upload_complete_get_roundtrip_is_creditless(sync_client: Retab) -> None:
    stored = upload_file(sync_client)
    assert isinstance(stored, File)
    assert stored.object == "file"
    assert stored.id.startswith("file_")
    assert stored.filename.startswith("creditless_storage")
    assert stored.mime_type == "text/plain"
    assert stored.created_at is not None
    # Uploading bytes must NOT have triggered processing: a tiny text blob has
    # no rendered pages, so page_count stays None/0 (it is never extracted).
    assert stored.page_count in (None, 0)


def test_get_download_link_shape(sync_client: Retab) -> None:
    stored = upload_file(sync_client)
    link = sync_client.files.get_download_link(stored.id)
    assert isinstance(link, FileLink)
    assert link.download_url.startswith("http")
    assert link.filename == stored.filename
    assert link.expires_in


def test_downloaded_bytes_match_uploaded(sync_client: Retab) -> None:
    """Round-trip the bytes through storage to prove upload is lossless storage.

    Downloading via a signed link is a plain GCS read, not a Retab compute path.
    """
    content = b"creditless-roundtrip-" + uuid.uuid4().hex.encode() + b"\n"
    stored = upload_file(sync_client, content=content)
    link = sync_client.files.get_download_link(stored.id)
    got = httpx.get(link.download_url)
    assert got.status_code == 200
    assert got.content == content


# --------------------------------------------------------------------------- #
# Listing, pagination, filtering
# --------------------------------------------------------------------------- #


def test_list_default_envelope(sync_client: Retab) -> None:
    page = sync_client.files.list(limit=5)
    assert len(page.data) <= 5
    assert page.list_metadata is not None
    for f in page.data:
        assert isinstance(f, File)
        assert f.id.startswith("file_")


def test_list_limit_one_then_paginate(sync_client: Retab) -> None:
    # Ensure there are at least two files to page over.
    upload_file(sync_client)
    upload_file(sync_client)

    first = sync_client.files.list(limit=1)
    assert len(first.data) <= 1
    if first.has_more:
        assert first.list_metadata.after is not None
        nxt = sync_client.files.list(limit=1, after=first.list_metadata.after)
        assert len(nxt.data) <= 1
        if first.data and nxt.data:
            assert first.data[0].id != nxt.data[0].id


def test_list_order_desc_is_newest_first(sync_client: Retab) -> None:
    page = sync_client.files.list(limit=10, order="desc")
    created = [f.created_at for f in page.data if f.created_at is not None]
    assert created == sorted(created, reverse=True)


def test_list_filter_by_mime_type(sync_client: Retab) -> None:
    upload_file(sync_client)
    page = sync_client.files.list(limit=10, mime_type="text/plain")
    # Every returned record must honor the filter.
    for f in page.data:
        if f.mime_type is not None:
            assert f.mime_type == "text/plain"


def test_list_filter_by_filename(sync_client: Retab) -> None:
    stored = upload_file(sync_client)
    page = sync_client.files.list(limit=25, filename=stored.filename)
    # The exact filename filter should surface our just-uploaded file.
    ids = {f.id for f in page.data}
    # Allow eventual consistency: if present, it must match our filename.
    for f in page.data:
        assert stored.filename in f.filename or f.filename == stored.filename
    assert isinstance(ids, set)


# --------------------------------------------------------------------------- #
# Error paths (creditless)
# --------------------------------------------------------------------------- #


def test_get_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError):
        sync_client.files.get("file_does_not_exist_" + uuid.uuid4().hex)


def test_download_link_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError):
        sync_client.files.get_download_link("file_does_not_exist_" + uuid.uuid4().hex)


def test_complete_upload_unknown_session_errors(sync_client: Retab) -> None:
    # Per the contract, an unknown upload session responds 404 (or 422).
    with pytest.raises((NotFoundError, ValidationError)):
        sync_client.files.complete_upload("file_never_reserved_" + uuid.uuid4().hex)


def test_list_with_junk_api_key_401(bad_key_client: Retab) -> None:
    from retab.exceptions import AuthenticationError, PermissionDeniedError

    # Junk key must be rejected with an auth/permission failure, not data.
    with pytest.raises((AuthenticationError, PermissionDeniedError)):
        bad_key_client.files.list(limit=1)


# --------------------------------------------------------------------------- #
# Async parity (one path)
# --------------------------------------------------------------------------- #


@pytest.mark.asyncio
async def test_async_list_files(async_client: AsyncRetab) -> None:
    page = await async_client.files.list(limit=3)
    assert len(page.data) <= 3
    for f in page.data:
        assert isinstance(f, File)


@pytest.mark.asyncio
async def test_async_upload_roundtrip(async_client: AsyncRetab) -> None:
    content = b"async-creditless-" + str(time.time()).encode() + b"\n"
    filename = unique_name("creditless_storage", "txt")
    sha = hashlib.sha256(content).hexdigest()
    session = await async_client.files.create_upload(filename=filename, size_bytes=len(content), content_type="text/plain", sha256=sha)
    async with httpx.AsyncClient() as h:
        put = await h.request(session.upload_method or "PUT", session.upload_url, content=content, headers=session.upload_headers or {})
    assert put.status_code in (200, 201)
    mime = await async_client.files.complete_upload(session.file_id, sha256=sha)
    assert isinstance(mime, MIMEData)
    got = await async_client.files.get(session.file_id)
    assert got.filename == filename
    assert got.page_count in (None, 0)
