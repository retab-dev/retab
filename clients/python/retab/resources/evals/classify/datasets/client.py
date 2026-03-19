from typing import Any, List

from ....._resource import AsyncAPIResource, SyncAPIResource
from .....types.documents.classify import Category, ClassifyResponse
from .....types.evals.classify import (
    ClassifyDataset,
    ClassifyDatasetDocument,
    CreateClassifyDatasetRequest,
    PatchClassifyDatasetDocumentRequest,
    PatchClassifyDatasetRequest,
)
from .....types.inference_settings import InferenceSettings
from .....types.projects.predictions import PredictionData
from .....types.mime import MIMEData
from .....types.standards import PreparedRequest
from ..iterations import AsyncIterations, Iterations

BASE = "/evals/classify"


class DatasetsMixin:
    def prepare_create(
        self,
        eval_id: str,
        name: str,
        base_categories: list[Category] | list[dict[str, str]] | None = None,
        base_inference_settings: InferenceSettings | None = None,
    ) -> PreparedRequest:
        data = CreateClassifyDatasetRequest(
            name=name,
            base_categories=base_categories or [],
            base_inference_settings=base_inference_settings or InferenceSettings(),
        ).model_dump(mode="json", exclude_unset=True)
        return PreparedRequest(method="POST", url=f"{BASE}/{eval_id}/datasets", data=data)

    def prepare_get(self, eval_id: str, dataset_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{BASE}/{eval_id}/datasets/{dataset_id}")

    def prepare_list(self, eval_id: str, before: str | None = None, after: str | None = None, limit: int = 10, order: str = "desc") -> PreparedRequest:
        params: dict[str, Any] = {"limit": limit, "order": order}
        if before is not None:
            params["before"] = before
        if after is not None:
            params["after"] = after
        return PreparedRequest(method="GET", url=f"{BASE}/{eval_id}/datasets", params=params)

    def prepare_update(self, eval_id: str, dataset_id: str, name: str | None = None) -> PreparedRequest:
        data = PatchClassifyDatasetRequest(name=name).model_dump(mode="json", exclude_unset=True)
        return PreparedRequest(method="PATCH", url=f"{BASE}/{eval_id}/datasets/{dataset_id}", data=data)

    def prepare_delete(self, eval_id: str, dataset_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"{BASE}/{eval_id}/datasets/{dataset_id}")

    def prepare_duplicate(self, eval_id: str, dataset_id: str, name: str | None = None) -> PreparedRequest:
        data = {"name": name} if name is not None else None
        return PreparedRequest(method="POST", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/duplicate", data=data)

    def prepare_add_document(self, eval_id: str, dataset_id: str, mime_data: MIMEData, prediction_data: PredictionData | None = None) -> PreparedRequest:
        data: dict[str, Any] = {
            "mime_data": mime_data.model_dump(mode="json"),
            "project_id": eval_id,
            "dataset_id": dataset_id,
        }
        if prediction_data is not None:
            data["prediction_data"] = prediction_data.model_dump(mode="json")
        return PreparedRequest(method="POST", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/dataset-documents", data=data)

    def prepare_get_document(self, eval_id: str, dataset_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/dataset-documents/{document_id}")

    def prepare_list_documents(self, eval_id: str, dataset_id: str, limit: int = 1000, offset: int = 0) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/dataset-documents", params={"limit": limit, "offset": offset})

    def prepare_update_document(
        self,
        eval_id: str,
        dataset_id: str,
        document_id: str,
        validation_flag: bool | None = None,
        prediction_data: PredictionData | None = None,
        classification_id: str | None = None,
        extraction_id: str | None = None,
    ) -> PreparedRequest:
        data = PatchClassifyDatasetDocumentRequest(
            validation_flag=validation_flag,
            prediction_data=prediction_data,
            classification_id=classification_id,
            extraction_id=extraction_id,
        ).model_dump(mode="json", exclude_unset=True)
        return PreparedRequest(method="PATCH", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/dataset-documents/{document_id}", data=data)

    def prepare_delete_document(self, eval_id: str, dataset_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/dataset-documents/{document_id}")

    def prepare_process_document(self, eval_id: str, dataset_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"{BASE}/{eval_id}/datasets/{dataset_id}/dataset-documents/{document_id}/process")


class Datasets(SyncAPIResource, DatasetsMixin):
    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.iterations = Iterations(client=client)

    def create(
        self,
        eval_id: str,
        name: str,
        base_categories: list[Category] | list[dict[str, str]] | None = None,
        base_inference_settings: InferenceSettings | None = None,
    ) -> ClassifyDataset:
        response = self._client._prepared_request(self.prepare_create(eval_id, name, base_categories, base_inference_settings))
        return ClassifyDataset.model_validate(response)

    def get(self, eval_id: str, dataset_id: str) -> ClassifyDataset:
        response = self._client._prepared_request(self.prepare_get(eval_id, dataset_id))
        return ClassifyDataset.model_validate(response)

    def list(self, eval_id: str, before: str | None = None, after: str | None = None, limit: int = 10, order: str = "desc") -> List[ClassifyDataset]:
        response = self._client._prepared_request(self.prepare_list(eval_id, before=before, after=after, limit=limit, order=order))
        return [ClassifyDataset.model_validate(item) for item in response.get("data", [])]

    def update(self, eval_id: str, dataset_id: str, name: str | None = None) -> ClassifyDataset:
        response = self._client._prepared_request(self.prepare_update(eval_id, dataset_id, name=name))
        return ClassifyDataset.model_validate(response)

    def delete(self, eval_id: str, dataset_id: str) -> dict:
        return self._client._prepared_request(self.prepare_delete(eval_id, dataset_id))

    def duplicate(self, eval_id: str, dataset_id: str, name: str | None = None) -> ClassifyDataset:
        response = self._client._prepared_request(self.prepare_duplicate(eval_id, dataset_id, name=name))
        return ClassifyDataset.model_validate(response)

    def add_document(self, eval_id: str, dataset_id: str, mime_data: MIMEData, prediction_data: PredictionData | None = None) -> ClassifyDatasetDocument:
        response = self._client._prepared_request(self.prepare_add_document(eval_id, dataset_id, mime_data, prediction_data=prediction_data))
        return ClassifyDatasetDocument.model_validate(response)

    def get_document(self, eval_id: str, dataset_id: str, document_id: str) -> ClassifyDatasetDocument:
        response = self._client._prepared_request(self.prepare_get_document(eval_id, dataset_id, document_id))
        return ClassifyDatasetDocument.model_validate(response)

    def list_documents(self, eval_id: str, dataset_id: str, limit: int = 1000, offset: int = 0) -> List[ClassifyDatasetDocument]:
        response = self._client._prepared_request(self.prepare_list_documents(eval_id, dataset_id, limit=limit, offset=offset))
        if isinstance(response, list):
            return [ClassifyDatasetDocument.model_validate(item) for item in response]
        return [ClassifyDatasetDocument.model_validate(item) for item in response.get("data", [])]

    def update_document(
        self,
        eval_id: str,
        dataset_id: str,
        document_id: str,
        validation_flag: bool | None = None,
        prediction_data: PredictionData | None = None,
        classification_id: str | None = None,
        extraction_id: str | None = None,
    ) -> ClassifyDatasetDocument:
        response = self._client._prepared_request(
            self.prepare_update_document(
                eval_id,
                dataset_id,
                document_id,
                validation_flag=validation_flag,
                prediction_data=prediction_data,
                classification_id=classification_id,
                extraction_id=extraction_id,
            )
        )
        return ClassifyDatasetDocument.model_validate(response)

    def delete_document(self, eval_id: str, dataset_id: str, document_id: str) -> dict:
        return self._client._prepared_request(self.prepare_delete_document(eval_id, dataset_id, document_id))

    def process_document(self, eval_id: str, dataset_id: str, document_id: str) -> ClassifyResponse:
        response = self._client._prepared_request(self.prepare_process_document(eval_id, dataset_id, document_id))
        return ClassifyResponse.model_validate(response)


class AsyncDatasets(AsyncAPIResource, DatasetsMixin):
    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.iterations = AsyncIterations(client=client)

    async def create(
        self,
        eval_id: str,
        name: str,
        base_categories: list[Category] | list[dict[str, str]] | None = None,
        base_inference_settings: InferenceSettings | None = None,
    ) -> ClassifyDataset:
        response = await self._client._prepared_request(self.prepare_create(eval_id, name, base_categories, base_inference_settings))
        return ClassifyDataset.model_validate(response)

    async def get(self, eval_id: str, dataset_id: str) -> ClassifyDataset:
        response = await self._client._prepared_request(self.prepare_get(eval_id, dataset_id))
        return ClassifyDataset.model_validate(response)

    async def list(self, eval_id: str, before: str | None = None, after: str | None = None, limit: int = 10, order: str = "desc") -> List[ClassifyDataset]:
        response = await self._client._prepared_request(self.prepare_list(eval_id, before=before, after=after, limit=limit, order=order))
        return [ClassifyDataset.model_validate(item) for item in response.get("data", [])]

    async def update(self, eval_id: str, dataset_id: str, name: str | None = None) -> ClassifyDataset:
        response = await self._client._prepared_request(self.prepare_update(eval_id, dataset_id, name=name))
        return ClassifyDataset.model_validate(response)

    async def delete(self, eval_id: str, dataset_id: str) -> dict:
        return await self._client._prepared_request(self.prepare_delete(eval_id, dataset_id))

    async def duplicate(self, eval_id: str, dataset_id: str, name: str | None = None) -> ClassifyDataset:
        response = await self._client._prepared_request(self.prepare_duplicate(eval_id, dataset_id, name=name))
        return ClassifyDataset.model_validate(response)

    async def add_document(self, eval_id: str, dataset_id: str, mime_data: MIMEData, prediction_data: PredictionData | None = None) -> ClassifyDatasetDocument:
        response = await self._client._prepared_request(self.prepare_add_document(eval_id, dataset_id, mime_data, prediction_data=prediction_data))
        return ClassifyDatasetDocument.model_validate(response)

    async def get_document(self, eval_id: str, dataset_id: str, document_id: str) -> ClassifyDatasetDocument:
        response = await self._client._prepared_request(self.prepare_get_document(eval_id, dataset_id, document_id))
        return ClassifyDatasetDocument.model_validate(response)

    async def list_documents(self, eval_id: str, dataset_id: str, limit: int = 1000, offset: int = 0) -> List[ClassifyDatasetDocument]:
        response = await self._client._prepared_request(self.prepare_list_documents(eval_id, dataset_id, limit=limit, offset=offset))
        if isinstance(response, list):
            return [ClassifyDatasetDocument.model_validate(item) for item in response]
        return [ClassifyDatasetDocument.model_validate(item) for item in response.get("data", [])]

    async def update_document(
        self,
        eval_id: str,
        dataset_id: str,
        document_id: str,
        validation_flag: bool | None = None,
        prediction_data: PredictionData | None = None,
        classification_id: str | None = None,
        extraction_id: str | None = None,
    ) -> ClassifyDatasetDocument:
        response = await self._client._prepared_request(
            self.prepare_update_document(
                eval_id,
                dataset_id,
                document_id,
                validation_flag=validation_flag,
                prediction_data=prediction_data,
                classification_id=classification_id,
                extraction_id=extraction_id,
            )
        )
        return ClassifyDatasetDocument.model_validate(response)

    async def delete_document(self, eval_id: str, dataset_id: str, document_id: str) -> dict:
        return await self._client._prepared_request(self.prepare_delete_document(eval_id, dataset_id, document_id))

    async def process_document(self, eval_id: str, dataset_id: str, document_id: str) -> ClassifyResponse:
        response = await self._client._prepared_request(self.prepare_process_document(eval_id, dataset_id, document_id))
        return ClassifyResponse.model_validate(response)
