from typing import Any, Dict, List

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.projects import Project, PatchProjectRequest, ListProjectParams, BaseProject
from ...types.inference_settings import InferenceSettings
from ...types.standards import PreparedRequest, DeleteResponse, FieldUnset
from .documents import Documents, AsyncDocuments
from .iterations import Iterations, AsyncIterations


class ProjectsMixin:
    def prepare_create(
        self,
        name: str,
        json_schema: dict[str, Any],
        default_inference_settings: InferenceSettings = FieldUnset,
    ) -> PreparedRequest:
        # Use BaseProject model
        eval_dict = {
            "name": name,
            "json_schema": json_schema,
        }
        if default_inference_settings is not FieldUnset:
            eval_dict["default_inference_settings"] = default_inference_settings

        eval_data = BaseProject(**eval_dict)
        return PreparedRequest(method="POST", url="/v1/evaluations", data=eval_data.model_dump(exclude_unset=True, mode="json"))

    def prepare_get(self, evaluation_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/projects/{evaluation_id}")

    def prepare_update(
        self,
        evaluation_id: str,
        name: str = FieldUnset,
        json_schema: dict[str, Any] = FieldUnset,
        default_inference_settings: InferenceSettings = FieldUnset,
    ) -> PreparedRequest:
        """
        Prepare a request to update an evaluation with partial updates.

        Only the provided fields will be updated. Fields set to None will be excluded from the update.
        """
        # Build a dictionary with only the provided fields
        update_dict = {}
        if name is not FieldUnset:
            update_dict["name"] = name
        if json_schema is not FieldUnset:
            update_dict["json_schema"] = json_schema
        if default_inference_settings is not FieldUnset:
            update_dict["default_inference_settings"] = default_inference_settings

        data = PatchProjectRequest(**update_dict).model_dump(exclude_unset=True, mode="json")

        return PreparedRequest(method="PATCH", url=f"/v1/projects/{evaluation_id}", data=data)

    def prepare_list(self) -> PreparedRequest:
        """
        Prepare a request to list evaluations.

        Usage:
        >>> client.evaluations.list()                          # List all evaluations

        Returns:
            PreparedRequest: The prepared request
        """
        params_dict = {}

        params = ListProjectParams(**params_dict).model_dump(exclude_unset=True, exclude_defaults=True, mode="json")
        return PreparedRequest(method="GET", url="/v1/evaluations", params=params)

    def prepare_delete(self, id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/projects/{id}")


class Projects(SyncAPIResource, ProjectsMixin):
    """Projects API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.documents = Documents(self._client)
        self.iterations = Iterations(self._client)

    def create(
        self,
        name: str,
        json_schema: dict[str, Any],
        default_inference_settings: InferenceSettings = FieldUnset,
    ) -> Project:
        """
        Create a new evaluation.

        Args:
            name: The name of the evaluation
            json_schema: The JSON schema for the evaluation
            documents: The documents to associate with the evaluation
            default_inference_settings: The default inference settings to associate with the evaluation

        Returns:
            Project: The created evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(name, json_schema, default_inference_settings=default_inference_settings)
        response = self._client._prepared_request(request)
        return Project(**response)

    def get(self, evaluation_id: str) -> Project:
        """
        Get an evaluation by ID.

        Args:
            evaluation_id: The ID of the evaluation to retrieve

        Returns:
            Project: The evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(evaluation_id)
        response = self._client._prepared_request(request)
        return Project(**response)

    def update(
        self,
        evaluation_id: str,
        name: str = FieldUnset,
        json_schema: dict[str, Any] = FieldUnset,
        default_inference_settings: InferenceSettings = FieldUnset,
    ) -> Project:
        """
        Update an evaluation with partial updates.

        Args:
            evaluation_id: The ID of the evaluation to update
            name: Optional new name for the evaluation
            json_schema: Optional new JSON schema
            documents: Optional list of documents to update
            iterations: Optional list of iterations to update
            default_inference_settings: Optional annotation properties

        Returns:
            Project: The updated evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(
            evaluation_id=evaluation_id,
            name=name,
            json_schema=json_schema,
            default_inference_settings=default_inference_settings,
        )
        response = self._client._prepared_request(request)
        return Project(**response)

    def list(self) -> List[Project]:
        """
        List evaluations for a project.

        Args:
        Returns:
            List[Project]: List of evaluations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list()
        response = self._client._prepared_request(request)
        return [Project(**item) for item in response.get("data", [])]

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


class AsyncProjects(AsyncAPIResource, ProjectsMixin):
    """Async Projects API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.documents = AsyncDocuments(self._client)
        self.iterations = AsyncIterations(self._client)

    async def create(self, name: str, json_schema: Dict[str, Any]) -> Project:
        """
        Create a new evaluation.

        Args:
            name: The name of the evaluation
            json_schema: The JSON schema for the evaluation

        Returns:
            Project: The created evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(name, json_schema)
        response = await self._client._prepared_request(request)
        return Project(**response)

    async def get(self, evaluation_id: str) -> Project:
        """
        Get an evaluation by ID.

        Args:
            evaluation_id: The ID of the evaluation to retrieve

        Returns:
            Project: The evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(evaluation_id)
        response = await self._client._prepared_request(request)
        return Project(**response)

    async def update(
        self,
        evaluation_id: str,
        name: str = FieldUnset,
        json_schema: dict[str, Any] = FieldUnset,
        default_inference_settings: InferenceSettings = FieldUnset,
    ) -> Project:
        """
        Update an evaluation with partial updates.

        Args:
            id: The ID of the evaluation to update
            name: Optional new name for the evaluation
            json_schema: Optional new JSON schema
            documents: Optional list of documents to update
            iterations: Optional list of iterations to update
            default_inference_settings: Optional annotation properties

        Returns:
            Project: The updated evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(
            evaluation_id=evaluation_id,
            name=name,
            json_schema=json_schema,
            default_inference_settings=default_inference_settings,
        )
        response = await self._client._prepared_request(request)
        return Project(**response)

    async def list(self) -> List[Project]:
        """
        List evaluations for a project.

        Returns:
            List[Project]: List of evaluations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list()
        response = await self._client._prepared_request(request)
        return [Project(**item) for item in response.get("data", [])]

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
