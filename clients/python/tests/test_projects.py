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
    assert project.json_schema == booking_confirmation_json_schema
    assert len(project.documents) == 0
    assert len(project.iterations) == 0

    project_id = project.id

    try:
        # READ - Get the evaluation by ID
        retrieved_evaluation = await await_or_return(client.projects.get(project_id))
        assert retrieved_evaluation.id == project_id
        assert retrieved_evaluation.name == evaluation_name

        # LIST - List evaluations
        evaluations = await await_or_return(client.projects.list())
        assert any(e.id == project_id for e in evaluations)

        # UPDATE - Update the evaluation
        updated_name = f"updated_{evaluation_name}"
        updated_evaluation = await await_or_return(
            client.projects.update(
                project_id,
                name=updated_name,
            )
        )
        assert updated_evaluation.name == updated_name

    finally:
        # DELETE - Clean up
        try:
            await await_or_return(client.projects.delete(project_id))
        except Exception:
            pass


@pytest.mark.asyncio
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_evaluation_with_documents(
    sync_client: Retab,
    async_client: AsyncRetab,
    client_type: ClientType,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: Dict[str, Any],
    booking_confirmation_data_1: Dict[str, Any],
    booking_confirmation_data_2: Dict[str, Any],
) -> None:
    """Test evaluation CRUD with documents (no iterations)."""
    evaluation_name = f"test_eval_docs_{nanoid.generate()}"
    client = sync_client if client_type == "sync" else async_client

    # Create an evaluation
    project = await await_or_return(
        client.projects.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
        )
    )

    project_id = project.id

    try:
        # CREATE - Add a document
        document = await await_or_return(
            client.projects.documents.create(
                project_id=project_id,
                document=booking_confirmation_file_path_1,
                annotation=booking_confirmation_data_1,
            )
        )

        assert document.annotation == booking_confirmation_data_1
        document_id = document.id

        # LIST - List documents in the evaluation
        documents = await await_or_return(client.projects.documents.list(project_id))
        assert len(documents) == 1
        assert documents[0].id == document_id

        # UPDATE - Update the document annotation
        # Change the first string value
        updated_document = await await_or_return(
            client.projects.documents.update(
                project_id=project_id,
                document_id=document_id,
                annotation=booking_confirmation_data_2,
            )
        )
        assert updated_document.annotation == booking_confirmation_data_2

        # Verify the evaluation still exists and has the document
        updated_eval = await await_or_return(client.projects.get(project_id))
        assert len(updated_eval.documents) == 1

        # DELETE - Remove the document
        await await_or_return(
            client.projects.documents.delete(
                project_id=project_id,
                document_id=document_id,
            )
        )

        # Verify document was removed
        documents_after = await await_or_return(client.projects.documents.list(project_id))
        assert len(documents_after) == 0

    finally:
        # DELETE - Clean up evaluation
        try:
            await await_or_return(client.projects.delete(project_id))
        except Exception:
            pass


