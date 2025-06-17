import asyncio
import json
from typing import Any, Dict, List

import nanoid  # type: ignore
import pytest

from uiform import AsyncUiForm, UiForm
from uiform.types.evaluations import Evaluation, Iteration, EvaluationDocument, DocumentStatus
from uiform.types.inference_settings import InferenceSettings


@pytest.mark.asyncio
async def test_sync_evaluation_crud(
    sync_client: UiForm,
    booking_confirmation_file_path: str,
    booking_confirmation_json_schema: Dict[str, Any],
) -> None:
    """Test CRUD operations for evaluations using sync client."""
    evaluation_name = f"test_eval_{nanoid.generate()}"
    
    with sync_client as client:
        # CREATE - Create a new evaluation
        evaluation = client.evaluations.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
            project_id="test_project",
        )
        
        assert evaluation.name == evaluation_name
        assert evaluation.json_schema == booking_confirmation_json_schema
        assert evaluation.project_id == "test_project"
        assert len(evaluation.documents) == 0
        assert len(evaluation.iterations) == 0
        
        evaluation_id = evaluation.id
        
        try:
            # READ - Get the evaluation by ID
            retrieved_evaluation = client.evaluations.get(evaluation_id)
            assert retrieved_evaluation.id == evaluation_id
            assert retrieved_evaluation.name == evaluation_name
            
            # LIST - List evaluations
            evaluations = client.evaluations.list(project_id="test_project")
            assert any(e.id == evaluation_id for e in evaluations)
            
            # UPDATE - Update the evaluation
            updated_name = f"updated_{evaluation_name}"
            updated_evaluation = client.evaluations.update(
                evaluation_id,
                name=updated_name,
                project_id="updated_project",
            )
            assert updated_evaluation.name == updated_name
            assert updated_evaluation.project_id == "updated_project"
            
        finally:
            # DELETE - Clean up
            try:
                client.evaluations.delete(evaluation_id)
            except Exception:
                pass


@pytest.mark.asyncio 
async def test_async_evaluation_crud(
    async_client: AsyncUiForm,
    booking_confirmation_file_path: str,
    booking_confirmation_json_schema: Dict[str, Any],
) -> None:
    """Test CRUD operations for evaluations using async client."""
    evaluation_name = f"test_eval_async_{nanoid.generate()}"
    
    async with async_client as client:
        # CREATE - Create a new evaluation
        evaluation = await client.evaluations.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
            project_id="test_project_async",
        )
        
        assert evaluation.name == evaluation_name
        assert evaluation.json_schema == booking_confirmation_json_schema
        assert evaluation.project_id == "test_project_async"
        
        evaluation_id = evaluation.id
        
        try:
            # READ - Get the evaluation by ID
            retrieved_evaluation = await client.evaluations.get(evaluation_id)
            assert retrieved_evaluation.id == evaluation_id
            assert retrieved_evaluation.name == evaluation_name
            
            # LIST - List evaluations
            evaluations = await client.evaluations.list(project_id="test_project_async")
            assert any(e.id == evaluation_id for e in evaluations)
            
            # UPDATE - Update the evaluation
            updated_name = f"updated_{evaluation_name}"
            updated_evaluation = await client.evaluations.update(
                evaluation_id,
                name=updated_name,
            )
            assert updated_evaluation.name == updated_name
            
        finally:
            # DELETE - Clean up
            try:
                await client.evaluations.delete(evaluation_id)
            except Exception:
                pass


