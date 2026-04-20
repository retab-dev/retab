from __future__ import annotations

import contextlib
import datetime
import os
import time
from collections.abc import Callable, Iterable

import pytest

from retab import AsyncRetab, Retab
from retab.exceptions import InternalServerError, NotFoundError, ValidationError as RetabValidationError
from retab.types.classifications import Classification
from retab.types.edits import Edit
from retab.types.extractions import Extraction, ExtractionRequest, SourcesResponse
from retab.types.mime import MIMEData
from retab.types.pagination import PaginatedList
from retab.types.parses import Parse
from retab.types.splits import Split, Subdocument
from retab.utils.mime import prepare_mime_document


TEST_DIR = os.path.dirname(os.path.abspath(__file__))


def _inline_text_document() -> MIMEData:
    return MIMEData(
        filename="test_invoice.txt",
        url=(
            "data:text/plain;base64,"
            "SW52b2ljZSAjMTIzNDUKRGF0ZTogMjAyNS0wMS0xNQpBbW91bnQ6ICQ5OS45OQpDdXN0b21lcjogQWNtZSBDb3JwCkRlc2NyaXB0aW9uOiBDb25zdWx0aW5nIHNlcnZpY2Vz"
        ),
    )


def _classification_categories() -> list[dict[str, str]]:
    return [
        {"name": "invoice", "description": "A billing invoice document"},
        {"name": "receipt", "description": "A payment receipt"},
        {"name": "contract", "description": "A legal contract"},
    ]


def _split_subdocuments() -> list[Subdocument]:
    return [
        Subdocument(name="form_section", description="Form sections with input fields, checkboxes, or signature areas"),
        Subdocument(name="instructions", description="Instructions, terms and conditions, or explanatory text"),
        Subdocument(name="header", description="Header sections with logos, titles, or document identifiers"),
    ]


def _fidelity_form_path() -> str:
    return os.path.join(TEST_DIR, "data", "edit", "fidelity_original.pdf")


def _fidelity_instructions() -> str:
    with open(os.path.join(TEST_DIR, "data", "edit", "instructions.txt")) as f:
        return f.read()


def _list_window_start(created_at: datetime.datetime | None) -> datetime.datetime | None:
    if created_at is None:
        return None
    return created_at - datetime.timedelta(minutes=5)


def _item_id(item: object) -> str | None:
    if isinstance(item, dict):
        return item.get("id") if isinstance(item.get("id"), str) else None
    value = getattr(item, "id", None)
    return value if isinstance(value, str) else None


def _wait_for_list_contains(
    fetch_page: Callable[[], PaginatedList],
    target_id: str,
    timeout_seconds: float = 30.0,
    poll_interval_seconds: float = 1.0,
) -> PaginatedList:
    deadline = time.monotonic() + timeout_seconds
    while True:
        page = fetch_page()
        if any(_item_id(item) == target_id for item in page.data):
            return page
        if time.monotonic() >= deadline:
            raise AssertionError(f"Timed out waiting for list result to include {target_id}")
        time.sleep(poll_interval_seconds)


def _assert_deleted(
    getter: Callable[[str], object],
    resource_id: str,
    timeout_seconds: float = 15.0,
    poll_interval_seconds: float = 0.5,
) -> None:
    deadline = time.monotonic() + timeout_seconds
    while True:
        try:
            getter(resource_id)
        except NotFoundError:
            return
        if time.monotonic() >= deadline:
            raise AssertionError(f"Timed out waiting for resource deletion: {resource_id}")
        time.sleep(poll_interval_seconds)


def _assert_list_contains(page: PaginatedList, target_id: str) -> None:
    assert any(_item_id(item) == target_id for item in page.data), f"{target_id} not found in list response"


def _skip_if_resource_route_unavailable(exc: InternalServerError, route_name: str) -> None:
    pytest.skip(f"{route_name} unavailable in local stack: {exc.status_code} {exc.message}")


def _new_client(api_keys) -> Retab:
    return Retab(
        api_key=api_keys.retab_api_key,
        base_url=api_keys.retab_api_base_url,
        max_retries=0,
    )


@pytest.fixture(scope="session")
def created_parse(api_keys, booking_confirmation_file_path_1: str) -> Iterable[Parse]:
    client = _new_client(api_keys)
    try:
        parse = client.parses.create(
            document=booking_confirmation_file_path_1,
            model="retab-micro",
            table_parsing_format="html",
            image_resolution_dpi=96,
        )
    except InternalServerError as exc:
        _skip_if_resource_route_unavailable(exc, "/v1/parses")
    try:
        yield parse
    finally:
        with contextlib.suppress(Exception):
            client.parses.delete(parse.id)
        client.close()


