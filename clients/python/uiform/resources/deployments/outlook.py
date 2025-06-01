import json
from io import IOBase
from pathlib import Path
from typing import Any, Dict, List, Literal, Optional

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from PIL.Image import Image
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.ai_models import assert_valid_model_extraction
from ...types.deployments.outlook import FetchParams, ListOutlooks, MatchParams, Outlook, UpdateOutlookRequest
from ...types.logs import DeploymentLog
from ...types.modalities import Modality
from ...types.standards import PreparedRequest


class OutlooksMixin:
    def prepare_create(
        self,
        name: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        # email specific opitonals Fields
        authorized_domains: List[str] = [],
        authorized_emails: List[str] = [],
        # HTTP Config Optional Fields
        webhook_headers: Dict[str, str] = {},
        # DocumentExtraction Config
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        # Optional Fields for data integration
        match_params: Optional[List[MatchParams]] = None,
        fetch_params: Optional[List[FetchParams]] = None,
        layout_schema: Optional[Dict[str, Any]] = None,
    ) -> PreparedRequest:
        assert_valid_model_extraction(model)

        data = {
            "name": name,
            "webhook_url": webhook_url,
            "webhook_headers": webhook_headers,
            "json_schema": json_schema,
            "authorized_domains": authorized_domains,
            "authorized_emails": authorized_emails,
            "image_resolution_dpi": image_resolution_dpi,
            "browser_canvas": browser_canvas,
            "modality": modality,
            "model": model,
            "temperature": temperature,
            "reasoning_effort": reasoning_effort,
            "layout_schema": layout_schema,
        }

        if match_params is not None:
            data["match_params"] = match_params
        if fetch_params is not None:
            data["fetch_params"] = fetch_params

        # Validate the data
        outlook_data = Outlook.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/deployments/outlook", data=outlook_data.model_dump(mode="json"))

    def prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        id: Optional[str] = None,
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
            "id": id,
            "name": name,
            "webhook_url": webhook_url,
            "schema_id": schema_id,
            "schema_data_id": schema_data_id,
        }
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}

        return PreparedRequest(method="GET", url="/v1/deployments/outlook", params=params)

    def prepare_get(self, id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/deployments/outlook/{id}")

    def prepare_update(
        self,
        id: str,
        name: Optional[str] = None,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        authorized_domains: Optional[List[str]] = None,
        authorized_emails: Optional[List[str]] = None,
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
        json_schema: Optional[Dict[str, Any]] = None,
        match_params: Optional[List[MatchParams]] = None,
        fetch_params: Optional[List[FetchParams]] = None,
        layout_schema: Optional[Dict[str, Any]] = None,
    ) -> PreparedRequest:
        data: dict[str, Any] = {}
        if name is not None:
            data["name"] = name
        if webhook_url is not None:
            data["webhook_url"] = webhook_url
        if webhook_headers is not None:
            data["webhook_headers"] = webhook_headers
        if authorized_domains is not None:
            data["authorized_domains"] = authorized_domains
        if authorized_emails is not None:
            data["authorized_emails"] = authorized_emails
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
        if match_params is not None:
            data["match_params"] = match_params
        if fetch_params is not None:
            data["fetch_params"] = fetch_params
        if reasoning_effort is not None:
            data["reasoning_effort"] = reasoning_effort
        if layout_schema is not None:
            data["layout_schema"] = layout_schema

        update_outlook_request = UpdateOutlookRequest.model_validate(data)

        return PreparedRequest(method="PUT", url=f"/v1/deployments/outlook/{id}", data=update_outlook_request.model_dump(mode="json"))

    def prepare_delete(self, id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/deployments/outlook/{id}")

    def prepare_logs(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
        match_params: Optional[List[MatchParams]] = None,
        fetch_params: Optional[List[FetchParams]] = None,
    ) -> PreparedRequest:
        params = {
            "id": id,
            "name": name,
            "webhook_url": webhook_url,
            "schema_id": schema_id,
            "schema_data_id": schema_data_id,
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "match_params": match_params,
            "fetch_params": fetch_params,
        }
        return PreparedRequest(method="GET", url=f"/v1/deployments/outlook/{id}/logs", params=params)


class Outlooks(SyncAPIResource, OutlooksMixin):
    """Outlook API wrapper for managing outlook automation configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    def create(
        self,
        name: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        authorized_domains: List[str] = [],
        authorized_emails: List[str] = [],
        webhook_headers: Dict[str, str] = {},
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        match_params: Optional[List[MatchParams]] = None,
        fetch_params: Optional[List[FetchParams]] = None,
    ) -> Outlook:
        """Create a new outlook automation configuration.

        Args:
            name: Name of the outlook plugin
            json_schema: JSON schema to validate extracted data
            webhook_url: Webhook URL to receive processed data
            webhook_headers: Webhook headers to send with processed data
            authorized_domains: List of authorized domains
            authorized_emails: List of authorized emails
            image_resolution_dpi: Optional image resolution DPI
            browser_canvas: Optional browser canvas size
            modality: Processing modality (currently only "native" supported)
            model: AI model to use for processing
            temperature: Model temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
            match_params: List of match parameters for the outlook automation
            fetch_params: List of fetch parameters for the outlook automation
        Returns:
            Outlook: The created outlook plugin configuration
        """

        request = self.prepare_create(
            name,
            json_schema,
            webhook_url,
            authorized_domains,
            authorized_emails,
            webhook_headers,
            image_resolution_dpi,
            browser_canvas,
            modality,
            model,
            temperature,
            reasoning_effort,
            match_params,
            fetch_params,
        )
        response = self._client._prepared_request(request)

        print(f"Outlook automation created. Outlook available at https://www.uiform.com/dashboard/deployments/{response['id']}")

        return Outlook.model_validate(response)

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> ListOutlooks:
        """List all outlook automation configurations.

        Args:
            before: Optional cursor for pagination - get results before this log ID
            after: Optional cursor for pagination - get results after this log ID
            limit: Maximum number of logs to return (1-100, default 10)
            order: Sort order by creation time - "asc" or "desc" (default "desc")
            id: Optional ID filter
            name: Optional name filter
            webhook_url: Optional webhook URL filter
            schema_id: Optional schema ID filter
            schema_data_id: Optional schema data ID filter
            match_params: Optional list of match parameters for the outlook automation
            fetch_params: Optional list of fetch parameters for the outlook automation
        Returns:
            List[Outlook]: List of outlook plugin configurations
        """
        request = self.prepare_list(before, after, limit, order, id, name, webhook_url, schema_id, schema_data_id)
        response = self._client._prepared_request(request)
        return ListOutlooks.model_validate(response)

    def get(self, id: str) -> Outlook:
        """Get a specific outlook automation configuration.

        Args:
            id: ID of the outlook plugin

        Returns:
            Outlook: The outlook plugin configuration
        """
        request = self.prepare_get(id)
        response = self._client._prepared_request(request)
        return Outlook.model_validate(response)

    def update(
        self,
        id: str,
        name: Optional[str] = None,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        authorized_domains: Optional[List[str]] = None,
        authorized_emails: Optional[List[str]] = None,
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
        json_schema: Optional[Dict[str, Any]] = None,
        match_params: Optional[List[MatchParams]] = None,
        fetch_params: Optional[List[FetchParams]] = None,
        layout_schema: Optional[Dict[str, Any]] = None,
    ) -> Outlook:
        """Update an outlook automation configuration.

        Args:
            id: ID of the outlook plugin to update
            name: New name for the outlook plugin
            webhook_url: New webhook URL
            webhook_headers: New webhook headers
            authorized_domains: New authorized domains
            authorized_emails: New authorized emails
            image_resolution_dpi: New image resolution DPI
            browser_canvas: New browser canvas size
            modality: New processing modality
            model: New AI model
            temperature: New temperature setting
            reasoning_effort: New reasoning effort
            json_schema: New JSON schema
            match_params: New match parameters for the outlook automation
            fetch_params: New fetch parameters for the outlook automation
            layout_schema: New layout schema for the outlook automation

        Returns:
            Outlook: The updated outlook plugin configuration
        """
        request = self.prepare_update(
            id,
            name,
            webhook_url,
            webhook_headers,
            authorized_domains,
            authorized_emails,
            image_resolution_dpi,
            browser_canvas,
            modality,
            model,
            temperature,
            reasoning_effort,
            json_schema,
            match_params,
            fetch_params,
            layout_schema,
        )
        response = self._client._prepared_request(request)
        return Outlook.model_validate(response)

    def delete(self, id: str) -> None:
        """Delete an outlook automation configuration.

        Args:
            id: ID of the outlook plugin to delete
        """
        request = self.prepare_delete(id)
        response = self._client._prepared_request(request)
        return None

    def logs(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
        match_params: Optional[List[MatchParams]] = None,
        fetch_params: Optional[List[FetchParams]] = None,
    ) -> List[DeploymentLog]:
        """Get logs for a specific outlook automation.

        Args:
            before: Optional cursor for pagination - get results before this log ID
            after: Optional cursor for pagination - get results after this log ID
            limit: Maximum number of logs to return (1-100, default 10)
            order: Sort order by creation time - "asc" or "desc" (default "desc")
            id: Optional ID filter
            webhook_url: Optional webhook URL filter
            schema_id: Optional schema ID filter
            schema_data_id: Optional schema data ID filter
            match_params: Optional list of match parameters for the outlook automation
            fetch_params: Optional list of fetch parameters for the outlook automation
        Returns:
            List[Dict[str, Any]]: List of log entries
        """
        request = self.prepare_logs(before, after, limit, order, id, name, webhook_url, schema_id, schema_data_id, match_params, fetch_params)
        response = self._client._prepared_request(request)
        return [DeploymentLog.model_validate(log) for log in response]


class AsyncOutlooks(AsyncAPIResource, OutlooksMixin):
    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    async def create(
        self,
        name: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        authorized_domains: List[str] = [],
        authorized_emails: List[str] = [],
        webhook_headers: Dict[str, str] = {},
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        match_params: List[MatchParams] = [],
        fetch_params: List[FetchParams] = [],
    ) -> Outlook:
        request = self.prepare_create(
            name,
            json_schema,
            webhook_url,
            authorized_domains,
            authorized_emails,
            webhook_headers,
            image_resolution_dpi,
            browser_canvas,
            modality,
            model,
            temperature,
            reasoning_effort,
            match_params,
            fetch_params,
        )
        response = await self._client._prepared_request(request)
        print(f"Outlook automation created. Outlook available at https://www.uiform.com/dashboard/deployments/{response['id']}")
        return Outlook.model_validate(response)

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> ListOutlooks:
        request = self.prepare_list(before, after, limit, order, id, name, webhook_url, schema_id, schema_data_id)
        response = await self._client._prepared_request(request)
        return ListOutlooks.model_validate(response)

    async def get(self, id: str) -> Outlook:
        request = self.prepare_get(id)
        response = await self._client._prepared_request(request)
        return Outlook.model_validate(response)

    async def update(
        self,
        id: str,
        name: Optional[str] = None,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        authorized_domains: Optional[List[str]] = None,
        authorized_emails: Optional[List[str]] = None,
        image_resolution_dpi: Optional[int] = None,
        browser_canvas: Optional[Literal['A3', 'A4', 'A5']] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
        json_schema: Optional[Dict[str, Any]] = None,
        match_params: Optional[List[MatchParams]] = None,
        fetch_params: Optional[List[FetchParams]] = None,
        layout_schema: Optional[Dict[str, Any]] = None,
    ) -> Outlook:
        request = self.prepare_update(
            id,
            name,
            webhook_url,
            webhook_headers,
            authorized_domains,
            authorized_emails,
            image_resolution_dpi,
            browser_canvas,
            modality,
            model,
            temperature,
            reasoning_effort,
            json_schema,
            match_params,
            fetch_params,
            layout_schema,
        )
        response = await self._client._prepared_request(request)
        return Outlook.model_validate(response)

    async def delete(self, id: str) -> None:
        request = self.prepare_delete(id)
        await self._client._prepared_request(request)
        return None

    async def logs(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        id: Optional[str] = None,
        name: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
        match_params: Optional[List[MatchParams]] = None,
        fetch_params: Optional[List[FetchParams]] = None,
    ) -> List[DeploymentLog]:
        request = self.prepare_logs(before, after, limit, order, id, name, webhook_url, schema_id, schema_data_id, match_params, fetch_params)
        response = await self._client._prepared_request(request)
        return [DeploymentLog.model_validate(log) for log in response]