@pytest.mark.asyncio
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_iteration_crud_and_processing(
    sync_client: Retab,
    async_client: AsyncRetab,
    client_type: ClientType,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: Dict[str, Any],
    booking_confirmation_data_1: Dict[str, Any],
) -> None:
    """Test iteration CRUD operations and processing."""
    evaluation_name = f"test_eval_iter_{nanoid.generate()}"
    client = sync_client if client_type == "sync" else async_client

    # First create an evaluation
    project = await await_or_return(
        client.projects.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
        )
    )

    project_id = project.id

    try:
        # Add a document to the evaluation
        document = await await_or_return(
            client.projects.documents.create(
                project_id=project_id,
                document=booking_confirmation_file_path_1,
                annotation=booking_confirmation_data_1,
            )
        )
        assert document.annotation == booking_confirmation_data_1

        # CREATE - Create a new iteration
        iteration = await await_or_return(
            client.projects.iterations.create(
                project_id=project_id,
                model=TEST_MODEL,
                temperature=0.1,
                modality=TEST_MODALITY,
            )
        )

        assert iteration.inference_settings.model == TEST_MODEL
        assert iteration.inference_settings.temperature == 0.1
        assert len(iteration.predictions) == 1
        assert iteration.predictions[document.id].prediction == {}

        iteration_id = iteration.id

        # LIST - List iterations for the evaluation
        iterations = await await_or_return(client.projects.iterations.list(project_id))
        assert any(i.id == iteration_id for i in iterations)

        # STATUS - Check document status
        status_response = await await_or_return(client.projects.iterations.status(project_id, iteration_id))
        assert len(status_response.documents) > 0

        # All documents should need updates initially
        for doc_status in status_response.documents:
            assert doc_status.needs_update is True
            assert doc_status.has_prediction is False

        # PROCESS - Process the iteration (run extractions)
        processed_iteration = await await_or_return(
            client.projects.iterations.process(
                project_id=project_id,
                iteration_id=iteration_id,
                only_outdated=True,
            )
        )

        # After processing, predictions should be populated
        assert len(processed_iteration.predictions) > 0

        # Check status again after processing
        status_after = await await_or_return(client.projects.iterations.status(project_id, iteration_id))
        for doc_status in status_after.documents:
            assert doc_status.has_prediction is True
            assert doc_status.needs_update is False  # Should not need updates anymore

        # DELETE - Clean up iteration
        try:
            await await_or_return(client.projects.iterations.delete(project_id, iteration_id))
        except Exception:
            pass

    finally:
        # DELETE - Clean up evaluation
        try:
            await await_or_return(client.projects.delete(project_id))
        except Exception:
            pass


@pytest.mark.asyncio
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_process_document_method(
    sync_client: Retab,
    async_client: AsyncRetab,
    client_type: ClientType,
    booking_confirmation_file_path_1: str,
    booking_confirmation_json_schema: Dict[str, Any],
    booking_confirmation_data_1: Dict[str, Any],
) -> None:
    """Test process_document method which is used by the frontend."""
    evaluation_name = f"test_eval_process_doc_{nanoid.generate()}"
    client = sync_client if client_type == "sync" else async_client

    # Create an evaluation
    project = await await_or_return(
        client.projects.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
        )
    )

    project_id = project.id

    try:
        # Add a document to the evaluation
        document = await await_or_return(
            client.projects.documents.create(
                project_id=project_id,
                document=booking_confirmation_file_path_1,
                annotation=booking_confirmation_data_1,
            )
        )

        # Create an iteration
        iteration = await await_or_return(
            client.projects.iterations.create(
                project_id=project_id,
                model=TEST_MODEL,
                temperature=0.0,
                modality=TEST_MODALITY,
            )
        )

        iteration_id = iteration.id
        document_id = document.id

        # Check initial status - document should need update
        initial_status = await await_or_return(client.projects.iterations.status(project_id, iteration_id))
        doc_status = next(d for d in initial_status.documents if d.document_id == document_id)
        assert doc_status.needs_update is True
        assert doc_status.has_prediction is False

        # PROCESS_DOCUMENT - Process a single document (frontend method)
        completion_response = await await_or_return(
            client.projects.iterations.process_document(
                project_id=project_id,
                iteration_id=iteration_id,
                document_id=document_id,
            )
        )

        # Validate the response
        assert isinstance(completion_response, RetabParsedChatCompletion)
        assert completion_response.choices is not None
        assert len(completion_response.choices) > 0
        assert completion_response.choices[0].message.content is not None

        # Verify that the parsed content is valid JSON
        try:
            parsed_content = json.loads(completion_response.choices[0].message.content)
            assert isinstance(parsed_content, dict)
        except json.JSONDecodeError:
            assert False, "Response content should be valid JSON"

        # Check status after processing - document should be updated
        final_status = await await_or_return(client.projects.iterations.status(project_id, iteration_id))
        doc_status_after = next(d for d in final_status.documents if d.document_id == document_id)
        assert doc_status_after.has_prediction is True
        assert doc_status_after.needs_update is False

    finally:
        # Clean up
        try:
            await await_or_return(client.projects.delete(project_id))
        except Exception:
            pass