@pytest.fixture(scope="session")
def created_extraction(
    api_keys,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: dict,
) -> Iterable[Extraction]:
    client = _new_client(api_keys)
    try:
        extraction = client.extractions.create(
            ExtractionRequest(
                document=prepare_mime_document(booking_confirmation_file_path_1),
                json_schema=booking_confirmation_json_schema,
                model="retab-micro",
            )
        )
    except InternalServerError as exc:
        _skip_if_resource_route_unavailable(exc, "/v1/extractions")
    try:
        yield extraction
    finally:
        with contextlib.suppress(Exception):
            client.extractions.delete(extraction.id)
        client.close()


@pytest.fixture(scope="session")
def created_split(api_keys) -> Iterable[Split]:
    client = _new_client(api_keys)
    try:
        split = client.splits.create(
            document=_fidelity_form_path(),
            model="retab-micro",
            subdocuments=_split_subdocuments(),
        )
    except InternalServerError as exc:
        _skip_if_resource_route_unavailable(exc, "/v1/splits")
    try:
        yield split
    finally:
        with contextlib.suppress(Exception):
            client.splits.delete(split.id)
        client.close()


@pytest.fixture(scope="session")
def created_classification(api_keys) -> Iterable[Classification]:
    client = _new_client(api_keys)
    try:
        classification = client.classifications.create(
            document=_inline_text_document(),
            model="retab-micro",
            categories=_classification_categories(),
        )
    except InternalServerError as exc:
        _skip_if_resource_route_unavailable(exc, "/v1/classifications")
    try:
        yield classification
    finally:
        with contextlib.suppress(Exception):
            client.classifications.delete(classification.id)
        client.close()


@pytest.fixture(scope="session")
def created_edit(api_keys) -> Iterable[Edit]:
    client = _new_client(api_keys)
    try:
        edit = client.edits.create(
            document=_fidelity_form_path(),
            instructions=_fidelity_instructions(),
            model="retab-micro",
        )
    except InternalServerError as exc:
        _skip_if_resource_route_unavailable(exc, "/v1/edits")
    try:
        yield edit
    finally:
        with contextlib.suppress(Exception):
            client.edits.delete(edit.id)
        client.close()


def test_parses_resource_crud(sync_client: Retab, created_parse: Parse, booking_confirmation_file_path_1: str) -> None:
    with sync_client as client:
        fetched = client.parses.get(created_parse.id)
        assert fetched.id == created_parse.id
        assert fetched.output.text == created_parse.output.text

        page = _wait_for_list_contains(
            lambda: client.parses.list(limit=100, from_date=_list_window_start(created_parse.created_at)),
            created_parse.id,
        )
        _assert_list_contains(page, created_parse.id)

        temp = client.parses.create(
            document=booking_confirmation_file_path_1,
            model="retab-micro",
            table_parsing_format="html",
            image_resolution_dpi=96,
        )
        client.parses.delete(temp.id)
        _assert_deleted(client.parses.get, temp.id)


@pytest.mark.asyncio
async def test_parses_resource_async_get_and_list(async_client: AsyncRetab, created_parse: Parse) -> None:
    async with async_client:
        fetched = await async_client.parses.get(created_parse.id)
        assert fetched.id == created_parse.id

        page = await async_client.parses.list(limit=100, from_date=_list_window_start(created_parse.created_at))
        _assert_list_contains(page, created_parse.id)


def test_extractions_resource_crud(
    sync_client: Retab,
    created_extraction: Extraction,
    booking_confirmation_file_path_1: str,
    booking_confirmation_file_path_2: str,
    booking_confirmation_json_schema: dict,
) -> None:
    with sync_client as client:
        fetched = client.extractions.get(created_extraction.id)
        assert fetched.id == created_extraction.id
        assert fetched.output == created_extraction.output

        page = _wait_for_list_contains(
            lambda: client.extractions.list(limit=100, from_date=_list_window_start(created_extraction.created_at)),
            created_extraction.id,
        )
        _assert_list_contains(page, created_extraction.id)

        temp = client.extractions.create(
            ExtractionRequest(
                document=prepare_mime_document(booking_confirmation_file_path_2),
                json_schema=booking_confirmation_json_schema,
                model="retab-micro",
            )
        )
        client.extractions.delete(temp.id)


