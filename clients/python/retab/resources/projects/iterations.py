from typing import Any, Dict, List, Optional

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.browser_canvas import BrowserCanvas
from ...types.projects import CreateIterationRequest, Iteration, ProcessIterationRequest, IterationDocumentStatusResponse, PatchIterationRequest
from ...types.inference_settings import InferenceSettings
from ...types.projects.metrics import DistancesResult
from ...types.modalities import Modality
from ...types.standards import DeleteResponse, PreparedRequest, FieldUnset
from ...types.documents.extract import RetabParsedChatCompletion


class IterationsMixin:
    def prepare_get(self, project_id: str, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/projects/{project_id}/iterations/{iteration_id}")

    def prepare_list(self, project_id: str, model: Optional[str] = None) -> PreparedRequest:
        params = {}
        if model:
            params["model"] = model
        return PreparedRequest(method="GET", url=f"/v1/projects/{project_id}/iterations", params=params)

    def prepare_create(
        self,
        project_id: str,
        model: str = FieldUnset,
        json_schema: Optional[Dict[str, Any]] = None,
        temperature: float = FieldUnset,
        modality: Modality = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        n_consensus: int = FieldUnset,
    ) -> PreparedRequest:
        inference_dict = {}
        if model is not FieldUnset:
            inference_dict["model"] = model
        if temperature is not FieldUnset:
            inference_dict["temperature"] = temperature
        if modality is not FieldUnset:
            inference_dict["modality"] = modality
        if reasoning_effort is not FieldUnset:
            inference_dict["reasoning_effort"] = reasoning_effort
        if image_resolution_dpi is not FieldUnset:
            inference_dict["image_resolution_dpi"] = image_resolution_dpi
        if browser_canvas is not FieldUnset:
            inference_dict["browser_canvas"] = browser_canvas
        if n_consensus is not FieldUnset:
            inference_dict["n_consensus"] = n_consensus

        inference_settings = InferenceSettings(**inference_dict)

        request = CreateIterationRequest(inference_settings=inference_settings, json_schema=json_schema)

        return PreparedRequest(method="POST", url=f"/v1/projects/{project_id}/iterations", data=request.model_dump(exclude_unset=True, exclude_defaults=True, mode="json"))

    def prepare_update(
        self,
        project_id: str,
        iteration_id: str,
        json_schema: Dict[str, Any] = FieldUnset,
        model: str = FieldUnset,
        temperature: float = FieldUnset,
        modality: Modality = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        n_consensus: int = FieldUnset,
    ) -> PreparedRequest:
        inference_dict = {}
        if model is not FieldUnset:
            inference_dict["model"] = model
        if temperature is not FieldUnset:
            inference_dict["temperature"] = temperature
        if modality is not FieldUnset:
            inference_dict["modality"] = modality
        if reasoning_effort is not FieldUnset:
            inference_dict["reasoning_effort"] = reasoning_effort
        if image_resolution_dpi is not FieldUnset:
            inference_dict["image_resolution_dpi"] = image_resolution_dpi
        if browser_canvas is not FieldUnset:
            inference_dict["browser_canvas"] = browser_canvas
        if n_consensus is not FieldUnset:
            inference_dict["n_consensus"] = n_consensus

        iteration_dict = {}
        if json_schema is not FieldUnset:
            iteration_dict["json_schema"] = json_schema
        if inference_dict:  # Only add inference_settings if we have at least one field
            iteration_dict["inference_settings"] = InferenceSettings(**inference_dict)

        iteration_data = PatchIterationRequest(**iteration_dict)

        return PreparedRequest(
            method="PATCH", url=f"/v1/projects/{project_id}/iterations/{iteration_id}", data=iteration_data.model_dump(exclude_unset=True, exclude_defaults=True, mode="json")
        )

    def prepare_delete(self, project_id: str, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/projects/{project_id}/iterations/{iteration_id}")

    def prepare_compute_distances(self, project_id: str, iteration_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/projects/{project_id}/iterations/{iteration_id}/documents/{document_id}/distances")

    def prepare_process(
        self,
        project_id: str,
        iteration_id: str,
        document_ids: Optional[List[str]] = None,
        only_outdated: bool = True,
    ) -> PreparedRequest:
        request = ProcessIterationRequest(
            document_ids=document_ids,
            only_outdated=only_outdated,
        )
        return PreparedRequest(method="POST", url=f"/v1/projects/{project_id}/iterations/{iteration_id}/process", data=request.model_dump(exclude_none=True, mode="json"))

    def prepare_process_document(self, project_id: str, iteration_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"/v1/projects/{project_id}/iterations/{iteration_id}/documents/{document_id}/process", data={"stream": False})

    def prepare_status(self, project_id: str, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/projects/{project_id}/iterations/{iteration_id}/status")


class Iterations(SyncAPIResource, IterationsMixin):
    """Iterations API wrapper for projects"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def get(self, project_id: str, iteration_id: str) -> Iteration:
        request = self.prepare_get(project_id, iteration_id)
        response = self._client._prepared_request(request)
        return Iteration(**response)

    def list(self, project_id: str, model: Optional[str] = None) -> List[Iteration]:
        """
        List iterations for an project.

        Args:
            project_id: The ID of the project
            model: Optional model to filter by

        Returns:
            List[Iteration]: List of iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(project_id, model)
        response = self._client._prepared_request(request)
        return [Iteration(**item) for item in response.get("data", [])]

    def create(
        self,
        project_id: str,
        model: str = FieldUnset,
        temperature: float = FieldUnset,
        modality: Modality = FieldUnset,
        json_schema: Optional[Dict[str, Any]] = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        n_consensus: int = FieldUnset,
    ) -> Iteration:
        """
        Create a new iteration for an project.

        Args:
            project_id: The ID of the project
            json_schema: The JSON schema for the iteration (if not set, we use the one of the eval)
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            modality: The modality to use (text, image, etc.)
            reasoning_effort: The reasoning effort setting for the model (auto, low, medium, high)
            image_resolution_dpi: The DPI of the image. Defaults to 96.
            browser_canvas: The canvas size of the browser. Must be one of:
                - "A3" (11.7in x 16.54in)
                - "A4" (8.27in x 11.7in)
                - "A5" (5.83in x 8.27in)
                Defaults to "A4".
            n_consensus: Number of consensus iterations to perform

        Returns:
            Iteration: The created iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(
            project_id=project_id,
            json_schema=json_schema,
            model=model,
            temperature=temperature,
            modality=modality,
            reasoning_effort=reasoning_effort,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            n_consensus=n_consensus,
        )
        response = self._client._prepared_request(request)
        return Iteration(**response)

    def delete(self, project_id: str, iteration_id: str) -> DeleteResponse:
        """
        Delete an iteration.

        Args:
            iteration_id: The ID of the iteration

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(project_id, iteration_id)
        return self._client._prepared_request(request)

    def compute_distances(self, project_id: str, iteration_id: str, document_id: str) -> DistancesResult:
        """
        Get distances for a document in an iteration.

        Args:
            iteration_id: The ID of the iteration
            document_id: The ID of the document

        Returns:
            DistancesResult: The distances
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_compute_distances(project_id, iteration_id, document_id)
        response = self._client._prepared_request(request)
        return DistancesResult(**response)

    def process(
        self,
        project_id: str,
        iteration_id: str,
        document_ids: Optional[List[str]] = None,
        only_outdated: bool = True,
    ) -> Iteration:
        """
        Process an iteration by running extractions on documents.

        Args:
            iteration_id: The ID of the iteration
            document_ids: Optional list of specific document IDs to process
            only_outdated: Whether to only process documents that need updates

        Returns:
            Iteration: The updated iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_process(project_id, iteration_id, document_ids, only_outdated)
        response = self._client._prepared_request(request)
        return Iteration(**response)

    def process_document(self, project_id: str, iteration_id: str, document_id: str) -> RetabParsedChatCompletion:
        """
        Process a single document within an iteration.
        This method updates the iteration document with the latest extraction.

        Args:
            iteration_id: The ID of the iteration
            document_id: The ID of the document

        Returns:
            RetabParsedChatCompletion: The parsed chat completion
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_process_document(project_id, iteration_id, document_id)
        response = self._client._prepared_request(request)
        return RetabParsedChatCompletion(**response)

    def status(self, project_id: str, iteration_id: str) -> IterationDocumentStatusResponse:
        """
        Get the status of documents in an iteration.

        Args:
            iteration_id: The ID of the iteration

        Returns:
            IterationDocumentStatusResponse: The status of documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_status(project_id, iteration_id)
        response = self._client._prepared_request(request)
        return IterationDocumentStatusResponse(**response)


class AsyncIterations(AsyncAPIResource, IterationsMixin):
    """Async Iterations API wrapper for projects"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def get(self, project_id: str, iteration_id: str) -> Iteration:
        """
        Get an iteration by ID.

        Args:
            iteration_id: The ID of the iteration

        Returns:
            Iteration: The iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(project_id, iteration_id)
        response = await self._client._prepared_request(request)
        return Iteration(**response)

    async def list(self, project_id: str, model: Optional[str] = None) -> List[Iteration]:
        """
        List iterations for an project.

        Args:
            project_id: The ID of the project
            model: Optional model to filter by

        Returns:
            List[Iteration]: List of iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(project_id, model)
        response = await self._client._prepared_request(request)
        return [Iteration(**item) for item in response.get("data", [])]

    async def create(
        self,
        project_id: str,
        model: str,
        temperature: float = 0.0,
        modality: Modality = "native",
        json_schema: Optional[Dict[str, Any]] = None,
        reasoning_effort: ChatCompletionReasoningEffort = "minimal",
        image_resolution_dpi: int = 96,
        browser_canvas: BrowserCanvas = "A4",
        n_consensus: int = 1,
    ) -> Iteration:
        """
        Create a new iteration for an project.

        Args:
            project_id: The ID of the project
            json_schema: The JSON schema for the iteration
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            modality: The modality to use (text, image, etc.)
            reasoning_effort: The reasoning effort setting for the model (auto, low, medium, high)
            image_resolution_dpi: The DPI of the image. Defaults to 96.
            browser_canvas: The canvas size of the browser. Must be one of:
                - "A3" (11.7in x 16.54in)
                - "A4" (8.27in x 11.7in)
                - "A5" (5.83in x 8.27in)
                Defaults to "A4".
            n_consensus: Number of consensus iterations to perform

        Returns:
            Iteration: The created iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(
            project_id=project_id,
            json_schema=json_schema,
            model=model,
            temperature=temperature,
            modality=modality,
            reasoning_effort=reasoning_effort,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            n_consensus=n_consensus,
        )
        response = await self._client._prepared_request(request)
        return Iteration(**response)

    async def delete(self, project_id: str, iteration_id: str) -> DeleteResponse:
        """
        Delete an iteration.

        Args:
            iteration_id: The ID of the iteration

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(project_id, iteration_id)
        return await self._client._prepared_request(request)

    async def compute_distances(self, project_id: str, iteration_id: str, document_id: str) -> DistancesResult:
        """
        Get distances for a document in an iteration.

        Args:
            iteration_id: The ID of the iteration
            document_id: The ID of the document

        Returns:
            DistancesResult: The distances
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_compute_distances(project_id, iteration_id, document_id)
        response = await self._client._prepared_request(request)
        return DistancesResult(**response)

    async def process(
        self,
        project_id: str,
        iteration_id: str,
        document_ids: Optional[List[str]] = None,
        only_outdated: bool = True,
    ) -> Iteration:
        """
        Process an iteration by running extractions on documents.

        Args:
            iteration_id: The ID of the iteration
            document_ids: Optional list of specific document IDs to process
            only_outdated: Whether to only process documents that need updates

        Returns:
            Iteration: The updated iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_process(project_id, iteration_id, document_ids, only_outdated)
        response = await self._client._prepared_request(request)
        return Iteration(**response)

    async def process_document(self, project_id: str, iteration_id: str, document_id: str) -> RetabParsedChatCompletion:
        """
        Process a single document within an iteration.
        This method updates the iteration document with the latest extraction.

        Args:
            iteration_id: The ID of the iteration
            document_id: The ID of the document

        Returns:
            RetabParsedChatCompletion: The parsed chat completion
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_process_document(project_id, iteration_id, document_id)
        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion(**response)

    async def status(self, project_id: str, iteration_id: str) -> IterationDocumentStatusResponse:
        """
        Get the status of documents in an iteration.

        Args:
            iteration_id: The ID of the iteration

        Returns:
            IterationDocumentStatusResponse: The status of documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_status(project_id, iteration_id)
        response = await self._client._prepared_request(request)
        return IterationDocumentStatusResponse(**response)
