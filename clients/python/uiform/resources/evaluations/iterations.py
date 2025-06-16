from typing import Any, Dict, List, Optional

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.evaluations.base import (
    CreateIterationRequest,
    DistancesResult,
    Iteration,
)
from ...types.jobs.base import InferenceSettings
from ...types.modalities import Modality
from ...types.browser_canvas import BrowserCanvas
from ...types.standards import PreparedRequest, DeleteResponse


class IterationsMixin:
    def prepare_get(self, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/evals/iterations/{iteration_id}")

    def prepare_list(self, evaluation_id: str, model: Optional[str] = None) -> PreparedRequest:
        params = {}
        if model:
            params["model"] = model
        return PreparedRequest(method="GET", url=f"/v1/evals/{evaluation_id}/iterations", params=params)

    def prepare_create(
        self,
        evaluation_id: str,
        model: str,
        json_schema: Optional[Dict[str, Any]] = None,
        temperature: float = 0.0,
        modality: Modality = "native",
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        image_resolution_dpi: int = 96,
        browser_canvas: BrowserCanvas = "A4",
        n_consensus: int = 1,
    ) -> PreparedRequest:
        inference_settings = InferenceSettings(
            model=model,
            temperature=temperature,
            modality=modality,
            reasoning_effort=reasoning_effort,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            n_consensus=n_consensus,
        )

        request = CreateIterationRequest(inference_settings=inference_settings, json_schema=json_schema)

        return PreparedRequest(method="POST", url=f"/v1/evals/{evaluation_id}/iterations/create", data=request.model_dump(exclude_none=True, mode="json"))

    def prepare_update(
        self,
        iteration_id: str,
        json_schema: Dict[str, Any],
        model: str,
        temperature: float = 0.0,
        modality: Modality = "native",
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        image_resolution_dpi: int = 96,
        browser_canvas: BrowserCanvas = "A4",
        n_consensus: int = 1,
    ) -> PreparedRequest:
        inference_settings = InferenceSettings(
            model=model,
            temperature=temperature,
            modality=modality,
            reasoning_effort=reasoning_effort,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            n_consensus=n_consensus,
        )

        iteration_data = Iteration(id=iteration_id, json_schema=json_schema, inference_settings=inference_settings, predictions=[])

        return PreparedRequest(method="PUT", url=f"/v1/evals/iterations/{iteration_id}", data=iteration_data.model_dump(exclude_none=True, mode="json"))

    def prepare_delete(self, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/evals/iterations/{iteration_id}")

    def prepare_compute_distances(self, iteration_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/evals/iterations/{iteration_id}/compute_distances/{document_id}")


class Iterations(SyncAPIResource, IterationsMixin):
    """Iterations API wrapper for evaluations"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def list(self, evaluation_id: str, model: Optional[str] = None) -> List[Iteration]:
        """
        List iterations for an evaluation.

        Args:
            evaluation_id: The ID of the evaluation
            model: Optional model to filter by

        Returns:
            List[Iteration]: List of iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(evaluation_id, model)
        response = self._client._prepared_request(request)
        return [Iteration(**item) for item in response.get("data", [])]

    def create(
        self,
        evaluation_id: str,
        model: str,
        temperature: float = 0.0,
        modality: Modality = "native",
        json_schema: Optional[Dict[str, Any]] = None,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        image_resolution_dpi: int = 96,
        browser_canvas: BrowserCanvas = "A4",
        n_consensus: int = 1,
    ) -> Iteration:
        """
        Create a new iteration for an evaluation.

        Args:
            evaluation_id: The ID of the evaluation
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
            evaluation_id=evaluation_id,
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

    def delete(self, iteration_id: str) -> DeleteResponse:
        """
        Delete an iteration.

        Args:
            iteration_id: The ID of the iteration

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(iteration_id)
        return self._client._prepared_request(request)

    def compute_distances(self, iteration_id: str, document_id: str) -> DistancesResult:
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
        request = self.prepare_compute_distances(iteration_id, document_id)
        response = self._client._prepared_request(request)
        return DistancesResult(**response)


class AsyncIterations(AsyncAPIResource, IterationsMixin):
    """Async Iterations API wrapper for evaluations"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def get(self, iteration_id: str) -> Iteration:
        """
        Get an iteration by ID.

        Args:
            iteration_id: The ID of the iteration

        Returns:
            Iteration: The iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(iteration_id)
        response = await self._client._prepared_request(request)
        return Iteration(**response)

    async def list(self, evaluation_id: str, model: Optional[str] = None) -> List[Iteration]:
        """
        List iterations for an evaluation.

        Args:
            evaluation_id: The ID of the evaluation
            model: Optional model to filter by

        Returns:
            List[Iteration]: List of iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(evaluation_id, model)
        response = await self._client._prepared_request(request)
        return [Iteration(**item) for item in response.get("data", [])]

    async def create(
        self,
        evaluation_id: str,
        model: str,
        temperature: float = 0.0,
        modality: Modality = "native",
        json_schema: Optional[Dict[str, Any]] = None,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        image_resolution_dpi: int = 96,
        browser_canvas: BrowserCanvas = "A4",
        n_consensus: int = 1,
    ) -> Iteration:
        """
        Create a new iteration for an evaluation.

        Args:
            evaluation_id: The ID of the evaluation
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
            evaluation_id=evaluation_id,
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

    async def delete(self, iteration_id: str) -> DeleteResponse:
        """
        Delete an iteration.

        Args:
            iteration_id: The ID of the iteration

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(iteration_id)
        return await self._client._prepared_request(request)

    async def compute_distances(self, iteration_id: str, document_id: str) -> DistancesResult:
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
        request = self.prepare_compute_distances(iteration_id, document_id)
        response = await self._client._prepared_request(request)
        return DistancesResult(**response)
