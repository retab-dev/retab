from typing import Any, List

from pydantic import BaseModel

from ...types.mime import BaseMIMEData
from ...types.pagination import ListMetadata, PaginatedList
from ...types.projects.model import BuilderDocument, Project
from ...types.standards import PreparedRequest


class TemplatesMixin:
    BASE: str

    def prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: str = "desc",
        fields: str | None = None,
    ) -> PreparedRequest:
        params: dict[str, Any] = {"limit": limit, "order": order}
        if before is not None:
            params["before"] = before
        if after is not None:
            params["after"] = after
        if fields is not None:
            params["fields"] = fields
        return PreparedRequest(method="GET", url=f"{self.BASE}/templates", params=params)

    def prepare_list_builder_document_previews(self, template_ids: List[str]) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"{self.BASE}/templates/builder-documents/previews",
            params={"template_ids": ",".join(template_ids)},
        )

    def prepare_get(self, template_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{self.BASE}/templates/{template_id}")

    def prepare_list_builder_documents(self, template_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{self.BASE}/templates/{template_id}/builder-documents")

    def prepare_clone(self, template_id: str, name: str | None = None) -> PreparedRequest:
        data = {"name": name} if name is not None else {}
        return PreparedRequest(method="POST", url=f"{self.BASE}/templates/{template_id}/clone", data=data)


class SyncTemplatesMixin(TemplatesMixin):
    PROJECT_MODEL: type[BaseModel] = Project
    BUILDER_DOCUMENT_MODEL: type[BaseModel] = BuilderDocument

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: str = "desc",
        fields: str | None = None,
    ) -> PaginatedList[Any]:
        request = self.prepare_list(before=before, after=after, limit=limit, order=order, fields=fields)
        response = self._client._prepared_request(request)
        return PaginatedList(
            data=[self.PROJECT_MODEL.model_validate(item) for item in response.get("data", [])],
            list_metadata=ListMetadata.model_validate(response.get("list_metadata", {})),
        )

    def list_builder_document_previews(self, template_ids: List[str]) -> dict[str, List[BaseMIMEData]]:
        request = self.prepare_list_builder_document_previews(template_ids)
        response = self._client._prepared_request(request)
        return {
            template_id: [BaseMIMEData.model_validate(item) for item in previews]
            for template_id, previews in response.get("data", {}).items()
        }

    def get(self, template_id: str) -> BaseModel:
        request = self.prepare_get(template_id)
        response = self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response)

    def list_builder_documents(self, template_id: str) -> List[BaseModel]:
        request = self.prepare_list_builder_documents(template_id)
        response = self._client._prepared_request(request)
        return [self.BUILDER_DOCUMENT_MODEL.model_validate(item) for item in response]

    def clone(self, template_id: str, name: str | None = None) -> BaseModel:
        request = self.prepare_clone(template_id, name=name)
        response = self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response["project"])


class AsyncTemplatesMixin(TemplatesMixin):
    PROJECT_MODEL: type[BaseModel] = Project
    BUILDER_DOCUMENT_MODEL: type[BaseModel] = BuilderDocument

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: str = "desc",
        fields: str | None = None,
    ) -> PaginatedList[Any]:
        request = self.prepare_list(before=before, after=after, limit=limit, order=order, fields=fields)
        response = await self._client._prepared_request(request)
        return PaginatedList(
            data=[self.PROJECT_MODEL.model_validate(item) for item in response.get("data", [])],
            list_metadata=ListMetadata.model_validate(response.get("list_metadata", {})),
        )

    async def list_builder_document_previews(self, template_ids: List[str]) -> dict[str, List[BaseMIMEData]]:
        request = self.prepare_list_builder_document_previews(template_ids)
        response = await self._client._prepared_request(request)
        return {
            template_id: [BaseMIMEData.model_validate(item) for item in previews]
            for template_id, previews in response.get("data", {}).items()
        }

    async def get(self, template_id: str) -> BaseModel:
        request = self.prepare_get(template_id)
        response = await self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response)

    async def list_builder_documents(self, template_id: str) -> List[BaseModel]:
        request = self.prepare_list_builder_documents(template_id)
        response = await self._client._prepared_request(request)
        return [self.BUILDER_DOCUMENT_MODEL.model_validate(item) for item in response]

    async def clone(self, template_id: str, name: str | None = None) -> BaseModel:
        request = self.prepare_clone(template_id, name=name)
        response = await self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response["project"])
