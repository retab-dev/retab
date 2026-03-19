from typing import Any, Optional

from pydantic import BaseModel

from ...types.projects import CreateProjectRequest, PatchProjectRequest, Project
from ...types.standards import DeleteResponse, PreparedRequest, UNSET, _Unset


class EvalCRUDMixin:
    BASE: str
    PROJECT_MODEL: type[BaseModel] = Project
    CREATE_MODEL: type[BaseModel] = CreateProjectRequest
    PATCH_MODEL: type[BaseModel] = PatchProjectRequest

    def prepare_create(
        self,
        name: str,
        json_schema: dict[str, Any],
        **extra_body: Any,
    ) -> PreparedRequest:
        eval_dict: dict[str, Any] = {
            "name": name,
            "json_schema": json_schema,
        }
        if extra_body:
            eval_dict.update(extra_body)

        eval_data = self.CREATE_MODEL(**eval_dict)
        return PreparedRequest(method="POST", url=self.BASE, data=eval_data.model_dump(exclude_unset=True, mode="json"))

    def prepare_get(self, eval_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{self.BASE}/{eval_id}")

    def prepare_update(
        self,
        eval_id: str,
        name: str | _Unset = UNSET,
        json_schema: dict[str, Any] | _Unset = UNSET,
        **extra_body: Any,
    ) -> PreparedRequest:
        update_dict: dict[str, Any] = {}
        if name is not UNSET:
            update_dict["name"] = name
        if json_schema is not UNSET:
            update_dict["json_schema"] = json_schema
        if extra_body:
            update_dict.update(extra_body)

        data = self.PATCH_MODEL(**update_dict).model_dump(exclude_unset=True, mode="json")
        return PreparedRequest(method="PATCH", url=f"{self.BASE}/{eval_id}", data=data)

    def prepare_list(self, **extra_params: Any) -> PreparedRequest:
        params = extra_params or None
        return PreparedRequest(method="GET", url=self.BASE, params=params)

    def prepare_delete(self, eval_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"{self.BASE}/{eval_id}")

    def prepare_publish(self, eval_id: str, origin: Optional[str] = None) -> PreparedRequest:
        params = {"origin": origin} if origin else None
        return PreparedRequest(method="POST", url=f"{self.BASE}/{eval_id}/publish", params=params)


class SyncEvalCRUDMixin(EvalCRUDMixin):
    def create(
        self,
        name: str,
        json_schema: dict[str, Any],
        **extra_body: Any,
    ) -> Project:
        request = self.prepare_create(name, json_schema, **extra_body)
        response = self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response)

    def get(self, eval_id: str) -> Project:
        request = self.prepare_get(eval_id)
        response = self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response)

    def list(self, **extra_params: Any) -> list[Project]:
        request = self.prepare_list(**extra_params)
        response = self._client._prepared_request(request)
        return [self.PROJECT_MODEL.model_validate(item) for item in response.get("data", [])]

    def update(
        self,
        eval_id: str,
        name: str | _Unset = UNSET,
        json_schema: dict[str, Any] | _Unset = UNSET,
        **extra_body: Any,
    ) -> Project:
        request = self.prepare_update(eval_id, name=name, json_schema=json_schema, **extra_body)
        response = self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response)

    def delete(self, eval_id: str) -> DeleteResponse:
        request = self.prepare_delete(eval_id)
        return self._client._prepared_request(request)

    def publish(self, eval_id: str, origin: Optional[str] = None) -> Project:
        request = self.prepare_publish(eval_id, origin=origin)
        response = self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response)


class AsyncEvalCRUDMixin(EvalCRUDMixin):
    async def create(self, name: str, json_schema: dict[str, Any], **extra_body: Any) -> Project:
        request = self.prepare_create(name, json_schema, **extra_body)
        response = await self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response)

    async def get(self, eval_id: str) -> Project:
        request = self.prepare_get(eval_id)
        response = await self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response)

    async def list(self, **extra_params: Any) -> list[Project]:
        request = self.prepare_list(**extra_params)
        response = await self._client._prepared_request(request)
        return [self.PROJECT_MODEL.model_validate(item) for item in response.get("data", [])]

    async def update(
        self,
        eval_id: str,
        name: str | _Unset = UNSET,
        json_schema: dict[str, Any] | _Unset = UNSET,
        **extra_body: Any,
    ) -> Project:
        request = self.prepare_update(eval_id, name=name, json_schema=json_schema, **extra_body)
        response = await self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response)

    async def delete(self, eval_id: str) -> DeleteResponse:
        request = self.prepare_delete(eval_id)
        return await self._client._prepared_request(request)

    async def publish(self, eval_id: str, origin: Optional[str] = None) -> Project:
        request = self.prepare_publish(eval_id, origin=origin)
        response = await self._client._prepared_request(request)
        return self.PROJECT_MODEL.model_validate(response)
