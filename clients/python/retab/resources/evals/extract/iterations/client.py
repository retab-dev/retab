import datetime
from typing import Any, Dict, List

from ....._resource import AsyncAPIResource, SyncAPIResource
from .....types.projects.iterations import (
    Iteration,
    IterationDocument,
    SchemaOverrides,
)
from .....types.projects.metrics import OptimizedIterationMetrics
from .....types.inference_settings import InferenceSettings
from .....types.standards import PreparedRequest
from .....types.documents.extract import RetabParsedChatCompletion

BASE = "/evals/extract"


class IterationsMixin:

    def prepare_create(
        self,
        eval_id: str,
        dataset_id: str,
        inference_settings: InferenceSettings | None = None,
        schema_overrides: SchemaOverrides | None = None,
        parent_id: str | None = None,
    ) -> PreparedRequest:
        data: Dict[str, Any] = {
            "project_id": eval_id,
            "dataset_id": dataset_id,
        }
        if inference_settings is not None:
            data["inference_settings"] = inference_settings.model_dump(mode="json")
        if schema_overrides is not None:
            data["schema_overrides"] = schema_overrides.model_dump(mode="json", exclude_none=True)
        if parent_id is not None:
            data["parent_id"] = parent_id
        return PreparedRequest(method="POST", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations", data=data)

    def prepare_get(self, eval_id: str, dataset_id: str, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}")

    def prepare_list(
        self,
        eval_id: str,
        dataset_id: str,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: str = "desc",
    ) -> PreparedRequest:
        params: Dict[str, Any] = {"limit": limit, "order": order}
        if before is not None:
            params["before"] = before
        if after is not None:
            params["after"] = after
        return PreparedRequest(method="GET", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations", params=params)

    def prepare_update_draft(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        inference_settings: InferenceSettings | None = None,
        schema_overrides: SchemaOverrides | None = None,
    ) -> PreparedRequest:
        draft: Dict[str, Any] = {
            "updated_at": datetime.datetime.now(tz=datetime.timezone.utc).isoformat(),
        }
        if inference_settings is not None:
            draft["inference_settings"] = inference_settings.model_dump(mode="json")
        if schema_overrides is not None:
            draft["schema_overrides"] = schema_overrides.model_dump(mode="json", exclude_none=True)
        return PreparedRequest(
            method="PATCH",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}",
            data={"draft": draft},
        )

    def prepare_delete(self, eval_id: str, dataset_id: str, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}")

    def prepare_finalize(self, eval_id: str, dataset_id: str, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/finalize")

    def prepare_get_schema(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        use_draft: bool = False,
    ) -> PreparedRequest:
        params: Dict[str, Any] = {}
        if use_draft:
            params["use_draft"] = True
        return PreparedRequest(
            method="GET",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/schema",
            params=params or None,
        )

    def prepare_process_documents(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        dataset_document_id: str,
    ) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents/processDocumentsFromDatasetId",
            data={"dataset_document_id": dataset_document_id},
        )

    def prepare_get_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents/{document_id}",
        )

    def prepare_list_documents(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        limit: int = 1000,
        offset: int = 0,
    ) -> PreparedRequest:
        params: Dict[str, Any] = {"limit": limit, "offset": offset}
        return PreparedRequest(
            method="GET",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents",
            params=params,
        )

    def prepare_update_document(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        document_id: str,
        prediction_data: Any | None = None,
        extraction_id: str | None = None,
    ) -> PreparedRequest:
        data: Dict[str, Any] = {}
        if prediction_data is not None:
            data["prediction_data"] = prediction_data.model_dump(mode="json") if hasattr(prediction_data, "model_dump") else prediction_data
        if extraction_id is not None:
            data["extraction_id"] = extraction_id
        return PreparedRequest(
            method="PATCH",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents/{document_id}",
            data=data,
        )

    def prepare_delete_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="DELETE",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents/{document_id}",
        )

    def prepare_get_metrics(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        force_refresh: bool = False,
    ) -> PreparedRequest:
        params: Dict[str, Any] = {}
        if force_refresh:
            params["force_refresh"] = True
        return PreparedRequest(
            method="GET",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/metrics",
            params=params or None,
        )

    def prepare_process_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents/{document_id}/process",
        )


