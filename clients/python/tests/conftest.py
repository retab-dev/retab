import pytest
import os
import json
import shutil
from typing import IO, Any
from uiform import UiForm, AsyncUiForm
from typing import Generator

from dotenv import load_dotenv

load_dotenv()

uiform_api_key = os.getenv("UIFORM_API_KEY")
uiform_base_url = os.getenv("UIFORM_API_BASE_URL")
openai_api_key = os.getenv("OPENAI_API_KEY")
claude_api_key = os.getenv("CLAUDE_API_KEY")
gemini_api_key = os.getenv("GEMINI_API_KEY")
xai_api_key = os.getenv("XAI_API_KEY")

# Fixture to create clients
@pytest.fixture
def sync_client() -> UiForm:
    return UiForm(
        api_key=uiform_api_key,
        base_url=uiform_base_url,
        openai_api_key=openai_api_key,
        claude_api_key=claude_api_key,
        gemini_api_key=gemini_api_key,
        xai_api_key=xai_api_key,
        max_retries=2
    )

@pytest.fixture
def async_client() -> AsyncUiForm:
    return AsyncUiForm(
        api_key=uiform_api_key,
        base_url=uiform_base_url,
        openai_api_key=openai_api_key,
        claude_api_key=claude_api_key,
        gemini_api_key=gemini_api_key,
        xai_api_key=xai_api_key,
        max_retries=2
    )

# Get the directory containing the tests
TEST_DIR = os.path.dirname(os.path.abspath(__file__))

def pytest_addoption(parser: pytest.Parser) -> None:
    parser.addoption(
        "--production",
        action="store_true",
        default=False,
        help="run tests against production API"
    )

@pytest.fixture(scope="session")
def base_url(request: pytest.FixtureRequest) -> str:
    if request.config.getoption("--production"):
        return "https://api.uiform.com"
    elif request.config.getoption("--local"):
        return "http://localhost:4000"
    else:
        raise ValueError("No environment specified. Please use --production or --local.")

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
def booking_confirmation_file_path(test_data_dir: str) -> str:
    return os.path.join(test_data_dir, "freight", "booking_confirmation.jpg")

@pytest.fixture
def booking_confirmation_bytes(booking_confirmation_file_path: str) -> bytes:   # Not Working!
    with open(booking_confirmation_file_path, "rb") as f:
        return f.read()

@pytest.fixture
def booking_confirmation_io_bytes(booking_confirmation_file_path: str) -> IO[bytes]:    # Not Working!
    with open(booking_confirmation_file_path, "rb") as f:
        return f

