import datetime
from pathlib import Path
from typing import Any, Dict, Literal, Optional

import httpx
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.ai_models import assert_valid_model_extraction
from ...types.automations.endpoints import Endpoint, ListEndpoints, UpdateEndpointRequest
from ...types.image_settings import ImageSettings
from ...types.logs import ExternalRequestLog

# from ...types.documents.extractions import DocumentExtractResponse
from ...types.mime import BaseMIMEData, MIMEData
from ...types.modalities import Modality
from ...types.standards import PreparedRequest


class EndpointsMixin:
    def prepare_create(
        self,
        name: str,
        webhook_url: HttpUrl,
        json_schema: Dict[str, Any],
        webhook_headers: Optional[Dict[str, str]] = None,
        # DocumentExtraction Config
        image_settings: Optional[Dict[str, Any]] = None,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
    ) -> PreparedRequest:
        assert_valid_model_extraction(model)

        data = {
            "name": name,
            "webhook_url": webhook_url,
            "webhook_headers": webhook_headers or {},
            "json_schema": json_schema,
            "image_settings": image_settings or ImageSettings(),
            "modality": modality,
            "model": model,
            "temperature": temperature,
            "reasoning_effort": reasoning_effort,
        }

        request = Endpoint.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/automations/endpoints", data=request.model_dump(mode='json'))

    def prepare_list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        # Filtering parameters
        endpoint_id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
    ) -> PreparedRequest:
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "endpoint_id": endpoint_id,
            "name": name,
            "webhook_url": webhook_url,
        }
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}

        return PreparedRequest(method="GET", url="/v1/automations/endpoints", params=params)

    def prepare_get(self, endpoint_id: str) -> PreparedRequest:
        """Get a specific endpoint configuration.

        Args:
            endpoint_id: ID of the endpoint

        Returns:
            Endpoint: The endpoint configuration
        """
        return PreparedRequest(method="GET", url=f"/v1/automations/endpoints/{endpoint_id}")

    def prepare_update(
        self,
        endpoint_id: str,
        name: Optional[str] = None,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        json_schema: Optional[Dict[str, Any]] = None,
        image_settings: Optional[Dict[str, Any]] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
    ) -> PreparedRequest:
        data: dict[str, Any] = {}

        if endpoint_id is not None:
            data["id"] = endpoint_id
        if name is not None:
            data["name"] = name
        if webhook_url is not None:
            data["webhook_url"] = webhook_url
        if webhook_headers is not None:
            data["webhook_headers"] = webhook_headers
        if json_schema is not None:
            data["json_schema"] = json_schema
        if image_settings is not None:
            data["image_settings"] = image_settings
        if modality is not None:
            data["modality"] = modality
        if model is not None:
            assert_valid_model_extraction(model)
            data["model"] = model
        if temperature is not None:
            data["temperature"] = temperature
        if reasoning_effort is not None:
            data["reasoning_effort"] = reasoning_effort
        request = UpdateEndpointRequest.model_validate(data)
        return PreparedRequest(method="PUT", url=f"/v1/automations/endpoints/{endpoint_id}", data=request.model_dump(mode='json'))

    def prepare_delete(self, endpoint_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/automations/endpoints/{endpoint_id}")


class Endpoints(SyncAPIResource, EndpointsMixin):
    """Endpoints API wrapper for managing endpoint configurations"""

    def create(
        self,
        name: str,
        webhook_url: HttpUrl,
        json_schema: Dict[str, Any],
        webhook_headers: Optional[Dict[str, str]] = None,
        # DocumentExtraction Config
        image_settings: Optional[Dict[str, Any]] = None,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
    ) -> Endpoint:
        """Create a new endpoint configuration.

        Args:
            name: Name of the endpoint
            webhook_url: Webhook endpoint URL
            json_schema: JSON schema for the endpoint
            webhook_headers: Optional HTTP headers for webhook requests
            image_settings: Optional image preprocessing operations
            modality: Processing modality (currently only "native" supported)
            model: AI model to use for processing
            temperature: Model temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
        Returns:
            Endpoint: The created endpoint configuration
        """
        request = self.prepare_create(name, webhook_url, json_schema, webhook_headers, image_settings, modality, model, temperature, reasoning_effort)
        response = self._client._prepared_request(request)
        print(f"Endpoint ID: {response['id']}. Send files to {self._client.base_url + f'/v1/endpoints/{response["id"]}'}")
        return Endpoint.model_validate(response)

    def list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        endpoint_id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
    ) -> ListEndpoints:
        """List endpoint configurations with pagination support.

        Args:
            before: Optional cursor for pagination before a specific endpoint ID
            after: Optional cursor for pagination after a specific endpoint ID
            limit: Optional limit on number of results (max 100)
            order: Optional sort order ("asc" or "desc")
            endpoint_id: Optional filter by endpoint ID
            name: Optional filter by endpoint name
            webhook_url: Optional filter by webhook URL

        Returns:
            ListEndpoints: Paginated list of endpoint configurations with metadata
        """
        request = self.prepare_list(before, after, limit, order, endpoint_id, name, webhook_url)
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
        name: Optional[str] = None,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        json_schema: Optional[Dict[str, Any]] = None,
        image_settings: Optional[Dict[str, Any]] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
    ) -> Endpoint:
        """Update an endpoint configuration.

        Args:
            endpoint_id: ID of the endpoint to update
            name: New name for the endpoint
            webhook_url: New webhook URL
            webhook_headers: New webhook headers
            json_schema: New JSON schema for the endpoint
            image_settings: New image preprocessing operations
            modality: New processing modality
            model: New AI model
            temperature: New temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
        Returns:
            Endpoint: The updated endpoint configuration
        """
        request = self.prepare_update(endpoint_id, name, webhook_url, webhook_headers, json_schema, image_settings, modality, model, temperature, reasoning_effort)
        response = self._client._prepared_request(request)
        return Endpoint.model_validate(response)

    def delete(self, endpoint_id: str) -> None:
        """Delete an endpoint configuration.

        Args:
            endpoint_id: ID of the endpoint to delete
        """
        request = self.prepare_delete(endpoint_id)
        self._client._prepared_request(request)
        print(f"Endpoint Deleted. ID: {endpoint_id}")


