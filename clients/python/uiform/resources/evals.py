from typing import Any, Dict, List, Optional

from .._resource import AsyncAPIResource, SyncAPIResource
from ..types.standards import PreparedRequest


class EvalsMixin:
    def prepare_create(self, name: str, json_schema: Dict[str, Any], project_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url="/v1/experiments",
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
            url=f"/v1/experiments/{id}"
        )

    def prepare_update(self, id: str, name: str) -> PreparedRequest:
        return PreparedRequest(
            method="PUT",
            url=f"/v1/experiments/{id}",
            data={"name": name}
        )

    def prepare_list(self, project_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url="/v1/experiments",
            params={"project_id": project_id}
        )

    def prepare_delete(self, id: str) -> PreparedRequest:
        return PreparedRequest(
            method="DELETE",
            url=f"/v1/experiments/{id}"
        )


class DocumentsMixin:
    def prepare_import_jsonl(self, eval_id: str, path: str) -> PreparedRequest:
        # This would be implemented based on the actual API endpoint
        return PreparedRequest(
            method="POST",
            url=f"/v1/experiments/{eval_id}/import_documents",
            data={"path": path}
        )

    def prepare_save_to_jsonl(self, eval_id: str, path: str) -> PreparedRequest:
        # This would be implemented based on the actual API endpoint
        return PreparedRequest(
            method="POST",
            url=f"/v1/experiments/{eval_id}/export_documents",
            data={"path": path}
        )

    def prepare_create(self, eval_id: str, document: str, ground_truth: Dict[str, Any]) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"/v1/experiments/{eval_id}/documents",
            data={"document": document, "ground_truth": ground_truth}
        )

    def prepare_list(self, eval_id: str, filename: Optional[str] = None) -> PreparedRequest:
        params = {}
        if filename:
            params["filename"] = filename
        return PreparedRequest(
            method="GET",
            url=f"/v1/experiments/{eval_id}/documents",
            params=params
        )


class DocumentMixin:
    def prepare_update(self, eval_id: str, id: str, ground_truth: Dict[str, Any]) -> PreparedRequest:
        return PreparedRequest(
            method="PUT",
            url=f"/v1/experiments/{eval_id}/documents/{id}",
            data={"ground_truth": ground_truth}
        )

    def prepare_delete(self, eval_id: str, id: str) -> PreparedRequest:
        return PreparedRequest(
            method="DELETE",
            url=f"/v1/experiments/{eval_id}/documents/{id}"
        )