class Iterations(SyncAPIResource, IterationsMixin):

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    def create(
        self,
        eval_id: str,
        dataset_id: str,
        inference_settings: InferenceSettings | None = None,
        schema_overrides: SchemaOverrides | None = None,
        parent_id: str | None = None,
    ) -> Iteration:
        request = self.prepare_create(eval_id=eval_id, dataset_id=dataset_id, inference_settings=inference_settings, schema_overrides=schema_overrides, parent_id=parent_id)
        response = self._client._prepared_request(request)
        return Iteration.model_validate(response)

    def get(self, eval_id: str, dataset_id: str, iteration_id: str) -> Iteration:
        request = self.prepare_get(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id)
        response = self._client._prepared_request(request)
        return Iteration.model_validate(response)

    def list(
        self,
        eval_id: str,
        dataset_id: str,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: str = "desc",
    ) -> List[Iteration]:
        request = self.prepare_list(eval_id=eval_id, dataset_id=dataset_id, before=before, after=after, limit=limit, order=order)
        response = self._client._prepared_request(request)
        return [Iteration.model_validate(item) for item in response.get("data", [])]

    def update_draft(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        inference_settings: InferenceSettings | None = None,
        schema_overrides: SchemaOverrides | None = None,
    ) -> Iteration:
        """Update the iteration's draft configuration.

        Iterations use a draft system: edits are staged in the draft sub-object.
        Call finalize() to promote the draft to the main iteration fields.
        Use get_schema(use_draft=True) to preview the effect of draft changes.
        """
        request = self.prepare_update_draft(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id, inference_settings=inference_settings, schema_overrides=schema_overrides)
        response = self._client._prepared_request(request)
        return Iteration.model_validate(response)

    def delete(self, eval_id: str, dataset_id: str, iteration_id: str) -> dict:
        request = self.prepare_delete(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id)
        return self._client._prepared_request(request)

    def finalize(self, eval_id: str, dataset_id: str, iteration_id: str) -> Iteration:
        """Finalize a draft iteration: promotes draft values to main fields, sets status to 'completed'."""
        request = self.prepare_finalize(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id)
        response = self._client._prepared_request(request)
        return Iteration.model_validate(response)

    def get_schema(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        use_draft: bool = False,
    ) -> dict:
        """Get the materialized JSON schema for this iteration.

        Args:
            use_draft: If True, applies pending draft overrides on top of finalized overrides.
        """
        request = self.prepare_get_schema(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id, use_draft=use_draft)
        return self._client._prepared_request(request)

    def process_documents(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        dataset_document_id: str,
    ) -> dict:
        """Trigger background extraction for an iteration document."""
        request = self.prepare_process_documents(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id, dataset_document_id=dataset_document_id)
        return self._client._prepared_request(request)

    def get_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> IterationDocument:
        request = self.prepare_get_document(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id, document_id=document_id)
        response = self._client._prepared_request(request)
        return IterationDocument.model_validate(response)

    def list_documents(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        limit: int = 1000,
        offset: int = 0,
    ) -> List[IterationDocument]:
        request = self.prepare_list_documents(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id, limit=limit, offset=offset)
        response = self._client._prepared_request(request)
        if isinstance(response, list):
            return [IterationDocument.model_validate(item) for item in response]
        return [IterationDocument.model_validate(item) for item in response.get("data", [])]

    def update_document(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        document_id: str,
        prediction_data: Any | None = None,
        extraction_id: str | None = None,
    ) -> IterationDocument:
        request = self.prepare_update_document(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id, document_id=document_id, prediction_data=prediction_data, extraction_id=extraction_id)
        response = self._client._prepared_request(request)
        return IterationDocument.model_validate(response)

    def delete_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> dict:
        request = self.prepare_delete_document(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id, document_id=document_id)
        return self._client._prepared_request(request)

    def get_metrics(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        force_refresh: bool = False,
    ) -> OptimizedIterationMetrics:
        request = self.prepare_get_metrics(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id, force_refresh=force_refresh)
        response = self._client._prepared_request(request)
        return OptimizedIterationMetrics.model_validate(response)

    def process_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> RetabParsedChatCompletion:
        request = self.prepare_process_document(eval_id=eval_id, dataset_id=dataset_id, iteration_id=iteration_id, document_id=document_id)
        response = self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)


