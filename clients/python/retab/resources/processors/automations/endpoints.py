from typing import Literal, Optional

from pydantic_core import PydanticUndefined

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....utils.ai_models import assert_valid_model_extraction
from ....types.automations.endpoints import Endpoint, ListEndpoints, UpdateEndpointRequest
from ....types.standards import PreparedRequest


class EndpointsMixin:
    def prepare_create(
        self,
        processor_id: str,
        name: str,
        webhook_url: str,
        model: str = "gpt-4o-mini",
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
    ) -> PreparedRequest:
        assert_valid_model_extraction(model)

        request = Endpoint(
            processor_id=processor_id,
            name=name,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
        )
        return PreparedRequest(method="POST", url="/v1/processors/automations/endpoints", data=request.model_dump(mode="json"))

    def prepare_list(
        self,
        processor_id: str,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        # Filtering parameters
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
    ) -> PreparedRequest:
        params = {
            "processor_id": processor_id,
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "name": name,
            "webhook_url": webhook_url,
        }
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}

        return PreparedRequest(method="GET", url="/v1/processors/automations/endpoints", params=params)

    def prepare_get(self, endpoint_id: str) -> PreparedRequest:
        """Get a specific endpoint configuration.

        Args:
            endpoint_id: ID of the endpoint

        Returns:
            Endpoint: The endpoint configuration
        """
        return PreparedRequest(method="GET", url=f"/v1/processors/automations/endpoints/{endpoint_id}")

    def prepare_update(
        self,
        endpoint_id: str,
        name: str = PydanticUndefined,  # type: ignore[assignment]
        default_language: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_url: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
    ) -> PreparedRequest:
        request = UpdateEndpointRequest(
            name=name,
            default_language=default_language,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
        )
        return PreparedRequest(method="PUT", url=f"/v1/processors/automations/endpoints/{endpoint_id}", data=request.model_dump(mode="json"))

    def prepare_delete(self, endpoint_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/processors/automations/endpoints/{endpoint_id}")


class Endpoints(SyncAPIResource, EndpointsMixin):
    """Endpoints API wrapper for managing endpoint configurations"""

    def create(
        self,
        processor_id: str,
        name: str,
        webhook_url: str,
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
    ) -> Endpoint:
        """Create a new endpoint configuration.

        Args:
            name: Name of the endpoint
            webhook_url: Webhook endpoint URL
            json_schema: JSON schema for the endpoint
            webhook_headers: Optional HTTP headers for webhook requests
            image_resolution_dpi: Optional image resolution DPI
            browser_canvas: Optional browser canvas size
            modality: Processing modality (currently only "native" supported)
            model: AI model to use for processing
            temperature: Model temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
        Returns:
            Endpoint: The created endpoint configuration
        """
        request = self.prepare_create(
            processor_id=processor_id,
            name=name,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
        )
        response = self._client._prepared_request(request)
        print(f"Endpoint Created. Url: https://www.retab.com/dashboard/processors/automations/{response['id']}")
        return Endpoint.model_validate(response)

    def list(
        self,
        processor_id: str,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
    ) -> ListEndpoints:
        """List endpoint configurations with pagination support.

        Args:
            before: Optional cursor for pagination before a specific endpoint ID
            after: Optional cursor for pagination after a specific endpoint ID
            limit: Optional limit on number of results (max 100)
            order: Optional sort order ("asc" or "desc")
            name: Optional filter by endpoint name
            webhook_url: Optional filter by webhook URL

        Returns:
            ListEndpoints: Paginated list of endpoint configurations with metadata
        """
        request = self.prepare_list(processor_id, before, after, limit, order, name, webhook_url)
        response = self._client._prepared_request(request)
        return ListEndpoints.model_validate(response)

    def get(self, endpoint_id: str) -> Endpoint:
        """Get a specific endpoint configuration.

        Args:
            endpoint_id: ID of the endpoint

        Returns:
            Endpoint: The endpoint configuration
        """
        request = self.prepare_get(endpoint_id)
        response = self._client._prepared_request(request)
        return Endpoint.model_validate(response)

    def update(
        self,
        endpoint_id: str,
        name: str = PydanticUndefined,  # type: ignore[assignment]
        default_language: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_url: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
    ) -> Endpoint:
        """Update an endpoint configuration.

        Args:
            id: ID of the endpoint to update
            name: New name for the endpoint
            webhook_url: New webhook URL
            webhook_headers: New webhook headers
            json_schema: New JSON schema for the endpoint
            image_resolution_dpi: New image resolution DPI
            browser_canvas: New browser canvas size
            modality: New processing modality
            model: New AI model
            temperature: New temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
        Returns:
            Endpoint: The updated endpoint configuration
        """
        request = self.prepare_update(
            endpoint_id=endpoint_id,
            name=name,
            default_language=default_language,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
        )
        response = self._client._prepared_request(request)
        return Endpoint.model_validate(response)

    def delete(self, endpoint_id: str) -> None:
        """Delete an endpoint configuration.

        Args:
            id: ID of the endpoint to delete
        """
        request = self.prepare_delete(endpoint_id)
        self._client._prepared_request(request)
        print(f"Endpoint Deleted. ID: {endpoint_id}")


class AsyncEndpoints(AsyncAPIResource, EndpointsMixin):
    """Async Endpoints API wrapper for managing endpoint configurations"""

    async def create(
        self,
        processor_id: str,
        name: str,
        webhook_url: str,
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
    ) -> Endpoint:
        request = self.prepare_create(
            processor_id=processor_id,
            name=name,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
        )
        response = await self._client._prepared_request(request)
        print(f"Endpoint Created. Url: https://www.retab.com/dashboard/processors/automations/{response['id']}")

        return Endpoint.model_validate(response)

    async def list(
        self,
        processor_id: str,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
    ) -> ListEndpoints:
        request = self.prepare_list(processor_id, before, after, limit, order, name, webhook_url)
        response = await self._client._prepared_request(request)
        return ListEndpoints.model_validate(response)

    async def get(self, endpoint_id: str) -> Endpoint:
        request = self.prepare_get(endpoint_id)
        response = await self._client._prepared_request(request)
        return Endpoint.model_validate(response)

    async def update(
        self,
        endpoint_id: str,
        name: str = PydanticUndefined,  # type: ignore[assignment]
        default_language: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_url: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
    ) -> Endpoint:
        request = self.prepare_update(
            endpoint_id=endpoint_id,
            name=name,
            default_language=default_language,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
        )
        response = await self._client._prepared_request(request)
        return Endpoint.model_validate(response)

    async def delete(self, endpoint_id: str) -> None:
        request = self.prepare_delete(endpoint_id)
        await self._client._prepared_request(request)