@pytest.mark.asyncio
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_complete_evaluation_workflow(
    sync_client: Retab,
    async_client: AsyncRetab,
    client_type: ClientType,
    booking_confirmation_file_path_1: str,
    booking_confirmation_file_path_2: str,
    booking_confirmation_json_schema: Dict[str, Any],
    booking_confirmation_data_1: Dict[str, Any],
    booking_confirmation_data_2: Dict[str, Any],
) -> None:
    """Test complete workflow: evaluation + 2 documents + 2 iterations + processing + cleanup."""
    evaluation_name = f"test_eval_complete_{nanoid.generate()}"
    client = sync_client if client_type == "sync" else async_client

    # Step 1: Create an evaluation
    project = await await_or_return(
        client.projects.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
        )
    )

    project_id = project.id

    try:
        # Step 2: Add 2 documents
        doc1 = await await_or_return(
            client.projects.documents.create(
                project_id=project_id,
                document=booking_confirmation_file_path_1,
                annotation=booking_confirmation_data_1,
            )
        )

        # Change the first string value
        doc2 = await await_or_return(
            client.projects.documents.create(
                project_id=project_id,
                document=booking_confirmation_file_path_2,
                annotation=booking_confirmation_data_2,
            )
        )

        # Verify we have 2 documents
        documents = await await_or_return(client.projects.documents.list(project_id))
        assert len(documents) == 2

        # Step 3: Create 2 iterations
        iteration1 = await await_or_return(
            client.projects.iterations.create(
                project_id=project_id,
                model=TEST_MODEL,
                temperature=0.0,
                modality=TEST_MODALITY,
            )
        )

        iteration2 = await await_or_return(
            client.projects.iterations.create(
                project_id=project_id,
                model=TEST_MODEL,
                temperature=0.5,
                modality=TEST_MODALITY,
            )
        )

        # Verify we have 2 iterations
        iterations = await await_or_return(client.projects.iterations.list(project_id))
        assert len(iterations) == 2

        # Step 4: Process using both process methods
        # 4a: Use process method (bulk processing)
        processed_iter1 = await await_or_return(
            client.projects.iterations.process(
                project_id=project_id,
                iteration_id=iteration1.id,
                only_outdated=True,
            )
        )

        # 4b: Use process_document method (individual document processing)
        completion1 = await await_or_return(
            client.projects.iterations.process_document(
                project_id=project_id,
                iteration_id=iteration2.id,
                document_id=doc1.id,
            )
        )
        completion2 = await await_or_return(
            client.projects.iterations.process_document(
                project_id=project_id,
                iteration_id=iteration2.id,
                document_id=doc2.id,
            )
        )

        # Verify both iterations have predictions
        assert len(processed_iter1.predictions) > 0

        # Verify process_document responses
        assert isinstance(completion1, RetabParsedChatCompletion)
        assert isinstance(completion2, RetabParsedChatCompletion)
        assert completion1.choices[0].message.content is not None
        assert completion2.choices[0].message.content is not None

        # Step 5: Delete one iteration
        await await_or_return(client.projects.iterations.delete(project_id, iteration2.id))

        # Verify we now have only 1 iteration
        iterations_after_delete = await await_or_return(client.projects.iterations.list(project_id))
        assert len(iterations_after_delete) == 1
        assert iterations_after_delete[0].id == iteration1.id

        # Step 6: Delete one document (should affect remaining iteration)
        await await_or_return(client.projects.documents.delete(project_id=project_id, document_id=doc2.id))

        # Verify we now have only 1 document
        documents_after_delete = await await_or_return(client.projects.documents.list(project_id))
        assert len(documents_after_delete) == 1
        assert documents_after_delete[0].id == doc1.id

        # Step 7: Check status of remaining iteration (should reflect document deletion)
        final_status = await await_or_return(client.projects.iterations.status(project_id, iteration1.id))
        assert len(final_status.documents) == 1  # Only one document should remain

    finally:
        # Cleanup - Delete evaluation (should cascade to remaining documents and iterations)
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