class AsyncIterations(AsyncAPIResource, IterationsMixin):

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    async def create(
        self,
        eval_id: str,
        dataset_id: str,
        inference_settings: InferenceSettings | None = None,
        schema_overrides: SchemaOverrides | None = None,
        parent_id: str | None = None,
    ) -> Iteration:
        request = self.prepare_create(eval_id, dataset_id, inference_settings=inference_settings, schema_overrides=schema_overrides, parent_id=parent_id)
        response = await self._client._prepared_request(request)
        return Iteration.model_validate(response)

    async def get(self, eval_id: str, dataset_id: str, iteration_id: str) -> Iteration:
        request = self.prepare_get(eval_id, dataset_id, iteration_id)
        response = await self._client._prepared_request(request)
        return Iteration.model_validate(response)

    async def list(
        self,
        eval_id: str,
        dataset_id: str,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: str = "desc",
    ) -> List[Iteration]:
        request = self.prepare_list(eval_id, dataset_id, before=before, after=after, limit=limit, order=order)
        response = await self._client._prepared_request(request)
        return [Iteration.model_validate(item) for item in response.get("data", [])]

    async def update_draft(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        inference_settings: InferenceSettings | None = None,
        schema_overrides: SchemaOverrides | None = None,
    ) -> Iteration:
        """Update the iteration's draft configuration."""
        request = self.prepare_update_draft(eval_id, dataset_id, iteration_id, inference_settings=inference_settings, schema_overrides=schema_overrides)
        response = await self._client._prepared_request(request)
        return Iteration.model_validate(response)

    async def delete(self, eval_id: str, dataset_id: str, iteration_id: str) -> dict:
        request = self.prepare_delete(eval_id, dataset_id, iteration_id)
        return await self._client._prepared_request(request)

    async def finalize(self, eval_id: str, dataset_id: str, iteration_id: str) -> Iteration:
        """Finalize a draft iteration: promotes draft values to main fields, sets status to 'completed'."""
        request = self.prepare_finalize(eval_id, dataset_id, iteration_id)
        response = await self._client._prepared_request(request)
        return Iteration.model_validate(response)

    async def get_schema(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        use_draft: bool = False,
    ) -> dict:
        """Get the materialized JSON schema for this iteration."""
        request = self.prepare_get_schema(eval_id, dataset_id, iteration_id, use_draft=use_draft)
        return await self._client._prepared_request(request)

    async def process_documents(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        dataset_document_id: str,
    ) -> dict:
        """Trigger background extraction for an iteration document."""
        request = self.prepare_process_documents(eval_id, dataset_id, iteration_id, dataset_document_id)
        return await self._client._prepared_request(request)

    async def get_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> IterationDocument:
        request = self.prepare_get_document(eval_id, dataset_id, iteration_id, document_id)
        response = await self._client._prepared_request(request)
        return IterationDocument.model_validate(response)

    async def list_documents(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        limit: int = 1000,
        offset: int = 0,
    ) -> List[IterationDocument]:
        request = self.prepare_list_documents(eval_id, dataset_id, iteration_id, limit=limit, offset=offset)
        response = await self._client._prepared_request(request)
        if isinstance(response, list):
            return [IterationDocument.model_validate(item) for item in response]
        return [IterationDocument.model_validate(item) for item in response.get("data", [])]

    async def update_document(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        document_id: str,
        prediction_data: Any | None = None,
        extraction_id: str | None = None,
    ) -> IterationDocument:
        request = self.prepare_update_document(eval_id, dataset_id, iteration_id, document_id, prediction_data=prediction_data, extraction_id=extraction_id)
        response = await self._client._prepared_request(request)
        return IterationDocument.model_validate(response)

    async def delete_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> dict:
        request = self.prepare_delete_document(eval_id, dataset_id, iteration_id, document_id)
        return await self._client._prepared_request(request)

    async def get_metrics(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        force_refresh: bool = False,
    ) -> OptimizedIterationMetrics:
        request = self.prepare_get_metrics(eval_id, dataset_id, iteration_id, force_refresh=force_refresh)
        response = await self._client._prepared_request(request)
        return OptimizedIterationMetrics.model_validate(response)

    async def process_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> RetabParsedChatCompletion:
        request = self.prepare_process_document(eval_id, dataset_id, iteration_id, document_id)
        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)
