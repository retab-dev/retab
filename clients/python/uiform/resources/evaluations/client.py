from typing import Any, Dict, List, Optional


from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.evaluations.base import (
    DocumentItem,
    Evaluation,
    EvaluationDocument,
    Iteration,
    UpdateEvaluationRequest,
)
from ...types.jobs.base import InferenceSettings
from ...types.standards import PreparedRequest, DeleteResponse
from .documents import Documents, AsyncDocuments
from .iterations import Iterations, AsyncIterations


class EvalsMixin:
    def prepare_create(
        self,
        name: str,
        json_schema: Dict[str, Any],
        project_id: str | None = None,
        documents: List[DocumentItem] = [],
        default_inference_settings: Optional[InferenceSettings] = None,
    ) -> PreparedRequest:
        # Convert DocumentItem to EvaluationDocument
        documents_list = [
            EvaluationDocument(
                id=doc.mime_data.id,
                mime_data=doc.mime_data,
                annotation=doc.annotation,
                annotation_metadata=doc.annotation_metadata,
            )
            for doc in documents
        ]

        eval_data = Evaluation(
            name=name,
            json_schema=json_schema,
            project_id=project_id if project_id else "default_spreadsheets",
            documents=documents_list,
            iterations=[],  # We don't allow filling iterations in the SDK
            default_inference_settings=default_inference_settings,
        )
        return PreparedRequest(method="POST", url="/v1/evals", data=eval_data.model_dump(exclude_none=True, mode="json"))

    def prepare_get(self, evaluation_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/evals/{evaluation_id}")

    def prepare_update(
        self,
        evaluation_id: str,
        name: Optional[str] = None,
        project_id: Optional[str] = None,
        json_schema: Optional[Dict[str, Any]] = None,
        documents: Optional[List[EvaluationDocument]] = None,
        iterations: Optional[List[Iteration]] = None,
        default_inference_settings: Optional[InferenceSettings] = None,
    ) -> PreparedRequest:
        """
        Prepare a request to update an evaluation with partial updates.

        Only the provided fields will be updated. Fields set to None will be excluded from the update.
        """
        # Build a dictionary with only the provided fields

        update_request = UpdateEvaluationRequest(
            name=name,
            project_id=project_id,
            json_schema=json_schema,
            documents=documents,
            iterations=iterations,
            default_inference_settings=default_inference_settings,
        )

        return PreparedRequest(method="PATCH", url=f"/v1/evals/{evaluation_id}", data=update_request.model_dump(exclude_none=True, mode="json"))

    def prepare_list(self, project_id: Optional[str] = None) -> PreparedRequest:
        params = {}
        if project_id:
            params["project_id"] = project_id
        return PreparedRequest(method="GET", url="/v1/evals", params=params)

    def prepare_delete(self, id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/evals/{id}")


class Evals(SyncAPIResource, EvalsMixin):
    """Evals API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.documents = Documents(self._client)
        self.iterations = Iterations(self._client)

    def create(
        self,
        name: str,
        json_schema: Dict[str, Any],
        project_id: str | None = None,
        documents: List[DocumentItem] = [],
        default_inference_settings: Optional[InferenceSettings] = None,
    ) -> Evaluation:
        """
        Create a new evaluation.

        Args:
            name: The name of the evaluation
            json_schema: The JSON schema for the evaluation
            project_id: The project ID to associate with the evaluation
            documents: The documents to associate with the evaluation
            default_inference_settings: The default inference settings to associate with the evaluation

        Returns:
            Evaluation: The created evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(name, json_schema, project_id, documents=documents, default_inference_settings=default_inference_settings)
        response = self._client._prepared_request(request)
        return Evaluation(**response)

    def get(self, evaluation_id: str) -> Evaluation:
        """
        Get an evaluation by ID.

        Args:
            evaluation_id: The ID of the evaluation to retrieve

        Returns:
            Evaluation: The evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(evaluation_id)
        response = self._client._prepared_request(request)
        return Evaluation(**response)

    def update(
        self,
        evaluation_id: str,
        name: Optional[str] = None,
        project_id: Optional[str] = None,
        json_schema: Optional[Dict[str, Any]] = None,
        documents: Optional[List[EvaluationDocument]] = None,
        iterations: Optional[List[Iteration]] = None,
        default_inference_settings: Optional[InferenceSettings] = None,
    ) -> Evaluation:
        """
        Update an evaluation with partial updates.

        Args:
            evaluation_id: The ID of the evaluation to update
            name: Optional new name for the evaluation
            project_id: Optional new project ID
            json_schema: Optional new JSON schema
            documents: Optional list of documents to update
            iterations: Optional list of iterations to update
            default_inference_settings: Optional annotation properties

        Returns:
            Evaluation: The updated evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(
            evaluation_id=evaluation_id,
            name=name,
            project_id=project_id,
            json_schema=json_schema,
            documents=documents,
            iterations=iterations,
            default_inference_settings=default_inference_settings,
        )
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

    def delete(self, evaluation_id: str) -> DeleteResponse:
        """
        Delete an evaluation.

        Args:
            evaluation_id: The ID of the evaluation to delete

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(evaluation_id)
        return self._client._prepared_request(request)


class AsyncEvals(AsyncAPIResource, EvalsMixin):
    """Async Evals API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.documents = AsyncDocuments(self._client)
        self.iterations = AsyncIterations(self._client)

    async def create(self, name: str, json_schema: Dict[str, Any], project_id: str | None = None) -> Evaluation:
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

    async def get(self, evaluation_id: str) -> Evaluation:
        """
        Get an evaluation by ID.

        Args:
            evaluation_id: The ID of the evaluation to retrieve

        Returns:
            Evaluation: The evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(evaluation_id)
        response = await self._client._prepared_request(request)
        return Evaluation(**response)

    async def update(
        self,
        evaluation_id: str,
        name: Optional[str] = None,
        project_id: Optional[str] = None,
        json_schema: Optional[Dict[str, Any]] = None,
        documents: Optional[List[EvaluationDocument]] = None,
        iterations: Optional[List[Iteration]] = None,
        default_inference_settings: Optional[InferenceSettings] = None,
    ) -> Evaluation:
        """
        Update an evaluation with partial updates.

        Args:
            id: The ID of the evaluation to update
            name: Optional new name for the evaluation
            project_id: Optional new project ID
            json_schema: Optional new JSON schema
            documents: Optional list of documents to update
            iterations: Optional list of iterations to update
            default_inference_settings: Optional annotation properties

        Returns:
            Evaluation: The updated evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(
            evaluation_id=evaluation_id,
            name=name,
            project_id=project_id,
            json_schema=json_schema,
            documents=documents,
            iterations=iterations,
            default_inference_settings=default_inference_settings,
        )
        response = await self._client._prepared_request(request)
        return Evaluation(**response)

    async def list(self, project_id: Optional[str] = None) -> List[Evaluation]:
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

    async def delete(self, evaluation_id: str) -> DeleteResponse:
        """
        Delete an evaluation.

        Args:
            evaluation_id: The ID of the evaluation to delete

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(evaluation_id)
        return await self._client._prepared_request(request)
