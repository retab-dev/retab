from typing import Any, Dict, List, Optional

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.projects.datasets import (
    CreateDatasetRequest,
    Dataset,
    DatasetDocument,
    PatchDatasetDocumentRequest,
    PatchDatasetRequest,
)
from ....types.projects.predictions import PredictionData
from ....types.mime import MIMEData
from ....types.standards import PreparedRequest
from ..iterations import Iterations, AsyncIterations


class DatasetsMixin:

    def prepare_create(
        self,
        project_id: str,
        name: str,
        base_json_schema: dict[str, Any] | None = None,
    ) -> PreparedRequest:
        data: Dict[str, Any] = {"name": name}
        if base_json_schema is not None:
            data["base_json_schema"] = base_json_schema
        return PreparedRequest(method="POST", url=f"/projects/{project_id}/datasets", data=data)

    def prepare_get(self, project_id: str, dataset_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/projects/{project_id}/datasets/{dataset_id}")

    def prepare_list(
        self,
        project_id: str,
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
        return PreparedRequest(method="GET", url=f"/projects/{project_id}/datasets", params=params)

    def prepare_update(
        self,
        project_id: str,
        dataset_id: str,
        name: str | None = None,
    ) -> PreparedRequest:
        data: Dict[str, Any] = {}
        if name is not None:
            data["name"] = name
        return PreparedRequest(method="PATCH", url=f"/projects/{project_id}/datasets/{dataset_id}", data=data)

    def prepare_delete(self, project_id: str, dataset_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/projects/{project_id}/datasets/{dataset_id}")

    def prepare_duplicate(self, project_id: str, dataset_id: str, name: str | None = None) -> PreparedRequest:
        data: Dict[str, Any] = {}
        if name is not None:
            data["name"] = name
        return PreparedRequest(method="POST", url=f"/projects/{project_id}/datasets/{dataset_id}/duplicate", data=data)

    def prepare_add_document(
        self,
        project_id: str,
        dataset_id: str,
        mime_data: MIMEData,
        prediction_data: PredictionData | None = None,
    ) -> PreparedRequest:
        data: Dict[str, Any] = {
            "mime_data": mime_data.model_dump(mode="json"),
            "project_id": project_id,
            "dataset_id": dataset_id,
        }
        if prediction_data is not None:
            data["prediction_data"] = prediction_data.model_dump(mode="json")
        return PreparedRequest(method="POST", url=f"/projects/{project_id}/datasets/{dataset_id}/dataset-documents", data=data)

    def prepare_get_document(self, project_id: str, dataset_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/projects/{project_id}/datasets/{dataset_id}/dataset-documents/{document_id}")

    def prepare_list_documents(
        self,
        project_id: str,
        dataset_id: str,
        limit: int = 1000,
        offset: int = 0,
    ) -> PreparedRequest:
        params: Dict[str, Any] = {"limit": limit, "offset": offset}
        return PreparedRequest(method="GET", url=f"/projects/{project_id}/datasets/{dataset_id}/dataset-documents", params=params)

    def prepare_update_document(
        self,
        project_id: str,
        dataset_id: str,
        document_id: str,
        validation_flags: dict[str, Any] | None = None,
        prediction_data: PredictionData | None = None,
        extraction_id: str | None = None,
    ) -> PreparedRequest:
        data: Dict[str, Any] = {}
        if validation_flags is not None:
            data["validation_flags"] = validation_flags
        if prediction_data is not None:
            data["prediction_data"] = prediction_data.model_dump(mode="json")
        if extraction_id is not None:
            data["extraction_id"] = extraction_id
        return PreparedRequest(method="PATCH", url=f"/projects/{project_id}/datasets/{dataset_id}/dataset-documents/{document_id}", data=data)

    def prepare_delete_document(self, project_id: str, dataset_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/projects/{project_id}/datasets/{dataset_id}/dataset-documents/{document_id}")


class Datasets(SyncAPIResource, DatasetsMixin):

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.iterations = Iterations(client=client)

    def create(
        self,
        project_id: str,
        name: str,
        base_json_schema: dict[str, Any] | None = None,
    ) -> Dataset:
        request = self.prepare_create(project_id, name, base_json_schema=base_json_schema)
        response = self._client._prepared_request(request)
        return Dataset.model_validate(response)

    def get(self, project_id: str, dataset_id: str) -> Dataset:
        request = self.prepare_get(project_id, dataset_id)
        response = self._client._prepared_request(request)
        return Dataset.model_validate(response)

    def list(
        self,
        project_id: str,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: str = "desc",
    ) -> List[Dataset]:
        request = self.prepare_list(project_id, before=before, after=after, limit=limit, order=order)
        response = self._client._prepared_request(request)
        return [Dataset.model_validate(item) for item in response.get("data", [])]

    def update(
        self,
        project_id: str,
        dataset_id: str,
        name: str | None = None,
    ) -> Dataset:
        request = self.prepare_update(project_id, dataset_id, name=name)
        response = self._client._prepared_request(request)
        return Dataset.model_validate(response)

    def delete(self, project_id: str, dataset_id: str) -> dict:
        request = self.prepare_delete(project_id, dataset_id)
        return self._client._prepared_request(request)

    def duplicate(self, project_id: str, dataset_id: str, name: str | None = None) -> Dataset:
        request = self.prepare_duplicate(project_id, dataset_id, name=name)
        response = self._client._prepared_request(request)
        return Dataset.model_validate(response)

    def add_document(
        self,
        project_id: str,
        dataset_id: str,
        mime_data: MIMEData,
        prediction_data: PredictionData | None = None,
    ) -> DatasetDocument:
        request = self.prepare_add_document(project_id, dataset_id, mime_data, prediction_data=prediction_data)
        response = self._client._prepared_request(request)
        return DatasetDocument.model_validate(response)

    def get_document(self, project_id: str, dataset_id: str, document_id: str) -> DatasetDocument:
        request = self.prepare_get_document(project_id, dataset_id, document_id)
        response = self._client._prepared_request(request)
        return DatasetDocument.model_validate(response)

    def list_documents(
        self,
        project_id: str,
        dataset_id: str,
        limit: int = 1000,
        offset: int = 0,
    ) -> List[DatasetDocument]:
        request = self.prepare_list_documents(project_id, dataset_id, limit=limit, offset=offset)
        response = self._client._prepared_request(request)
        if isinstance(response, list):
            return [DatasetDocument.model_validate(item) for item in response]
        return [DatasetDocument.model_validate(item) for item in response.get("data", [])]

    def update_document(
        self,
        project_id: str,
        dataset_id: str,
        document_id: str,
        validation_flags: dict[str, Any] | None = None,
        prediction_data: PredictionData | None = None,
        extraction_id: str | None = None,
    ) -> DatasetDocument:
        request = self.prepare_update_document(project_id, dataset_id, document_id, validation_flags=validation_flags, prediction_data=prediction_data, extraction_id=extraction_id)
        response = self._client._prepared_request(request)
        return DatasetDocument.model_validate(response)

    def delete_document(self, project_id: str, dataset_id: str, document_id: str) -> dict:
        request = self.prepare_delete_document(project_id, dataset_id, document_id)
        return self._client._prepared_request(request)


class AsyncDatasets(AsyncAPIResource, DatasetsMixin):

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.iterations = AsyncIterations(client=client)

    async def create(
        self,
        project_id: str,
        name: str,
        base_json_schema: dict[str, Any] | None = None,
    ) -> Dataset:
        request = self.prepare_create(project_id, name, base_json_schema=base_json_schema)
        response = await self._client._prepared_request(request)
        return Dataset.model_validate(response)

    async def get(self, project_id: str, dataset_id: str) -> Dataset:
        request = self.prepare_get(project_id, dataset_id)
        response = await self._client._prepared_request(request)
        return Dataset.model_validate(response)

    async def list(
        self,
        project_id: str,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: str = "desc",
    ) -> List[Dataset]:
        request = self.prepare_list(project_id, before=before, after=after, limit=limit, order=order)
        response = await self._client._prepared_request(request)
        return [Dataset.model_validate(item) for item in response.get("data", [])]

    async def update(
        self,
        project_id: str,
        dataset_id: str,
        name: str | None = None,
    ) -> Dataset:
        request = self.prepare_update(project_id, dataset_id, name=name)
        response = await self._client._prepared_request(request)
        return Dataset.model_validate(response)

    async def delete(self, project_id: str, dataset_id: str) -> dict:
        request = self.prepare_delete(project_id, dataset_id)
        return await self._client._prepared_request(request)

    async def duplicate(self, project_id: str, dataset_id: str, name: str | None = None) -> Dataset:
        request = self.prepare_duplicate(project_id, dataset_id, name=name)
        response = await self._client._prepared_request(request)
        return Dataset.model_validate(response)

    async def add_document(
        self,
        project_id: str,
        dataset_id: str,
        mime_data: MIMEData,
        prediction_data: PredictionData | None = None,
    ) -> DatasetDocument:
        request = self.prepare_add_document(project_id, dataset_id, mime_data, prediction_data=prediction_data)
        response = await self._client._prepared_request(request)
        return DatasetDocument.model_validate(response)

    async def get_document(self, project_id: str, dataset_id: str, document_id: str) -> DatasetDocument:
        request = self.prepare_get_document(project_id, dataset_id, document_id)
        response = await self._client._prepared_request(request)
        return DatasetDocument.model_validate(response)

    async def list_documents(
        self,
        project_id: str,
        dataset_id: str,
        limit: int = 1000,
        offset: int = 0,
    ) -> List[DatasetDocument]:
        request = self.prepare_list_documents(project_id, dataset_id, limit=limit, offset=offset)
        response = await self._client._prepared_request(request)
        if isinstance(response, list):
            return [DatasetDocument.model_validate(item) for item in response]
        return [DatasetDocument.model_validate(item) for item in response.get("data", [])]

    async def update_document(
        self,
        project_id: str,
        dataset_id: str,
        document_id: str,
        validation_flags: dict[str, Any] | None = None,
        prediction_data: PredictionData | None = None,
        extraction_id: str | None = None,
    ) -> DatasetDocument:
        request = self.prepare_update_document(project_id, dataset_id, document_id, validation_flags=validation_flags, prediction_data=prediction_data, extraction_id=extraction_id)
        response = await self._client._prepared_request(request)
        return DatasetDocument.model_validate(response)

    async def delete_document(self, project_id: str, dataset_id: str, document_id: str) -> dict:
        request = self.prepare_delete_document(project_id, dataset_id, document_id)
        return await self._client._prepared_request(request)