@pytest.mark.asyncio
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_iteration_selective_processing(
    sync_client: Retab,
    async_client: AsyncRetab,
    client_type: ClientType,
    booking_confirmation_file_path_1: str,
    booking_confirmation_file_path_2: str,
    booking_confirmation_json_schema: Dict[str, Any],
    booking_confirmation_data_1: Dict[str, Any],
    booking_confirmation_data_2: Dict[str, Any],
) -> None:
    """Test selective processing of specific documents in an iteration."""
    evaluation_name = f"test_eval_selective_{nanoid.generate()}"
    client = sync_client if client_type == "sync" else async_client

    # Create an evaluation
    project = await await_or_return(
        client.projects.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
        )
    )

    project_id = project.id

    try:
        # Add multiple documents
        doc1 = await await_or_return(
            client.projects.documents.create(
                project_id=project_id,
                document=booking_confirmation_file_path_1,
                annotation=booking_confirmation_data_1,
            )
        )

        # Change the first string value
        doc2 = await await_or_return(
            client.projects.documents.create(
                project_id=project_id,
                document=booking_confirmation_file_path_2,
                annotation=booking_confirmation_data_2,
            )
        )

        # Create an iteration
        iteration = await await_or_return(
            client.projects.iterations.create(
                project_id=project_id,
                model=TEST_MODEL,
                temperature=0.0,
                modality=TEST_MODALITY,
            )
        )

        iteration_id = iteration.id

        # Process only the first document using process method
        await await_or_return(
            client.projects.iterations.process(
                project_id=project_id,
                iteration_id=iteration_id,
                document_ids=[doc1.id],
                only_outdated=False,
            )
        )

        # Check that only one document was processed via process method
        status_response = await await_or_return(client.projects.iterations.status(project_id, iteration_id))
        doc1_status = next(d for d in status_response.documents if d.document_id == doc1.id)
        doc2_status = next(d for d in status_response.documents if d.document_id == doc2.id)

        assert doc1_status.has_prediction is True
        assert doc2_status.has_prediction is False
        assert doc1_status.needs_update is False
        assert doc2_status.needs_update is True

        # Now process the second document using process_document method
        completion_response = await await_or_return(
            client.projects.iterations.process_document(
                project_id=project_id,
                iteration_id=iteration_id,
                document_id=doc2.id,
            )
        )

        # Verify the response
        assert isinstance(completion_response, RetabParsedChatCompletion)
        assert completion_response.choices[0].message.content is not None

        # Check that both documents are now processed
        final_status = await await_or_return(client.projects.iterations.status(project_id, iteration_id))
        doc1_final = next(d for d in final_status.documents if d.document_id == doc1.id)
        doc2_final = next(d for d in final_status.documents if d.document_id == doc2.id)

        assert doc1_final.has_prediction is True
        assert doc2_final.has_prediction is True
        assert doc1_final.needs_update is False
        assert doc2_final.needs_update is False

        # DELETE - Clean up iteration
        try:
            await await_or_return(client.projects.iterations.delete(project_id, iteration_id))
        except Exception:
            pass

    finally:
        # DELETE - Clean up evaluation
        try:
            await await_or_return(client.projects.delete(project_id))
        except Exception:
            pass


# FAILED test_evaluations.py::test_complete_evaluation_workflow[sync] - RuntimeError: Request failed (409): {"detail":{"code":"HTTP_EXCEPTION","message":"An HTTP exception occurred.","details":{"error":"Document with this ID already exists in the e...
# FAILED test_evaluations.py::test_evaluation_with_documents[sync] - Exception: Max tries exceeded after 1 tries.
