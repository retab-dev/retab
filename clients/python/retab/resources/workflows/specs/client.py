from typing import Any, Dict

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.standards import PreparedRequest
from ....types.workflows import (
    DeclarativeApplyResponse,
    DeclarativeExportResponse,
    DeclarativePlanResponse,
    DeclarativeValidationResponse,
)


class WorkflowSpecsMixin:
    """Shared request builders for declarative workflow specs."""

    def prepare_validate(self, yaml_definition: str) -> PreparedRequest:
        data: Dict[str, Any] = {"yaml_definition": yaml_definition}
        return PreparedRequest(method="POST", url="/workflows/spec/validate", data=data)

    def prepare_plan(self, yaml_definition: str) -> PreparedRequest:
        data: Dict[str, Any] = {"yaml_definition": yaml_definition}
        return PreparedRequest(method="POST", url="/workflows/spec/plan", data=data)

    def prepare_apply(self, yaml_definition: str) -> PreparedRequest:
        data: Dict[str, Any] = {"yaml_definition": yaml_definition}
        return PreparedRequest(method="POST", url="/workflows/spec/apply", data=data)

    def prepare_export(self, workflow_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}/spec")


class WorkflowSpecs(SyncAPIResource, WorkflowSpecsMixin):
    """Declarative workflow spec operations."""

    def validate(self, yaml_definition: str) -> DeclarativeValidationResponse:
        request = self.prepare_validate(yaml_definition)
        response = self._client._prepared_request(request)
        return DeclarativeValidationResponse.model_validate(response)

    def plan(self, yaml_definition: str) -> DeclarativePlanResponse:
        request = self.prepare_plan(yaml_definition)
        response = self._client._prepared_request(request)
        return DeclarativePlanResponse.model_validate(response)

    def apply(self, yaml_definition: str) -> DeclarativeApplyResponse:
        request = self.prepare_apply(yaml_definition)
        response = self._client._prepared_request(request)
        return DeclarativeApplyResponse.model_validate(response)

    def export(self, workflow_id: str) -> DeclarativeExportResponse:
        request = self.prepare_export(workflow_id)
        response = self._client._prepared_request(request)
        return DeclarativeExportResponse.model_validate(response)


class AsyncWorkflowSpecs(AsyncAPIResource, WorkflowSpecsMixin):
    """Declarative workflow spec operations for async clients."""

    async def validate(self, yaml_definition: str) -> DeclarativeValidationResponse:
        request = self.prepare_validate(yaml_definition)
        response = await self._client._prepared_request(request)
        return DeclarativeValidationResponse.model_validate(response)

    async def plan(self, yaml_definition: str) -> DeclarativePlanResponse:
        request = self.prepare_plan(yaml_definition)
        response = await self._client._prepared_request(request)
        return DeclarativePlanResponse.model_validate(response)

    async def apply(self, yaml_definition: str) -> DeclarativeApplyResponse:
        request = self.prepare_apply(yaml_definition)
        response = await self._client._prepared_request(request)
        return DeclarativeApplyResponse.model_validate(response)

    async def export(self, workflow_id: str) -> DeclarativeExportResponse:
        request = self.prepare_export(workflow_id)
        response = await self._client._prepared_request(request)
        return DeclarativeExportResponse.model_validate(response)