class AsyncEndpoints(AsyncAPIResource, EndpointsMixin):
    """Async Endpoints API wrapper for managing endpoint configurations"""

    async def create(
        self,
        name: str,
        webhook_url: HttpUrl,
        json_schema: Dict[str, Any],
        webhook_headers: Optional[Dict[str, str]] = None,
        # DocumentExtraction Config
        image_settings: Optional[Dict[str, Any]] = None,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
    ) -> Endpoint:
        request = self.prepare_create(name, webhook_url, json_schema, webhook_headers, image_settings, modality, model, temperature, reasoning_effort)
        response = await self._client._prepared_request(request)
        print(f"Endpoint ID: {response['id']}. Send files to {self._client.base_url + f'/v1/endpoints/{response["id"]}'}")
        return Endpoint.model_validate(response)

    async def list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        endpoint_id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
    ) -> ListEndpoints:
        request = self.prepare_list(before, after, limit, order, endpoint_id, name, webhook_url)
        response = await self._client._prepared_request(request)
        return ListEndpoints.model_validate(response)

    async def get(self, endpoint_id: str) -> Endpoint:
        request = self.prepare_get(endpoint_id)
        response = await self._client._prepared_request(request)
        return Endpoint.model_validate(response)

    async def update(
        self,
        endpoint_id: str,
        name: Optional[str] = None,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        json_schema: Optional[Dict[str, Any]] = None,
        image_settings: Optional[Dict[str, Any]] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
    ) -> Endpoint:
        request = self.prepare_update(endpoint_id, name, webhook_url, webhook_headers, json_schema, image_settings, modality, model, temperature, reasoning_effort)
        response = await self._client._prepared_request(request)
        return Endpoint.model_validate(response)

    async def delete(self, endpoint_id: str) -> None:
        request = self.prepare_delete(endpoint_id)
        await self._client._prepared_request(request)
        print(f"Endpoint Deleted. ID: {endpoint_id}")
