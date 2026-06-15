import json
import re
import warnings
import os
import shutil
import sys
from typing import IO, Any

import pytest
import pytest_asyncio

os.environ["EMAIL_DOMAIN"] = "mailbox.retab.com"
from enum import Enum
from typing import AsyncGenerator, Generator

from dotenv import load_dotenv
from pydantic import BaseModel, Field

# Get the directory containing the tests
TEST_DIR = os.path.dirname(os.path.abspath(__file__))
SDK_ROOT = os.path.dirname(TEST_DIR)
if SDK_ROOT not in sys.path:
    sys.path.insert(0, SDK_ROOT)

from retab import AsyncRetab, Retab


def pytest_addoption(parser: pytest.Parser) -> None:
    parser.addoption("--production", action="store_true", default=False, help="run tests against production API")
    parser.addoption("--local", action="store_true", default=False, help="run tests against local API")
    parser.addoption("--staging", action="store_true", default=False, help="run tests against staging API")
    parser.addoption("--env-file", type=str, help="path to the .env file to use")


def pytest_configure(config: pytest.Config) -> None:
    """Register the credit-cost markers so the suite can be sliced safely.

    ``creditless`` — exercises only storage / config CRUD / list / get / error
    paths; never runs a model or document-processing primitive, so it is safe to
    run against any environment (including production) without consuming credits.

    ``billable`` — creates extractions/parses/splits/classifications/edits (the
    ``created_*`` fixtures), which run real inference and DO consume credits.

    Run the safe subset with ``pytest -m creditless`` (or exclude the costly one
    with ``-m "not billable"``).
    """
    config.addinivalue_line("markers", "creditless: makes no billable/LLM/processing calls; safe to run anywhere")
    config.addinivalue_line("markers", "billable: creates primitives that consume credits (runs real inference)")
    config.addinivalue_line("markers", "unit: pure offline test (mocked client / local logic); no server or credentials needed")


@pytest.fixture(scope="session", autouse=True)
def load_env(request: pytest.FixtureRequest) -> None:
    """Load the appropriate .env file based on the environment flag"""
    env_file = request.config.getoption("--env-file")

    if env_file:
        env_path = env_file
    elif request.config.getoption("--production"):
        env_path = os.path.join(os.path.dirname(TEST_DIR), "../../.env.production")
    elif request.config.getoption("--local"):
        env_path = os.path.join(os.path.dirname(TEST_DIR), "../../.env.local")
    elif request.config.getoption("--staging"):
        env_path = os.path.join(os.path.dirname(TEST_DIR), "../../.env.staging")
    else:
        # No environment specified -> offline mode. The unit suite (``-m unit``)
        # needs no server or credentials, so this must NOT hard-fail; any live
        # test (creditless/billable) that needs a key will fail clearly at the
        # ``api_keys`` fixture instead.
        warnings.warn(
            "No environment specified (--env-file/--production/--local/--staging); running offline. Live tests will fail at the api_keys fixture.",
            UserWarning,
        )
        return

    print("loading env file: ", env_path)
    if not os.path.exists(env_path):
        warnings.warn(f"Environment file not found: {env_path}", UserWarning)
    else:
        load_dotenv(env_path, override=True)
    print("EMAIL_DOMAIN", os.environ["EMAIL_DOMAIN"])


_BASE_URL_VERSION_SUFFIX_RE = re.compile(r"/v\d+/?$")


def _strip_legacy_version_suffix(base_url: str) -> str:
    """Normalize a legacy ``…/v<N>`` base URL to the new convention.

    The SDK constructor will do the same stripping internally and emit a
    deprecation ``UserWarning`` — handling it here at test-setup time keeps
    the warning out of every test run while still letting developer
    ``.env`` files lag behind the new convention.
    """
    return _BASE_URL_VERSION_SUFFIX_RE.sub("", base_url).rstrip("/")


class EnvConfig(BaseModel):
    retab_api_key: str = Field(..., description="Retab API key")
    retab_api_base_url: str = Field(..., description="Retab API base URL")


@pytest.fixture(scope="session")
def api_keys(load_env: None) -> EnvConfig:
    _ = load_env
    retab_api_key = os.getenv("RETAB_API_KEY")
    retab_api_base_url = os.getenv("RETAB_API_BASE_URL")

    assert retab_api_key is not None, "RETAB_API_KEY must be set in environment"
    assert retab_api_base_url is not None, "RETAB_API_BASE_URL must be set in environment"

    return EnvConfig(
        retab_api_key=retab_api_key,
        retab_api_base_url=_strip_legacy_version_suffix(retab_api_base_url),
    )


@pytest.fixture(scope="session")
def base_url(api_keys: EnvConfig) -> str:
    return api_keys.retab_api_base_url


@pytest.fixture(scope="session")
def retab_api_key(api_keys: EnvConfig) -> str:
    return api_keys.retab_api_key


