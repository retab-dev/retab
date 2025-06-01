import datetime
import json
from io import IOBase
from pathlib import Path
from typing import Any, Dict, Literal, Optional

import httpx
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from PIL.Image import Image
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.ai_models import assert_valid_model_extraction
from ..._utils.mime import prepare_mime_document
from ...types.deployments.links import Link, ListLinks, UpdateLinkRequest
from ...types.documents.extractions import UiParsedChatCompletion
from ...types.logs import DeploymentLog, ExternalRequestLog, ListLogs
from ...types.mime import BaseMIMEData, MIMEData
from ...types.modalities import Modality
from ...types.standards import PreparedRequest


class LinksMixin:
    def prepare_create(
        self,
        name: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        webhook_headers: Optional[Dict[str, str]] = None,
        password: str | None = None,
        # DocumentExtraction Config
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[str] = None,
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
            "password": password,
            "image_resolution_dpi": image_resolution_dpi,
            "browser_canvas": browser_canvas,
            "modality": modality,
            "model": model,
            "temperature": temperature,
            "reasoning_effort": reasoning_effort,
        }

        request = Link.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/deployments/links", data=request.model_dump(mode='json'))

    def prepare_list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        # Filtering parameters
        link_id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> PreparedRequest:
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "automation_id": link_id,
            "name": name,
            "webhook_url": webhook_url,
            "schema_id": schema_id,
            "schema_data_id": schema_data_id,
        }
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}

        return PreparedRequest(method="GET", url="/v1/deployments", params=params)

    def prepare_get(self, link_id: str) -> PreparedRequest:
        """Get a specific extraction link configuration.

        Args:
            link_id: ID of the extraction link

        Returns:
            Link: The extraction link configuration
        """
        return PreparedRequest(method="GET", url=f"/v1/deployments/{link_id}")

    def prepare_update(
        self,
        link_id: str,
        name: Optional[str] = None,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        password: Optional[str] = None,
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[str] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
        json_schema: Optional[Dict[str, Any]] = None,
    ) -> PreparedRequest:
        data: dict[str, Any] = {}

        if link_id is not None:
            data["id"] = link_id
        if name is not None:
            data["name"] = name
        if webhook_url is not None:
            data["webhook_url"] = webhook_url
        if webhook_headers is not None:
            data["webhook_headers"] = webhook_headers
        if password is not None:
            data["password"] = password
        if image_resolution_dpi is not None:
            data["image_resolution_dpi"] = image_resolution_dpi
        if browser_canvas is not None:
            data["browser_canvas"] = browser_canvas
        if modality is not None:
            data["modality"] = modality
        if model is not None:
            assert_valid_model_extraction(model)
            data["model"] = model
        if temperature is not None:
            data["temperature"] = temperature
        if json_schema is not None:
            data["json_schema"] = json_schema
        if reasoning_effort is not None:
            data["reasoning_effort"] = reasoning_effort

        request = UpdateLinkRequest.model_validate(data)
        return PreparedRequest(method="PUT", url=f"/v1/deployments/{link_id}", data=request.model_dump(mode='json'))

    def prepare_delete(self, link_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/deployments/links/{link_id}", raise_for_status=True)

    def prepare_logs(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        # Filtering parameters
        link_id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> PreparedRequest:
        """Get logs for extraction links with pagination support.

        Args:
            before: Optional cursor for pagination - get results before this log ID
            after: Optional cursor for pagination - get results after this log ID
            limit: Maximum number of logs to return (1-100, default 10)
            order: Sort order by creation time - "asc" or "desc" (default "desc")
            link_id: Optional ID of a specific extraction link to filter logs for
            name: Optional filter by link name
            webhook_url: Optional filter by webhook URL
            schema_id: Optional filter by schema ID
            schema_data_id: Optional filter by schema data ID

        Returns:
            ListLinkLogsResponse: Paginated list of logs and metadata
        """
        params = {
            "automation_id": link_id,
            "name": name,
            "webhook_url": webhook_url,
            "schema_id": schema_id,
            "schema_data_id": schema_data_id,
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
        }
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}

        return PreparedRequest(method="GET", url="/v1/deployments/logs", params=params)


class Links(SyncAPIResource, LinksMixin):
    """Extraction Link API wrapper for managing extraction link configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    def create(
        self,
        name: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        webhook_headers: Optional[Dict[str, str]] = None,
        password: str | None = None,
        # DocumentExtraction Config
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[str] = None,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
    ) -> Link:
        """Create a new extraction link configuration.

        Args:
            name: Name of the extraction link
            json_schema: JSON schema to validate extracted data
            webhook_url: Webhook endpoint for forwarding processed files
            webhook_headers: Optional HTTP headers for webhook requests
            password: Optional password for protected links
            image_resolution_dpi: Optional image resolution DPI
            browser_canvas: Optional browser canvas
            modality: Processing modality (currently only "native" supported)
            model: AI model to use for processing
            temperature: Model temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
        Returns:
            Link: The created extraction link configuration
        """

        request = self.prepare_create(name, json_schema, webhook_url, webhook_headers, password, image_resolution_dpi, browser_canvas, modality, model, temperature, reasoning_effort)
        response = self._client._prepared_request(request)

        print(f"Extraction Link Created. Link available at https://www.uiform.com/dashboard/deployments/{response['id']}")
        return Link.model_validate(response)

    def list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        # Filtering parameters
        link_id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> ListLinks:
        """List extraction link configurations with pagination support.

        Args:
            before: Optional cursor for pagination before a specific link ID
            after: Optional cursor for pagination after a specific link ID
            limit: Optional limit on number of results (max 100)
            order: Optional sort order ("asc" or "desc")
            link_id: Optional filter by extraction link ID
            name: Optional filter by link name
            webhook_url: Optional filter by webhook URL
            schema_id: Optional filter by schema ID
            schema_data_id: Optional filter by schema data ID

        Returns:
            ListLinks: Paginated list of extraction link configurations with metadata
        """
        request = self.prepare_list(before, after, limit, order, link_id, name, webhook_url, schema_id, schema_data_id)
        response = self._client._prepared_request(request)
        return ListLinks.model_validate(response)

    def get(self, link_id: str) -> Link:
        """Get a specific extraction link configuration.

        Args:
            link_id: ID of the extraction link

        Returns:
            Link: The extraction link configuration
        """
        request = self.prepare_get(link_id)
        response = self._client._prepared_request(request)
        return Link.model_validate(response)

    def update(
        self,
        link_id: str,
        name: Optional[str] = None,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        password: Optional[str] = None,
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[str] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
        json_schema: Optional[Dict[str, Any]] = None,
    ) -> Link:
        """Update an extraction link configuration.

        Args:
            link_id: ID of the extraction link to update
            name: New name for the link
            webhook_url: New webhook endpoint URL
            webhook_headers: New webhook headers
            password: New password for protected links
            image_resolution_dpi: New image resolution DPI
            browser_canvas: New browser canvas
            modality: New processing modality
            model: New AI model
            temperature: New temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
            json_schema: New JSON schema

        Returns:
            Link: The updated extraction link configuration
        """

        request = self.prepare_update(link_id, name, webhook_url, webhook_headers, password, image_resolution_dpi, browser_canvas, modality, model, temperature, reasoning_effort, json_schema)
        response = self._client._prepared_request(request)
        return Link.model_validate(response)

    def delete(self, link_id: str) -> None:
        """Delete an extraction link configuration.

        Args:
            link_id: ID of the extraction link to delete

        Returns:
            Dict[str, str]: Response message confirming deletion
        """
        request = self.prepare_delete(link_id)
        self._client._prepared_request(request)

    def logs(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        # Filtering parameters
        link_id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> ListLogs:
        """Get logs for extraction links with pagination support.

        Args:
            before: Optional cursor for pagination - get results before this log ID
            after: Optional cursor for pagination - get results after this log ID
            limit: Maximum number of logs to return (1-100, default 10)
            order: Sort order by creation time - "asc" or "desc" (default "desc")
            link_id: Optional ID of a specific extraction link to filter logs for
            name: Optional filter by link name
            webhook_url: Optional filter by webhook URL
            schema_id: Optional filter by schema ID
            schema_data_id: Optional filter by schema data ID

        Returns:
            ListLinkLogsResponse: Paginated list of logs and metadata
        """
        request = self.prepare_logs(before, after, limit, order, link_id, name, webhook_url, schema_id, schema_data_id)
        response = self._client._prepared_request(request)
        return ListLogs.model_validate(response)


class AsyncLinks(AsyncAPIResource, LinksMixin):
    """Async Extraction Link API wrapper for managing extraction link configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    async def create(
        self,
        name: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        webhook_headers: Optional[Dict[str, str]] = None,
        password: str | None = None,
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[str] = None,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
    ) -> Link:
        request = self.prepare_create(name, json_schema, webhook_url, webhook_headers, password, image_resolution_dpi, browser_canvas, modality, model, temperature, reasoning_effort)
        response = await self._client._prepared_request(request)
        print(f"Extraction Link Created. Link available at https://www.uiform.com/dashboard/deployments/{response['id']}")
        return Link.model_validate(response)

    async def list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        link_id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> ListLinks:
        request = self.prepare_list(before, after, limit, order, link_id, name, webhook_url, schema_id, schema_data_id)
        response = await self._client._prepared_request(request)
        return ListLinks.model_validate(response)

    async def get(self, link_id: str) -> Link:
        request = self.prepare_get(link_id)
        response = await self._client._prepared_request(request)
        return Link.model_validate(response)

    async def update(
        self,
        link_id: str,
        name: Optional[str] = None,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        password: Optional[str] = None,
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[str] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
        json_schema: Optional[Dict[str, Any]] = None,
    ) -> Link:
        request = self.prepare_update(link_id, name, webhook_url, webhook_headers, password, image_resolution_dpi, browser_canvas, modality, model, temperature, reasoning_effort, json_schema)
        response = await self._client._prepared_request(request)
        return Link.model_validate(response)

    async def delete(self, link_id: str) -> None:
        request = self.prepare_delete(link_id)
        await self._client._prepared_request(request)

    async def logs(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        link_id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> ListLogs:
        request = self.prepare_logs(before, after, limit, order, link_id, name, webhook_url, schema_id, schema_data_id)
        response = await self._client._prepared_request(request)
        return ListLogs.model_validate(response)