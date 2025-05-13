from typing import Any, Dict, List, Optional, TypedDict

from .._resource import AsyncAPIResource, SyncAPIResource
from ..types.standards import PreparedRequest
from ..types.evals import Evaluation, EvaluationDocument, Iteration, MetricResult



class EvalsMixin:
    def prepare_create(self, name: str, json_schema: Dict[str, Any], project_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url="/v1/evals",
            data={
                "name": name,
                "json_schema": json_schema,
                "project_id": project_id,
                "documents": [],
                "iterations": []
            }
        )

    def prepare_get(self, id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/v1/evals/{id}"
        )

    def prepare_update(self, id: str, name: str) -> PreparedRequest:
        return PreparedRequest(
            method="PUT",
            url=f"/v1/evals/{id}",
            data={"name": name}
        )

    def prepare_list(self, project_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url="/v1/evals",
            params={"project_id": project_id}
        )

    def prepare_delete(self, id: str) -> PreparedRequest:
        return PreparedRequest(
            method="DELETE",
            url=f"/v1/evals/{id}"
        )


class DocumentsMixin:
    def prepare_import_jsonl(self, eval_id: str, path: str) -> PreparedRequest:
        # This would be implemented based on the actual API endpoint
        return PreparedRequest(
            method="POST",
            url=f"/v1/evals/{eval_id}/import_documents",
            data={"path": path}
        )

    def prepare_save_to_jsonl(self, eval_id: str, path: str) -> PreparedRequest:
        # This would be implemented based on the actual API endpoint
        return PreparedRequest(
            method="POST",
            url=f"/v1/evals/{eval_id}/export_documents",
            data={"path": path}
        )

    def prepare_create(self, eval_id: str, document: str, ground_truth: Dict[str, Any]) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"/v1/evals/{eval_id}/documents",
            data={"document": document, "ground_truth": ground_truth}
        )

    def prepare_list(self, eval_id: str, filename: Optional[str] = None) -> PreparedRequest:
        params = {}
        if filename:
            params["filename"] = filename
        return PreparedRequest(
            method="GET",
            url=f"/v1/evals/{eval_id}/documents",
            params=params
        )


class DocumentMixin:
    def prepare_update(self, eval_id: str, id: str, ground_truth: Dict[str, Any]) -> PreparedRequest:
        return PreparedRequest(
            method="PUT",
            url=f"/v1/evals/{eval_id}/documents/{id}",
            data={"ground_truth": ground_truth}
        )

    def prepare_delete(self, eval_id: str, id: str) -> PreparedRequest:
        return PreparedRequest(
            method="DELETE",
            url=f"/v1/evals/{eval_id}/documents/{id}"
        )