@pytest.mark.asyncio
async def test_sync_iteration_crud_and_processing(
    sync_client: UiForm,
    booking_confirmation_file_path: str,
    booking_confirmation_json_schema: Dict[str, Any],
) -> None:
    """Test iteration CRUD operations and processing using sync client."""
    evaluation_name = f"test_eval_iter_{nanoid.generate()}"
    
    with sync_client as client:
        # First create an evaluation
        evaluation = client.evaluations.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
            project_id="test_project",
        )
        
        evaluation_id = evaluation.id
        
        try:
            # Add a document to the evaluation
            document = client.evaluations.documents.add(
                evaluation_id=evaluation_id,
                file_path=booking_confirmation_file_path,
                annotation={"test": "annotation"},
            )
            assert document.annotation == {"test": "annotation"}
            
            # CREATE - Create a new iteration
            iteration = client.evaluations.iterations.create(
                evaluation_id=evaluation_id,
                model="gpt-4o-mini",
                temperature=0.1,
                modality="native",
            )
            
            assert iteration.inference_settings.model == "gpt-4o-mini"
            assert iteration.inference_settings.temperature == 0.1
            assert len(iteration.predictions) == 0  # Should be empty initially
            
            iteration_id = iteration.id
            
            # LIST - List iterations for the evaluation
            iterations = client.evaluations.iterations.list(evaluation_id)
            assert any(i.id == iteration_id for i in iterations)
            
            # STATUS - Check document status
            status_response = client.evaluations.iterations.status(iteration_id)
            assert len(status_response.documents) > 0
            
            # All documents should need updates initially
            for doc_status in status_response.documents:
                assert doc_status.needs_update == True
                assert doc_status.has_prediction == False
            
            # PROCESS - Process the iteration (run extractions)
            processed_iteration = client.evaluations.iterations.process(
                iteration_id=iteration_id,
                only_outdated=True,
            )
            
            # After processing, predictions should be populated
            assert len(processed_iteration.predictions) > 0
            
            # Check status again after processing
            status_after = client.evaluations.iterations.status(iteration_id)
            for doc_status in status_after.documents:
                assert doc_status.has_prediction == True
                assert doc_status.needs_update == False  # Should not need updates anymore
            
            # DELETE - Clean up iteration
            try:
                client.evaluations.iterations.delete(iteration_id)
            except Exception:
                pass
                
        finally:
            # DELETE - Clean up evaluation
            try:
                client.evaluations.delete(evaluation_id)
            except Exception:
                pass


@pytest.mark.asyncio
async def test_async_iteration_crud_and_processing(
    async_client: AsyncUiForm,
    booking_confirmation_file_path: str,
    booking_confirmation_json_schema: Dict[str, Any],
) -> None:
    """Test iteration CRUD operations and processing using async client."""
    evaluation_name = f"test_eval_iter_async_{nanoid.generate()}"
    
    async with async_client as client:
        # First create an evaluation
        evaluation = await client.evaluations.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
            project_id="test_project_async",
        )
        
        evaluation_id = evaluation.id
        
        try:
            # Add a document to the evaluation
            document = await client.evaluations.documents.add(
                evaluation_id=evaluation_id,
                file_path=booking_confirmation_file_path,
                annotation={"test": "annotation"},
            )
            assert document.annotation == {"test": "annotation"}
            
            # CREATE - Create a new iteration
            iteration = await client.evaluations.iterations.create(
                evaluation_id=evaluation_id,
                model="gpt-4o-mini",
                temperature=0.1,
                modality="native",
            )
            
            assert iteration.inference_settings.model == "gpt-4o-mini"
            assert iteration.inference_settings.temperature == 0.1
            assert len(iteration.predictions) == 0  # Should be empty initially
            
            iteration_id = iteration.id
            
            # LIST - List iterations for the evaluation
            iterations = await client.evaluations.iterations.list(evaluation_id)
            assert any(i.id == iteration_id for i in iterations)
            
            # STATUS - Check document status
            status_response = await client.evaluations.iterations.status(iteration_id)
            assert len(status_response.documents) > 0
            
            # All documents should need updates initially
            for doc_status in status_response.documents:
                assert doc_status.needs_update == True
                assert doc_status.has_prediction == False
            
            # PROCESS - Process the iteration (run extractions)
            processed_iteration = await client.evaluations.iterations.process(
                iteration_id=iteration_id,
                only_outdated=True,
            )
            
            # After processing, predictions should be populated
            assert len(processed_iteration.predictions) > 0
            
            # Check status again after processing
            status_after = await client.evaluations.iterations.status(iteration_id)
            for doc_status in status_after.documents:
                assert doc_status.has_prediction == True
                assert doc_status.needs_update == False  # Should not need updates anymore
            
            # DELETE - Clean up iteration
            try:
                await client.evaluations.iterations.delete(iteration_id)
            except Exception:
                pass
                
        finally:
            # DELETE - Clean up evaluation
            try:
                await client.evaluations.delete(evaluation_id)
            except Exception:
                pass


