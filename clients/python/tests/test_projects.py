import json
from typing import Any, Awaitable, Dict, Literal, TypeVar, get_args

import nanoid  # type: ignore
import pytest

from retab import AsyncRetab, Retab
from retab.types.documents.extract import RetabParsedChatCompletion

T = TypeVar("T")

# Global test constants
TEST_MODEL = "gpt-4.1-nano"
TEST_MODALITY = "native"


async def await_or_return(obj: T | Awaitable[T]) -> T:
    """
    Await an object if it is an awaitable, otherwise return it.
    """
    if isinstance(obj, Awaitable):
        return await obj
    else:
        return obj


ClientType = Literal[
    "sync",
    "async",
]


@pytest.mark.asyncio
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_evaluation_crud_basic(
    sync_client: Retab,
    async_client: AsyncRetab,
    client_type: ClientType,
    booking_confirmation_json_schema: Dict[str, Any],
) -> None:
    """Test basic CRUD operations for evaluations (no documents or iterations)."""
    evaluation_name = f"test_eval_basic_{nanoid.generate()}"
    client = sync_client if client_type == "sync" else async_client

    # CREATE - Create a new evaluation
    project = await await_or_return(client.projects.create(name=evaluation_name, json_schema=booking_confirmation_json_schema))

    assert project.name == evaluation_name
    assert project.draft_config.json_schema == booking_confirmation_json_schema

    project_id = project.id

    try:
        # READ - Get the evaluation by ID
        retrieved_evaluation = await await_or_return(client.projects.get(project_id))
        assert retrieved_evaluation.id == project_id
        assert retrieved_evaluation.name == evaluation_name

        # PUBLISH - Promote draft configuration
        published_project = await await_or_return(
            client.projects.publish(project_id)  # type: ignore[attr-defined]
        )
        assert published_project.id == project_id

        # LIST - List evaluations
        evaluations = await await_or_return(client.projects.list())
        assert any(e.id == project_id for e in evaluations)

    finally:
        # DELETE - Clean up
        try:
            await await_or_return(client.projects.delete(project_id))
        except Exception:
            pass


@pytest.mark.asyncio
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_extract_without_iteration_id(
    sync_client: Retab,
    async_client: AsyncRetab,
    client_type: ClientType,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: Dict[str, Any],
) -> None:
    """Ensure extract works when iteration_id is omitted (defaults to base configuration)."""
    evaluation_name = f"test_extract_no_iter_{nanoid.generate()}"
    client = sync_client if client_type == "sync" else async_client

    # Create a project
    project = await await_or_return(
        client.projects.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
        )
    )

    project_id = project.id

    # PUBLISH - Promote draft configuration
    published_project = await await_or_return(
        client.projects.publish(project_id)  # type: ignore[attr-defined]
    )
    assert published_project.id == project_id

    try:
        # Call extract without providing iteration_id
        completion_response = await await_or_return(
            client.projects.extract(
                project_id=project_id,
                document=booking_confirmation_file_path_1,
            )
        )

        # Validate the response
        assert isinstance(completion_response, RetabParsedChatCompletion)
        assert completion_response.choices is not None
        assert len(completion_response.choices) > 0
        assert completion_response.choices[0].message.content is not None

        # The parsed content should be valid JSON
        try:
            parsed_content = json.loads(completion_response.choices[0].message.content)
            assert isinstance(parsed_content, dict)
        except json.JSONDecodeError:
            assert False, "Response content should be valid JSON"

    finally:
        # Cleanup
        try:
            await await_or_return(client.projects.delete(project_id))
        except Exception:
            pass



# FAILED test_evaluations.py::test_complete_evaluation_workflow[sync] - RuntimeError: Request failed (409): {"detail":{"code":"HTTP_EXCEPTION","message":"An HTTP exception occurred.","details":{"error":"Document with this ID already exists in the e...
# FAILED test_evaluations.py::test_evaluation_with_documents[sync] - Exception: Max tries exceeded after 1 tries.
