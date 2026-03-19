from typing import Any, List

from ....._resource import AsyncAPIResource, SyncAPIResource
from .....types.documents.classify import ClassifyResponse
from .....types.evals.classify import (
    CategoryOverrides,
    ClassifyIteration,
    ClassifyIterationDocument,
    CreateClassifyIterationRequest,
    PatchClassifyIterationRequest,
)
from .....types.inference_settings import InferenceSettings
from .....types.projects.metrics import OptimizedIterationMetrics
from .....types.projects.predictions import PredictionData
from .....types.standards import PreparedRequest

BASE = "/evals/classify"


class IterationsMixin:
    def prepare_create(
        self,
        eval_id: str,
        dataset_id: str,
        inference_settings: InferenceSettings | None = None,
        category_overrides: CategoryOverrides | None = None,
        parent_id: str | None = None,
    ) -> PreparedRequest:
        data = CreateClassifyIterationRequest(
            inference_settings=inference_settings or InferenceSettings(),
            category_overrides=category_overrides or CategoryOverrides(),
            project_id=eval_id,
            dataset_id=dataset_id,
            parent_id=parent_id,
        ).model_dump(mode="json", exclude_unset=True)
        return PreparedRequest(method="POST", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations", data=data)

    def prepare_get(self, eval_id: str, dataset_id: str, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}")

    def prepare_list(self, eval_id: str, dataset_id: str, before: str | None = None, after: str | None = None, limit: int = 10, order: str = "desc") -> PreparedRequest:
        params: dict[str, Any] = {"limit": limit, "order": order}
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
        category_overrides: CategoryOverrides | None = None,
    ) -> PreparedRequest:
        data = PatchClassifyIterationRequest(
            inference_settings=inference_settings,
            category_overrides=category_overrides,
        ).model_dump(mode="json", exclude_unset=True)
        return PreparedRequest(method="PATCH", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}", data=data)

    def prepare_delete(self, eval_id: str, dataset_id: str, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}")

    def prepare_finalize(self, eval_id: str, dataset_id: str, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/finalize")

    def prepare_get_categories(self, eval_id: str, dataset_id: str, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/categories")

    def prepare_get_schema(self, eval_id: str, dataset_id: str, iteration_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/schema")

    def prepare_process_documents(self, eval_id: str, dataset_id: str, iteration_id: str, dataset_document_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents/processDocumentsFromDatasetId",
            data={"dataset_document_id": dataset_document_id},
        )

    def prepare_get_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents/{document_id}")

    def prepare_list_documents(self, eval_id: str, dataset_id: str, iteration_id: str, limit: int = 1000, offset: int = 0) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents",
            params={"limit": limit, "offset": offset},
        )

    def prepare_update_document(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        document_id: str,
        prediction_data: PredictionData | None = None,
        classification_id: str | None = None,
        extraction_id: str | None = None,
    ) -> PreparedRequest:
        data: dict[str, Any] = {}
        if prediction_data is not None:
            data["prediction_data"] = prediction_data.model_dump(mode="json")
        if classification_id is not None:
            data["classification_id"] = classification_id
        if extraction_id is not None:
            data["extraction_id"] = extraction_id
        return PreparedRequest(
            method="PATCH",
            url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents/{document_id}",
            data=data,
        )

    def prepare_delete_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents/{document_id}")

    def prepare_get_metrics(self, eval_id: str, dataset_id: str, iteration_id: str, force_refresh: bool = False) -> PreparedRequest:
        params = {"force_refresh": True} if force_refresh else None
        return PreparedRequest(method="GET", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/metrics", params=params)

    def prepare_process_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/iterations/{iteration_id}/iteration-documents/{document_id}/process")


class Iterations(SyncAPIResource, IterationsMixin):
    def create(
        self,
        eval_id: str,
        dataset_id: str,
        inference_settings: InferenceSettings | None = None,
        category_overrides: CategoryOverrides | None = None,
        parent_id: str | None = None,
    ) -> ClassifyIteration:
        response = self._client._prepared_request(self.prepare_create(eval_id, dataset_id, inference_settings, category_overrides, parent_id))
        return ClassifyIteration.model_validate(response)

    def get(self, eval_id: str, dataset_id: str, iteration_id: str) -> ClassifyIteration:
        response = self._client._prepared_request(self.prepare_get(eval_id, dataset_id, iteration_id))
        return ClassifyIteration.model_validate(response)

    def list(self, eval_id: str, dataset_id: str, before: str | None = None, after: str | None = None, limit: int = 10, order: str = "desc") -> List[ClassifyIteration]:
        response = self._client._prepared_request(self.prepare_list(eval_id, dataset_id, before=before, after=after, limit=limit, order=order))
        return [ClassifyIteration.model_validate(item) for item in response.get("data", [])]

    def update_draft(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        inference_settings: InferenceSettings | None = None,
        category_overrides: CategoryOverrides | None = None,
    ) -> ClassifyIteration:
        response = self._client._prepared_request(self.prepare_update_draft(eval_id, dataset_id, iteration_id, inference_settings, category_overrides))
        return ClassifyIteration.model_validate(response)

    def delete(self, eval_id: str, dataset_id: str, iteration_id: str) -> dict:
        return self._client._prepared_request(self.prepare_delete(eval_id, dataset_id, iteration_id))

    def finalize(self, eval_id: str, dataset_id: str, iteration_id: str) -> ClassifyIteration:
        response = self._client._prepared_request(self.prepare_finalize(eval_id, dataset_id, iteration_id))
        return ClassifyIteration.model_validate(response)

    def get_categories(self, eval_id: str, dataset_id: str, iteration_id: str) -> dict:
        return self._client._prepared_request(self.prepare_get_categories(eval_id, dataset_id, iteration_id))

    def get_schema(self, eval_id: str, dataset_id: str, iteration_id: str) -> dict:
        return self._client._prepared_request(self.prepare_get_schema(eval_id, dataset_id, iteration_id))

    def process_documents(self, eval_id: str, dataset_id: str, iteration_id: str, dataset_document_id: str) -> dict:
        return self._client._prepared_request(self.prepare_process_documents(eval_id, dataset_id, iteration_id, dataset_document_id))

    def get_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> ClassifyIterationDocument:
        response = self._client._prepared_request(self.prepare_get_document(eval_id, dataset_id, iteration_id, document_id))
        return ClassifyIterationDocument.model_validate(response)

    def list_documents(self, eval_id: str, dataset_id: str, iteration_id: str, limit: int = 1000, offset: int = 0) -> List[ClassifyIterationDocument]:
        response = self._client._prepared_request(self.prepare_list_documents(eval_id, dataset_id, iteration_id, limit=limit, offset=offset))
        if isinstance(response, list):
            return [ClassifyIterationDocument.model_validate(item) for item in response]
        return [ClassifyIterationDocument.model_validate(item) for item in response.get("data", [])]

    def update_document(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        document_id: str,
        prediction_data: PredictionData | None = None,
        classification_id: str | None = None,
        extraction_id: str | None = None,
    ) -> ClassifyIterationDocument:
        response = self._client._prepared_request(
            self.prepare_update_document(
                eval_id,
                dataset_id,
                iteration_id,
                document_id,
                prediction_data=prediction_data,
                classification_id=classification_id,
                extraction_id=extraction_id,
            )
        )
        return ClassifyIterationDocument.model_validate(response)

    def delete_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> dict:
        return self._client._prepared_request(self.prepare_delete_document(eval_id, dataset_id, iteration_id, document_id))

    def get_metrics(self, eval_id: str, dataset_id: str, iteration_id: str, force_refresh: bool = False) -> OptimizedIterationMetrics:
        response = self._client._prepared_request(self.prepare_get_metrics(eval_id, dataset_id, iteration_id, force_refresh=force_refresh))
        return OptimizedIterationMetrics.model_validate(response)

    def process_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> ClassifyResponse:
        response = self._client._prepared_request(self.prepare_process_document(eval_id, dataset_id, iteration_id, document_id))
        return ClassifyResponse.model_validate(response)


class AsyncIterations(AsyncAPIResource, IterationsMixin):
    async def create(
        self,
        eval_id: str,
        dataset_id: str,
        inference_settings: InferenceSettings | None = None,
        category_overrides: CategoryOverrides | None = None,
        parent_id: str | None = None,
    ) -> ClassifyIteration:
        response = await self._client._prepared_request(self.prepare_create(eval_id, dataset_id, inference_settings, category_overrides, parent_id))
        return ClassifyIteration.model_validate(response)

    async def get(self, eval_id: str, dataset_id: str, iteration_id: str) -> ClassifyIteration:
        response = await self._client._prepared_request(self.prepare_get(eval_id, dataset_id, iteration_id))
        return ClassifyIteration.model_validate(response)

    async def list(self, eval_id: str, dataset_id: str, before: str | None = None, after: str | None = None, limit: int = 10, order: str = "desc") -> List[ClassifyIteration]:
        response = await self._client._prepared_request(self.prepare_list(eval_id, dataset_id, before=before, after=after, limit=limit, order=order))
        return [ClassifyIteration.model_validate(item) for item in response.get("data", [])]

    async def update_draft(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        inference_settings: InferenceSettings | None = None,
        category_overrides: CategoryOverrides | None = None,
    ) -> ClassifyIteration:
        response = await self._client._prepared_request(self.prepare_update_draft(eval_id, dataset_id, iteration_id, inference_settings, category_overrides))
        return ClassifyIteration.model_validate(response)

    async def delete(self, eval_id: str, dataset_id: str, iteration_id: str) -> dict:
        return await self._client._prepared_request(self.prepare_delete(eval_id, dataset_id, iteration_id))

    async def finalize(self, eval_id: str, dataset_id: str, iteration_id: str) -> ClassifyIteration:
        response = await self._client._prepared_request(self.prepare_finalize(eval_id, dataset_id, iteration_id))
        return ClassifyIteration.model_validate(response)

    async def get_categories(self, eval_id: str, dataset_id: str, iteration_id: str) -> dict:
        return await self._client._prepared_request(self.prepare_get_categories(eval_id, dataset_id, iteration_id))

    async def get_schema(self, eval_id: str, dataset_id: str, iteration_id: str) -> dict:
        return await self._client._prepared_request(self.prepare_get_schema(eval_id, dataset_id, iteration_id))

    async def process_documents(self, eval_id: str, dataset_id: str, iteration_id: str, dataset_document_id: str) -> dict:
        return await self._client._prepared_request(self.prepare_process_documents(eval_id, dataset_id, iteration_id, dataset_document_id))

    async def get_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> ClassifyIterationDocument:
        response = await self._client._prepared_request(self.prepare_get_document(eval_id, dataset_id, iteration_id, document_id))
        return ClassifyIterationDocument.model_validate(response)

    async def list_documents(self, eval_id: str, dataset_id: str, iteration_id: str, limit: int = 1000, offset: int = 0) -> List[ClassifyIterationDocument]:
        response = await self._client._prepared_request(self.prepare_list_documents(eval_id, dataset_id, iteration_id, limit=limit, offset=offset))
        if isinstance(response, list):
            return [ClassifyIterationDocument.model_validate(item) for item in response]
        return [ClassifyIterationDocument.model_validate(item) for item in response.get("data", [])]

    async def update_document(
        self,
        eval_id: str,
        dataset_id: str,
        iteration_id: str,
        document_id: str,
        prediction_data: PredictionData | None = None,
        classification_id: str | None = None,
        extraction_id: str | None = None,
    ) -> ClassifyIterationDocument:
        response = await self._client._prepared_request(
            self.prepare_update_document(
                eval_id,
                dataset_id,
                iteration_id,
                document_id,
                prediction_data=prediction_data,
                classification_id=classification_id,
                extraction_id=extraction_id,
            )
        )
        return ClassifyIterationDocument.model_validate(response)

    async def delete_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> dict:
        return await self._client._prepared_request(self.prepare_delete_document(eval_id, dataset_id, iteration_id, document_id))

    async def get_metrics(self, eval_id: str, dataset_id: str, iteration_id: str, force_refresh: bool = False) -> OptimizedIterationMetrics:
        response = await self._client._prepared_request(self.prepare_get_metrics(eval_id, dataset_id, iteration_id, force_refresh=force_refresh))
        return OptimizedIterationMetrics.model_validate(response)

    async def process_document(self, eval_id: str, dataset_id: str, iteration_id: str, document_id: str) -> ClassifyResponse:
        response = await self._client._prepared_request(self.prepare_process_document(eval_id, dataset_id, iteration_id, document_id))
        return ClassifyResponse.model_validate(response)