@pytest.mark.asyncio
async def test_document_operations(
    sync_client: UiForm,
    booking_confirmation_file_path: str,
    booking_confirmation_json_schema: Dict[str, Any],
) -> None:
    """Test document operations within evaluations."""
    evaluation_name = f"test_eval_docs_{nanoid.generate()}"
    
    with sync_client as client:
        # Create an evaluation
        evaluation = client.evaluations.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
            project_id="test_project",
        )
        
        evaluation_id = evaluation.id
        
        try:
            # ADD - Add a document
            document = client.evaluations.documents.add(
                evaluation_id=evaluation_id,
                file_path=booking_confirmation_file_path,
                annotation={"invoice_number": "INV-001", "amount": 100.0},
            )
            
            assert document.annotation == {"invoice_number": "INV-001", "amount": 100.0}
            document_id = document.id
            
            # LIST - List documents in the evaluation
            documents = client.evaluations.documents.list(evaluation_id)
            assert len(documents) == 1
            assert documents[0].id == document_id
            
            # UPDATE - Update the document annotation
            updated_document = client.evaluations.documents.update(
                evaluation_id=evaluation_id,
                document_id=document_id,
                annotation={"invoice_number": "INV-002", "amount": 200.0},
            )
            assert updated_document.annotation == {"invoice_number": "INV-002", "amount": 200.0}
            
            # REMOVE - Remove the document
            client.evaluations.documents.remove(
                evaluation_id=evaluation_id,
                document_id=document_id,
            )
            
            # Verify document was removed
            documents_after = client.evaluations.documents.list(evaluation_id)
            assert len(documents_after) == 0
            
        finally:
            # DELETE - Clean up evaluation
            try:
                client.evaluations.delete(evaluation_id)
            except Exception:
                pass


@pytest.mark.asyncio
async def test_iteration_selective_processing(
    sync_client: UiForm,
    booking_confirmation_file_path: str,
    booking_confirmation_json_schema: Dict[str, Any],
) -> None:
    """Test selective processing of specific documents in an iteration."""
    evaluation_name = f"test_eval_selective_{nanoid.generate()}"
    
    with sync_client as client:
        # Create an evaluation
        evaluation = client.evaluations.create(
            name=evaluation_name,
            json_schema=booking_confirmation_json_schema,
            project_id="test_project",
        )
        
        evaluation_id = evaluation.id
        
        try:
            # Add multiple documents
            doc1 = client.evaluations.documents.add(
                evaluation_id=evaluation_id,
                file_path=booking_confirmation_file_path,
                annotation={"test": "doc1"},
            )
            
            doc2 = client.evaluations.documents.add(
                evaluation_id=evaluation_id,
                file_path=booking_confirmation_file_path,
                annotation={"test": "doc2"},
            )
            
            # Create an iteration
            iteration = client.evaluations.iterations.create(
                evaluation_id=evaluation_id,
                model="gpt-4o-mini",
                temperature=0.0,
                modality="native",
            )
            
            iteration_id = iteration.id
            
            # Process only the first document
            processed_iteration = client.evaluations.iterations.process(
                iteration_id=iteration_id,
                document_ids=[doc1.id],
                only_outdated=False,
            )
            
            # Check that only one document was processed
            status_response = client.evaluations.iterations.status(iteration_id)
            doc1_status = next(d for d in status_response.documents if d.document_id == doc1.id)
            doc2_status = next(d for d in status_response.documents if d.document_id == doc2.id)
            
            assert doc1_status.has_prediction == True
            assert doc2_status.has_prediction == False
            
            # DELETE - Clean up iteration
            try:
                client.evaluations.iterations.delete(iteration_id)
            except Exception:
                pass
                
        finally:
            # DELETE - Clean up evaluation
            try:
                client.evaluations.delete(evaluation_id)
            except Exception:
                pass