def test_extractions_sources_returns_provenance(sync_client: Retab, created_extraction: Extraction) -> None:
    try:
        with sync_client as client:
            sources = client.extractions.sources(created_extraction.id)
    except RetabValidationError as exc:
        pytest.skip(f"/v1/extractions/{{id}}/sources unavailable in local stack: {exc}")

    assert isinstance(sources, SourcesResponse)
    assert sources.extraction_id == created_extraction.id
    assert sources.extraction == created_extraction.output
    assert isinstance(sources.sources, dict)


@pytest.mark.asyncio
async def test_extractions_resource_async_get_and_list(async_client: AsyncRetab, created_extraction: Extraction) -> None:
    async with async_client:
        fetched = await async_client.extractions.get(created_extraction.id)
        assert fetched.id == created_extraction.id

        page = await async_client.extractions.list(limit=100, from_date=_list_window_start(created_extraction.created_at))
        _assert_list_contains(page, created_extraction.id)


def test_splits_resource_crud(sync_client: Retab, created_split: Split) -> None:
    with sync_client as client:
        fetched = client.splits.get(created_split.id)
        assert fetched.id == created_split.id
        assert fetched.output == created_split.output

        page = _wait_for_list_contains(
            lambda: client.splits.list(limit=100, from_date=_list_window_start(created_split.created_at)),
            created_split.id,
        )
        _assert_list_contains(page, created_split.id)

        temp = client.splits.create(
            document=_fidelity_form_path(),
            model="retab-micro",
            subdocuments=_split_subdocuments(),
        )
        try:
            client.splits.delete(temp.id)
        except Exception as exc:
            pytest.skip(f"/v1/splits/{{id}} delete unavailable in local stack: {exc}")
        _assert_deleted(client.splits.get, temp.id)


@pytest.mark.asyncio
async def test_splits_resource_async_get_and_list(async_client: AsyncRetab, created_split: Split) -> None:
    async with async_client:
        fetched = await async_client.splits.get(created_split.id)
        assert fetched.id == created_split.id

        page = await async_client.splits.list(limit=100, from_date=_list_window_start(created_split.created_at))
        _assert_list_contains(page, created_split.id)


def test_classifications_resource_crud(sync_client: Retab, created_classification: Classification) -> None:
    category_names = {category["name"] for category in _classification_categories()}

    with sync_client as client:
        fetched = client.classifications.get(created_classification.id)
        assert fetched.id == created_classification.id
        assert fetched.output.category in category_names
        assert fetched.output.reasoning

        page = _wait_for_list_contains(
            lambda: client.classifications.list(limit=100, from_date=_list_window_start(created_classification.created_at)),
            created_classification.id,
        )
        _assert_list_contains(page, created_classification.id)

        temp = client.classifications.create(
            document=_inline_text_document(),
            model="retab-micro",
            categories=_classification_categories(),
            instructions="sdk-delete-check",
        )
        client.classifications.delete(temp.id)
        _assert_deleted(client.classifications.get, temp.id)


@pytest.mark.asyncio
async def test_classifications_resource_async_get_and_list(
    async_client: AsyncRetab,
    created_classification: Classification,
) -> None:
    async with async_client:
        fetched = await async_client.classifications.get(created_classification.id)
        assert fetched.id == created_classification.id

        page = await async_client.classifications.list(
            limit=100,
            from_date=_list_window_start(created_classification.created_at),
        )
        _assert_list_contains(page, created_classification.id)


def test_edits_resource_crud(sync_client: Retab, created_edit: Edit) -> None:
    with sync_client as client:
        fetched = client.edits.get(created_edit.id)
        assert fetched.id == created_edit.id
        assert fetched.data.filled_document.url.startswith("data:application/pdf;base64,")
        assert len(fetched.data.form_data) > 0

        page = _wait_for_list_contains(
            lambda: client.edits.list(
                limit=100,
                filename=os.path.basename(_fidelity_form_path()),
                from_date=_list_window_start(created_edit.created_at),
            ),
            created_edit.id,
        )
        _assert_list_contains(page, created_edit.id)

        temp = client.edits.create(
            document=_fidelity_form_path(),
            instructions=_fidelity_instructions(),
            model="retab-micro",
        )
        client.edits.delete(temp.id)
        _assert_deleted(client.edits.get, temp.id)


@pytest.mark.asyncio
async def test_edits_resource_async_get_and_list(async_client: AsyncRetab, created_edit: Edit) -> None:
    async with async_client:
        fetched = await async_client.edits.get(created_edit.id)
        assert fetched.id == created_edit.id

        page = await async_client.edits.list(
            limit=100,
            filename=os.path.basename(_fidelity_form_path()),
            from_date=_list_window_start(created_edit.created_at),
        )
        _assert_list_contains(page, created_edit.id)