@pytest.fixture(scope="function")
def sync_client(api_keys: EnvConfig) -> Generator[Retab, None, None]:
    client = Retab(
        api_key=api_keys.retab_api_key,
        base_url=api_keys.retab_api_base_url,
        max_retries=0,
    )
    try:
        yield client
    finally:
        client.close()


@pytest_asyncio.fixture(scope="function")
async def async_client(api_keys: EnvConfig) -> AsyncGenerator[AsyncRetab, None]:
    client = AsyncRetab(
        api_key=api_keys.retab_api_key,
        base_url=api_keys.retab_api_base_url,
        max_retries=0,
    )
    try:
        yield client
    finally:
        await client.close()


@pytest.fixture(scope="session")
def test_data_dir() -> str:
    """Return the path to the test data directory"""
    return os.path.join(TEST_DIR, "data")


@pytest.fixture(scope="session")
def booking_confirmation_json_schema(test_data_dir: str) -> dict[str, Any]:
    schema_path = os.path.join(test_data_dir, "freight", "booking_confirmation_schema_small.json")
    with open(schema_path) as f:
        return json.load(f)


@pytest.fixture(scope="session")
def booking_confirmation_file_path_1(test_data_dir: str) -> str:
    return os.path.join(test_data_dir, "freight", "booking_confirmation_1.jpg")


@pytest.fixture(scope="session")
def booking_confirmation_file_path_2(test_data_dir: str) -> str:
    return os.path.join(test_data_dir, "freight", "booking_confirmation_2.jpg")


@pytest.fixture(scope="session")
def booking_confirmation_data_1(test_data_dir: str) -> dict[str, Any]:
    data_path = os.path.join(test_data_dir, "freight", "booking_confirmation_1_data.json")
    with open(data_path) as f:
        return json.load(f)


@pytest.fixture(scope="session")
def booking_confirmation_data_2(test_data_dir: str) -> dict[str, Any]:
    data_path = os.path.join(test_data_dir, "freight", "booking_confirmation_2_data.json")
    with open(data_path) as f:
        return json.load(f)


@pytest.fixture(scope="session")
def company_json_schema() -> dict[str, Any]:
    class CompanyEnum(str, Enum):
        school = "school"
        investor = "investor"
        startup = "startup"
        corporate = "corporate"

    class CompanyRelation(str, Enum):
        founderBackground = "founderBackground"
        investor = "investor"
        competitor = "competitor"
        client = "client"
        partnership = "partnership"

    class Company(BaseModel):
        name: str = Field(..., description="Name of the identified company", json_schema_extra={"X-FieldPrompt": "Look for the name of the company, or derive it from the logo"})
        type: CompanyEnum = Field(..., description="Type of the identified company", json_schema_extra={"X-FieldPrompt": "Guess the type depending on slide context"})
        relationship: CompanyRelation = Field(
            ...,
            description="Relationship of the identified company with the startup from the deck",
            json_schema_extra={"X-FieldPrompt": "Guess the relationship of the identified company with the startup from the deck"},
        )

    return Company.model_json_schema()


# Make the tests dir importable so the shared helper module (factories.py)
# resolves regardless of pytest's import mode.
if TEST_DIR not in sys.path:
    sys.path.insert(0, TEST_DIR)

import factories  # noqa: E402  (after sys.path setup, like the retab import above)
from retab.types.files import File  # noqa: E402
from retab.types.workflows import Workflow  # noqa: E402


@pytest.fixture
def project_id(sync_client: Retab) -> str:
    """An existing project id to attach creditless resources to.

    The SDK has no projects resource, so we reuse an existing project (every
    workflow carries its owner). Skips cleanly if the org has none.
    """
    pid = factories.discover_project_id(sync_client)
    if not pid:
        pytest.skip("no existing project on staging to attach resources to")
    return pid


@pytest.fixture
def uploaded_file(sync_client: Retab) -> File:
    """A tiny file uploaded to storage (creditless).

    No teardown: the Python SDK exposes no ``files.delete``. Content is a few
    dozen bytes and clearly tagged as test data.
    """
    return factories.upload_file(sync_client)


@pytest.fixture
def temp_workflow(sync_client: Retab, project_id: str) -> Generator[Workflow, None, None]:
    """A freshly-created workflow definition, deleted after the test."""
    with factories.temporary_workflow(sync_client, project_id) as workflow:
        yield workflow


@pytest.fixture
def bad_key_client(api_keys: EnvConfig) -> Generator[Retab, None, None]:
    """A sync client with an invalid API key, for 401/permission assertions."""
    client = factories.junk_key_client(api_keys.retab_api_base_url)
    try:
        yield client
    finally:
        client.close()


@pytest_asyncio.fixture
async def bad_key_async_client(api_keys: EnvConfig) -> AsyncGenerator[AsyncRetab, None]:
    """Async counterpart of :func:`bad_key_client`."""
    client = factories.junk_key_async_client(api_keys.retab_api_base_url)
    try:
        yield client
    finally:
        await client.close()