class IterationsMixin:
    def prepare_import_jsonl(self, eval_id: str, path: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"/v1/evals/{eval_id}/add_iteration_from_jsonl",
            data={"jsonl_gcs_path": path}
        )

    def prepare_save_to_jsonl(self, eval_id: str, path: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"/v1/evals/{eval_id}/export_iteration_as_jsonl/0",
            data={"path": path}
        )

    def prepare_get(self, id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/v1/iterations/{id}"
        )

    def prepare_list(self, eval_id: str, model: Optional[str] = None) -> PreparedRequest:
        params = {}
        if model:
            params["model"] = model
        return PreparedRequest(
            method="GET",
            url=f"/v1/evals/{eval_id}/iterations",
            params=params
        )

    def prepare_create(self, eval_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> PreparedRequest:
        data = {
            "json_schema": json_schema,
            "annotation_props": {
                "model": model,
                "temperature": temperature
            }
        }
        if image_settings:
            data["annotation_props"]["image_settings"] = image_settings
        return PreparedRequest(
            method="POST",
            url=f"/v1/evals/{eval_id}/iterations",
            data=data
        )

    def prepare_update(self, iteration_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> PreparedRequest:
        data = {
            "json_schema": json_schema,
            "annotation_props": {
                "model": model,
                "temperature": temperature
            }
        }
        if image_settings:
            data["annotation_props"]["image_settings"] = image_settings
        return PreparedRequest(
            method="PUT",
            url=f"/v1/iterations/{iteration_id}",
            data=data
        )

    def prepare_delete(self, id: str) -> PreparedRequest:
        return PreparedRequest(
            method="DELETE",
            url=f"/v1/iterations/{id}"
        )


class DistancesMixin:
    def prepare_get(self, iteration_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/v1/iterations/{iteration_id}/distances/{document_id}"
        )


class DeleteResponse(TypedDict):
    """Response from a delete operation"""
    success: bool
    id: str

class ExportResponse(TypedDict):
    """Response from an export operation"""
    success: bool
    path: str


class Evals(SyncAPIResource, EvalsMixin):
    """Evals API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.documents = Documents(self._client)
        self.document = Document(self._client)
        self.iterations = Iterations(self._client)

    def create(self, name: str, json_schema: Dict[str, Any], project_id: str) -> Evaluation:
        """
        Create a new evaluation.

        Args:
            name: The name of the evaluation
            json_schema: The JSON schema for the evaluation
            project_id: The project ID to associate with the evaluation

        Returns:
            Evaluation: The created evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(name, json_schema, project_id)
        response = self._client._prepared_request(request)
        return Evaluation(**response)

    def get(self, id: str) -> Evaluation:
        """
        Get an evaluation by ID.

        Args:
            id: The ID of the evaluation to retrieve

        Returns:
            Evaluation: The evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(id)
        response = self._client._prepared_request(request)
        return Evaluation(**response)

    def update(self, id: str, name: str) -> Evaluation:
        """
        Update an evaluation.

        Args:
            id: The ID of the evaluation to update
            name: The new name for the evaluation

        Returns:
            Evaluation: The updated evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(id, name)
        response = self._client._prepared_request(request)
        return Evaluation(**response)

    def list(self, project_id: str) -> List[Evaluation]:
        """
        List evaluations for a project.

        Args:
            project_id: The project ID to list evaluations for

        Returns:
            List[Evaluation]: List of evaluations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(project_id)
        response = self._client._prepared_request(request)
        return [Evaluation(**item) for item in response.get("data", [])]

    def delete(self, id: str) -> DeleteResponse:
        """
        Delete an evaluation.

        Args:
            id: The ID of the evaluation to delete

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(id)
        return self._client._prepared_request(request)


class Documents(SyncAPIResource, DocumentsMixin):
    """Documents API wrapper for evaluations"""

    def import_jsonl(self, eval_id: str, path: str) -> Evaluation:
        """
        Import documents from a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to the JSONL file

        Returns:
            Evaluation: The updated experiment with imported documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_import_jsonl(eval_id, path)
        response = self._client._prepared_request(request)
        return Evaluation(**response)

    def save_to_jsonl(self, eval_id: str, path: str) -> ExportResponse:
        """
        Save documents to a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to save the JSONL file

        Returns:
            ExportResponse: The response containing success status and path
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_save_to_jsonl(eval_id, path)
        return self._client._prepared_request(request)

    def create(self, eval_id: str, document: str, ground_truth: Dict[str, Any]) -> EvaluationDocument:
        """
        Create a document for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            document: The document file path or content
            ground_truth: The ground truth for the document

        Returns:
            EvaluationDocument: The created document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(eval_id, document, ground_truth)
        response = self._client._prepared_request(request)
        return EvaluationDocument(**response)

    def list(self, eval_id: str, filename: Optional[str] = None) -> List[EvaluationDocument]:
        """
        List documents for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            filename: Optional filename to filter by

        Returns:
            List[EvaluationDocument]: List of documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(eval_id, filename)
        response = self._client._prepared_request(request)
        return [EvaluationDocument(**item) for item in response.get("data", [])]


class Document(SyncAPIResource, DocumentMixin):
    """Document API wrapper for individual document operations"""

    def update(self, eval_id: str, id: str, ground_truth: Dict[str, Any]) -> EvaluationDocument:
        """
        Update a document.

        Args:
            eval_id: The ID of the evaluation
            id: The ID of the document
            ground_truth: The ground truth for the document

        Returns:
            EvaluationDocument: The updated document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(eval_id, id, ground_truth)
        response = self._client._prepared_request(request)
        return EvaluationDocument(**response)

    def delete(self, eval_id: str, id: str) -> DeleteResponse:
        """
        Delete a document.

        Args:
            eval_id: The ID of the evaluation
            id: The ID of the document

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(eval_id, id)
        return self._client._prepared_request(request)


class Iterations(SyncAPIResource, IterationsMixin):
    """Iterations API wrapper for evaluations"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.distances = Distances(self._client)

    def import_jsonl(self, eval_id: str, path: str) -> Evaluation:
        """
        Import iterations from a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to the JSONL file

        Returns:
            Evaluation: The updated experiment with imported iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_import_jsonl(eval_id, path)
        response = self._client._prepared_request(request)
        return Evaluation(**response)

    def save_to_jsonl(self, eval_id: str, path: str) -> ExportResponse:
        """
        Save iterations to a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to save the JSONL file

        Returns:
            ExportResponse: The response containing success status and path
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_save_to_jsonl(eval_id, path)
        return self._client._prepared_request(request)

    def get(self, id: str) -> Iteration:
        """
        Get an iteration by ID.

        Args:
            id: The ID of the iteration

        Returns:
            Iteration: The iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(id)
        response = self._client._prepared_request(request)
        return Iteration(**response)

    def list(self, eval_id: str, model: Optional[str] = None) -> List[Iteration]:
        """
        List iterations for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            model: Optional model to filter by

        Returns:
            List[Iteration]: List of iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(eval_id, model)
        response = self._client._prepared_request(request)
        return [Iteration(**item) for item in response.get("data", [])]

    def create(self, eval_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> Iteration:
        """
        Create a new iteration for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            json_schema: The JSON schema for the iteration
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            image_settings: Optional image settings

        Returns:
            Iteration: The created iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(eval_id, json_schema, model, temperature, image_settings)
        response = self._client._prepared_request(request)
        return Iteration(**response)

    def update(self, iteration_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> Iteration:
        """
        Update an iteration.

        Args:
            iteration_id: The ID of the iteration
            json_schema: The JSON schema for the iteration
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            image_settings: Optional image settings

        Returns:
            Iteration: The updated iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(iteration_id, json_schema, model, temperature, image_settings)
        response = self._client._prepared_request(request)
        return Iteration(**response)

    def delete(self, id: str) -> DeleteResponse:
        """
        Delete an iteration.

        Args:
            id: The ID of the iteration

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(id)
        return self._client._prepared_request(request)


class Distances(SyncAPIResource, DistancesMixin):
    """Distances API wrapper for iterations"""

    def get(self, iteration_id: str, document_id: str) -> MetricResult:
        """
        Get distances for a document in an iteration.

        Args:
            iteration_id: The ID of the iteration
            document_id: The ID of the document

        Returns:
            MetricResult: The distances
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(iteration_id, document_id)
        response = self._client._prepared_request(request)
        return MetricResult(**response)


class AsyncEvals(AsyncAPIResource, EvalsMixin):
    """Async Evals API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.documents = AsyncDocuments(self._client)
        self.document = AsyncDocument(self._client)
        self.iterations = AsyncIterations(self._client)

    async def create(self, name: str, json_schema: Dict[str, Any], project_id: str) -> Evaluation:
        """
        Create a new evaluation.

        Args:
            name: The name of the evaluation
            json_schema: The JSON schema for the evaluation
            project_id: The project ID to associate with the evaluation

        Returns:
            Evaluation: The created evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(name, json_schema, project_id)
        response = await self._client._prepared_request(request)
        return Evaluation(**response)

    async def get(self, id: str) -> Evaluation:
        """
        Get an evaluation by ID.

        Args:
            id: The ID of the evaluation to retrieve

        Returns:
            Evaluation: The evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(id)
        response = await self._client._prepared_request(request)
        return Evaluation(**response)

    async def update(self, id: str, name: str) -> Evaluation:
        """
        Update an evaluation.

        Args:
            id: The ID of the evaluation to update
            name: The new name for the evaluation

        Returns:
            Evaluation: The updated evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(id, name)
        response = await self._client._prepared_request(request)
        return Evaluation(**response)

    async def list(self, project_id: str) -> List[Evaluation]:
        """
        List evaluations for a project.

        Args:
            project_id: The project ID to list evaluations for

        Returns:
            List[Evaluation]: List of evaluations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(project_id)
        response = await self._client._prepared_request(request)
        return [Evaluation(**item) for item in response.get("data", [])]

    async def delete(self, id: str) -> DeleteResponse:
        """
        Delete an evaluation.

        Args:
            id: The ID of the evaluation to delete

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(id)
        return await self._client._prepared_request(request)


class AsyncDocuments(AsyncAPIResource, DocumentsMixin):
    """Async Documents API wrapper for evaluations"""

    async def import_jsonl(self, eval_id: str, path: str) -> Evaluation:
        """
        Import documents from a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to the JSONL file

        Returns:
            Evaluation: The updated experiment with imported documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_import_jsonl(eval_id, path)
        response = await self._client._prepared_request(request)
        return Evaluation(**response)

    async def save_to_jsonl(self, eval_id: str, path: str) -> ExportResponse:
        """
        Save documents to a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to save the JSONL file

        Returns:
            ExportResponse: The response containing success status and path
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_save_to_jsonl(eval_id, path)
        return await self._client._prepared_request(request)

    async def create(self, eval_id: str, document: str, ground_truth: Dict[str, Any]) -> EvaluationDocument:
        """
        Create a document for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            document: The document file path or content
            ground_truth: The ground truth for the document

        Returns:
            EvaluationDocument: The created document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(eval_id, document, ground_truth)
        response = await self._client._prepared_request(request)
        return EvaluationDocument(**response)

    async def list(self, eval_id: str, filename: Optional[str] = None) -> List[EvaluationDocument]:
        """
        List documents for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            filename: Optional filename to filter by

        Returns:
            List[EvaluationDocument]: List of documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(eval_id, filename)
        response = await self._client._prepared_request(request)
        return [EvaluationDocument(**item) for item in response.get("data", [])]


class AsyncDocument(AsyncAPIResource, DocumentMixin):
    """Async Document API wrapper for individual document operations"""

    async def update(self, eval_id: str, id: str, ground_truth: Dict[str, Any]) -> EvaluationDocument:
        """
        Update a document.

        Args:
            eval_id: The ID of the evaluation
            id: The ID of the document
            ground_truth: The ground truth for the document

        Returns:
            EvaluationDocument: The updated document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(eval_id, id, ground_truth)
        response = await self._client._prepared_request(request)
        return EvaluationDocument(**response)

    async def delete(self, eval_id: str, id: str) -> DeleteResponse:
        """
        Delete a document.

        Args:
            eval_id: The ID of the evaluation
            id: The ID of the document

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(eval_id, id)
        return await self._client._prepared_request(request)


class AsyncIterations(AsyncAPIResource, IterationsMixin):
    """Async Iterations API wrapper for evaluations"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.distances = AsyncDistances(self._client)

    async def import_jsonl(self, eval_id: str, path: str) -> Evaluation:
        """
        Import iterations from a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to the JSONL file

        Returns:
            Evaluation: The updated experiment with imported iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_import_jsonl(eval_id, path)
        response = await self._client._prepared_request(request)
        return Evaluation(**response)

    async def save_to_jsonl(self, eval_id: str, path: str) -> ExportResponse:
        """
        Save iterations to a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to save the JSONL file

        Returns:
            ExportResponse: The response containing success status and path
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_save_to_jsonl(eval_id, path)
        return await self._client._prepared_request(request)

    async def get(self, id: str) -> Iteration:
        """
        Get an iteration by ID.

        Args:
            id: The ID of the iteration

        Returns:
            Iteration: The iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(id)
        response = await self._client._prepared_request(request)
        return Iteration(**response)

    async def list(self, eval_id: str, model: Optional[str] = None) -> List[Iteration]:
        """
        List iterations for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            model: Optional model to filter by

        Returns:
            List[Iteration]: List of iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(eval_id, model)
        response = await self._client._prepared_request(request)
        return [Iteration(**item) for item in response.get("data", [])]

    async def create(self, eval_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> Iteration:
        """
        Create a new iteration for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            json_schema: The JSON schema for the iteration
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            image_settings: Optional image settings

        Returns:
            Iteration: The created iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(eval_id, json_schema, model, temperature, image_settings)
        response = await self._client._prepared_request(request)
        return Iteration(**response)

    async def update(self, iteration_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> Iteration:
        """
        Update an iteration.

        Args:
            iteration_id: The ID of the iteration
            json_schema: The JSON schema for the iteration
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            image_settings: Optional image settings

        Returns:
            Iteration: The updated iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(iteration_id, json_schema, model, temperature, image_settings)
        response = await self._client._prepared_request(request)
        return Iteration(**response)

    async def delete(self, id: str) -> DeleteResponse:
        """
        Delete an iteration.

        Args:
            id: The ID of the iteration

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(id)
        return await self._client._prepared_request(request)


class AsyncDistances(AsyncAPIResource, DistancesMixin):
    """Async Distances API wrapper for iterations"""

    async def get(self, iteration_id: str, document_id: str) -> MetricResult:
        """
        Get distances for a document in an iteration.

        Args:
            iteration_id: The ID of the iteration
            document_id: The ID of the document

        Returns:
            MetricResult: The distances
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(iteration_id, document_id)
        response = await self._client._prepared_request(request)
        return MetricResult(**response)