class IterationsMixin:
    def prepare_import_jsonl(self, eval_id: str, path: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"/v1/experiments/{eval_id}/add_iteration_from_jsonl",
            data={"jsonl_gcs_path": path}
        )

    def prepare_save_to_jsonl(self, eval_id: str, path: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"/v1/experiments/{eval_id}/export_iteration_as_jsonl/0",
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
            url=f"/v1/experiments/{eval_id}/iterations",
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
            url=f"/v1/experiments/{eval_id}/iterations",
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


class Evals(SyncAPIResource, EvalsMixin):
    """Evals API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.documents = Documents(self._client)
        self.document = Document(self._client)
        self.iterations = Iterations(self._client)

    def create(self, name: str, json_schema: Dict[str, Any], project_id: str) -> Dict[str, Any]:
        """
        Create a new evaluation.

        Args:
            name: The name of the evaluation
            json_schema: The JSON schema for the evaluation
            project_id: The project ID to associate with the evaluation

        Returns:
            Dict[str, Any]: The created evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(name, json_schema, project_id)
        return self._client._prepared_request(request)

    def get(self, id: str) -> Dict[str, Any]:
        """
        Get an evaluation by ID.

        Args:
            id: The ID of the evaluation to retrieve

        Returns:
            Dict[str, Any]: The evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(id)
        return self._client._prepared_request(request)

    def update(self, id: str, name: str) -> Dict[str, Any]:
        """
        Update an evaluation.

        Args:
            id: The ID of the evaluation to update
            name: The new name for the evaluation

        Returns:
            Dict[str, Any]: The updated evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(id, name)
        return self._client._prepared_request(request)

    def list(self, project_id: str) -> List[Dict[str, Any]]:
        """
        List evaluations for a project.

        Args:
            project_id: The project ID to list evaluations for

        Returns:
            List[Dict[str, Any]]: List of evaluations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(project_id)
        response = self._client._prepared_request(request)
        return response.get("data", [])

    def delete(self, id: str) -> Dict[str, Any]:
        """
        Delete an evaluation.

        Args:
            id: The ID of the evaluation to delete

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(id)
        return self._client._prepared_request(request)


class Documents(SyncAPIResource, DocumentsMixin):
    """Documents API wrapper for evaluations"""

    def import_jsonl(self, eval_id: str, path: str) -> Dict[str, Any]:
        """
        Import documents from a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to the JSONL file

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_import_jsonl(eval_id, path)
        return self._client._prepared_request(request)

    def save_to_jsonl(self, eval_id: str, path: str) -> Dict[str, Any]:
        """
        Save documents to a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to save the JSONL file

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_save_to_jsonl(eval_id, path)
        return self._client._prepared_request(request)

    def create(self, eval_id: str, document: str, ground_truth: Dict[str, Any]) -> Dict[str, Any]:
        """
        Create a document for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            document: The document file path or content
            ground_truth: The ground truth for the document

        Returns:
            Dict[str, Any]: The created document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(eval_id, document, ground_truth)
        return self._client._prepared_request(request)

    def list(self, eval_id: str, filename: Optional[str] = None) -> List[Dict[str, Any]]:
        """
        List documents for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            filename: Optional filename to filter by

        Returns:
            List[Dict[str, Any]]: List of documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(eval_id, filename)
        response = self._client._prepared_request(request)
        return response.get("data", [])


class Document(SyncAPIResource, DocumentMixin):
    """Document API wrapper for individual document operations"""

    def update(self, eval_id: str, id: str, ground_truth: Dict[str, Any]) -> Dict[str, Any]:
        """
        Update a document.

        Args:
            eval_id: The ID of the evaluation
            id: The ID of the document
            ground_truth: The ground truth for the document

        Returns:
            Dict[str, Any]: The updated document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(eval_id, id, ground_truth)
        return self._client._prepared_request(request)

    def delete(self, eval_id: str, id: str) -> Dict[str, Any]:
        """
        Delete a document.

        Args:
            eval_id: The ID of the evaluation
            id: The ID of the document

        Returns:
            Dict[str, Any]: The response
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

    def import_jsonl(self, eval_id: str, path: str) -> Dict[str, Any]:
        """
        Import iterations from a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to the JSONL file

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_import_jsonl(eval_id, path)
        return self._client._prepared_request(request)

    def save_to_jsonl(self, eval_id: str, path: str) -> Dict[str, Any]:
        """
        Save iterations to a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to save the JSONL file

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_save_to_jsonl(eval_id, path)
        return self._client._prepared_request(request)

    def get(self, id: str) -> Dict[str, Any]:
        """
        Get an iteration by ID.

        Args:
            id: The ID of the iteration

        Returns:
            Dict[str, Any]: The iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(id)
        return self._client._prepared_request(request)

    def list(self, eval_id: str, model: Optional[str] = None) -> List[Dict[str, Any]]:
        """
        List iterations for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            model: Optional model to filter by

        Returns:
            List[Dict[str, Any]]: List of iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(eval_id, model)
        response = self._client._prepared_request(request)
        return response.get("data", [])

    def create(self, eval_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Create a new iteration for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            json_schema: The JSON schema for the iteration
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            image_settings: Optional image settings

        Returns:
            Dict[str, Any]: The created iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(eval_id, json_schema, model, temperature, image_settings)
        return self._client._prepared_request(request)

    def update(self, iteration_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Update an iteration.

        Args:
            iteration_id: The ID of the iteration
            json_schema: The JSON schema for the iteration
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            image_settings: Optional image settings

        Returns:
            Dict[str, Any]: The updated iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(iteration_id, json_schema, model, temperature, image_settings)
        return self._client._prepared_request(request)

    def delete(self, id: str) -> Dict[str, Any]:
        """
        Delete an iteration.

        Args:
            id: The ID of the iteration

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(id)
        return self._client._prepared_request(request)


class Distances(SyncAPIResource, DistancesMixin):
    """Distances API wrapper for iterations"""

    def get(self, iteration_id: str, document_id: str) -> Dict[str, Any]:
        """
        Get distances for a document in an iteration.

        Args:
            iteration_id: The ID of the iteration
            document_id: The ID of the document

        Returns:
            Dict[str, Any]: The distances
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(iteration_id, document_id)
        return self._client._prepared_request(request)


class AsyncEvals(AsyncAPIResource, EvalsMixin):
    """Evals Asynchronous API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.documents = AsyncDocuments(self._client)
        self.document = AsyncDocument(self._client)
        self.iterations = AsyncIterations(self._client)

    async def create(self, name: str, json_schema: Dict[str, Any], project_id: str) -> Dict[str, Any]:
        """
        Create a new evaluation asynchronously.

        Args:
            name: The name of the evaluation
            json_schema: The JSON schema for the evaluation
            project_id: The project ID to associate with the evaluation

        Returns:
            Dict[str, Any]: The created evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(name, json_schema, project_id)
        return await self._client._prepared_request(request)

    async def get(self, id: str) -> Dict[str, Any]:
        """
        Get an evaluation by ID asynchronously.

        Args:
            id: The ID of the evaluation to retrieve

        Returns:
            Dict[str, Any]: The evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(id)
        return await self._client._prepared_request(request)

    async def update(self, id: str, name: str) -> Dict[str, Any]:
        """
        Update an evaluation asynchronously.

        Args:
            id: The ID of the evaluation to update
            name: The new name for the evaluation

        Returns:
            Dict[str, Any]: The updated evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(id, name)
        return await self._client._prepared_request(request)

    async def list(self, project_id: str) -> List[Dict[str, Any]]:
        """
        List evaluations for a project asynchronously.

        Args:
            project_id: The project ID to list evaluations for

        Returns:
            List[Dict[str, Any]]: List of evaluations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(project_id)
        response = await self._client._prepared_request(request)
        return response.get("data", [])

    async def delete(self, id: str) -> Dict[str, Any]:
        """
        Delete an evaluation asynchronously.

        Args:
            id: The ID of the evaluation to delete

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(id)
        return await self._client._prepared_request(request)


class AsyncDocuments(AsyncAPIResource, DocumentsMixin):
    """Documents Asynchronous API wrapper for evaluations"""

    async def import_jsonl(self, eval_id: str, path: str) -> Dict[str, Any]:
        """
        Import documents from a JSONL file asynchronously.

        Args:
            eval_id: The ID of the evaluation
            path: The path to the JSONL file

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_import_jsonl(eval_id, path)
        return await self._client._prepared_request(request)

    async def save_to_jsonl(self, eval_id: str, path: str) -> Dict[str, Any]:
        """
        Save documents to a JSONL file asynchronously.

        Args:
            eval_id: The ID of the evaluation
            path: The path to save the JSONL file

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_save_to_jsonl(eval_id, path)
        return await self._client._prepared_request(request)

    async def create(self, eval_id: str, document: str, ground_truth: Dict[str, Any]) -> Dict[str, Any]:
        """
        Create a document for an evaluation asynchronously.

        Args:
            eval_id: The ID of the evaluation
            document: The document file path or content
            ground_truth: The ground truth for the document

        Returns:
            Dict[str, Any]: The created document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(eval_id, document, ground_truth)
        return await self._client._prepared_request(request)

    async def list(self, eval_id: str, filename: Optional[str] = None) -> List[Dict[str, Any]]:
        """
        List documents for an evaluation asynchronously.

        Args:
            eval_id: The ID of the evaluation
            filename: Optional filename to filter by

        Returns:
            List[Dict[str, Any]]: List of documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(eval_id, filename)
        response = await self._client._prepared_request(request)
        return response.get("data", [])


class AsyncDocument(AsyncAPIResource, DocumentMixin):
    """Document Asynchronous API wrapper for individual document operations"""

    async def update(self, eval_id: str, id: str, ground_truth: Dict[str, Any]) -> Dict[str, Any]:
        """
        Update a document asynchronously.

        Args:
            eval_id: The ID of the evaluation
            id: The ID of the document
            ground_truth: The ground truth for the document

        Returns:
            Dict[str, Any]: The updated document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(eval_id, id, ground_truth)
        return await self._client._prepared_request(request)

    async def delete(self, eval_id: str, id: str) -> Dict[str, Any]:
        """
        Delete a document asynchronously.

        Args:
            eval_id: The ID of the evaluation
            id: The ID of the document

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(eval_id, id)
        return await self._client._prepared_request(request)


class AsyncIterations(AsyncAPIResource, IterationsMixin):
    """Iterations Asynchronous API wrapper for evaluations"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.distances = AsyncDistances(self._client)

    async def import_jsonl(self, eval_id: str, path: str) -> Dict[str, Any]:
        """
        Import iterations from a JSONL file asynchronously.

        Args:
            eval_id: The ID of the evaluation
            path: The path to the JSONL file

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_import_jsonl(eval_id, path)
        return await self._client._prepared_request(request)

    async def save_to_jsonl(self, eval_id: str, path: str) -> Dict[str, Any]:
        """
        Save iterations to a JSONL file asynchronously.

        Args:
            eval_id: The ID of the evaluation
            path: The path to save the JSONL file

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_save_to_jsonl(eval_id, path)
        return await self._client._prepared_request(request)

    async def get(self, id: str) -> Dict[str, Any]:
        """
        Get an iteration by ID asynchronously.

        Args:
            id: The ID of the iteration

        Returns:
            Dict[str, Any]: The iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(id)
        return await self._client._prepared_request(request)

    async def list(self, eval_id: str, model: Optional[str] = None) -> List[Dict[str, Any]]:
        """
        List iterations for an evaluation asynchronously.

        Args:
            eval_id: The ID of the evaluation
            model: Optional model to filter by

        Returns:
            List[Dict[str, Any]]: List of iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(eval_id, model)
        response = await self._client._prepared_request(request)
        return response.get("data", [])

    async def create(self, eval_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Create a new iteration for an evaluation asynchronously.

        Args:
            eval_id: The ID of the evaluation
            json_schema: The JSON schema for the iteration
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            image_settings: Optional image settings

        Returns:
            Dict[str, Any]: The created iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(eval_id, json_schema, model, temperature, image_settings)
        return await self._client._prepared_request(request)

    async def update(self, iteration_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Update an iteration asynchronously.

        Args:
            iteration_id: The ID of the iteration
            json_schema: The JSON schema for the iteration
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            image_settings: Optional image settings

        Returns:
            Dict[str, Any]: The updated iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(iteration_id, json_schema, model, temperature, image_settings)
        return await self._client._prepared_request(request)

    async def delete(self, id: str) -> Dict[str, Any]:
        """
        Delete an iteration asynchronously.

        Args:
            id: The ID of the iteration

        Returns:
            Dict[str, Any]: The response
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(id)
        return await self._client._prepared_request(request)


class AsyncDistances(AsyncAPIResource, DistancesMixin):
    """Distances Asynchronous API wrapper for iterations"""

    async def get(self, iteration_id: str, document_id: str) -> Dict[str, Any]:
        """
        Get distances for a document in an iteration asynchronously.

        Args:
            iteration_id: The ID of the iteration
            document_id: The ID of the document

        Returns:
            Dict[str, Any]: The distances
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(iteration_id, document_id)
        return await self._client._prepared_request(request) 