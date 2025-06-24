import json
import warnings
import os
import shutil
from typing import IO, Any

import pytest

os.environ["EMAIL_DOMAIN"] = "mailbox.retab.dev"
from enum import Enum
from typing import Generator

from dotenv import load_dotenv
from pydantic import BaseModel, Field

from retab import AsyncRetab, Retab

# Get the directory containing the tests
TEST_DIR = os.path.dirname(os.path.abspath(__file__))


def pytest_addoption(parser: pytest.Parser) -> None:
    parser.addoption("--production", action="store_true", default=False, help="run tests against production API")
    parser.addoption("--local", action="store_true", default=False, help="run tests against local API")
    parser.addoption("--staging", action="store_true", default=False, help="run tests against staging API")
    parser.addoption("--env-file", type=str, help="path to the .env file to use")


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
        raise ValueError("No environment specified. Please use --env-file, --production, --local, or --staging.")

    print("loading env file: ", env_path)
    if not os.path.exists(env_path):
        warnings.warn(f"Environment file not found: {env_path}", UserWarning)
    else:
        load_dotenv(env_path, override=True)
    print("EMAIL_DOMAIN", os.environ["EMAIL_DOMAIN"])


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
        retab_api_base_url=retab_api_base_url,
    )


@pytest.fixture(scope="session")
def base_url(api_keys: EnvConfig) -> str:
    return api_keys.retab_api_base_url


@pytest.fixture(scope="session")
def retab_api_key(api_keys: EnvConfig) -> str:
    return api_keys.retab_api_key


@pytest.fixture(scope="function")
def sync_client(api_keys: EnvConfig) -> Retab:
    return Retab(
        api_key=api_keys.retab_api_key,
        base_url=api_keys.retab_api_base_url,
        max_retries=0,
    )


@pytest.fixture(scope="function")
def async_client(api_keys: EnvConfig) -> AsyncRetab:
    return AsyncRetab(
        api_key=api_keys.retab_api_key,
        base_url=api_keys.retab_api_base_url,
        max_retries=0,
    )


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
