from typing import Any, Literal, Optional

from pydantic_core import PydanticUndefined

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.automations.links import Link, ListLinks, UpdateLinkRequest
from ....types.standards import PreparedRequest


class LinksMixin:
    links_base_url: str = "/v1/processors/automations/links"

    def prepare_create(
        self,
        processor_id: str,
        name: str,
        webhook_url: str,
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
        password: str | None = PydanticUndefined,  # type: ignore[assignment]
    ) -> PreparedRequest:
        request = Link(
            processor_id=processor_id,
            name=name,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            password=password,
        )
        return PreparedRequest(method="POST", url=self.links_base_url, data=request.model_dump(mode="json"))

    def prepare_list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        # Filtering parameters
        processor_id: Optional[str] = None,
        name: Optional[str] = None,
    ) -> PreparedRequest:
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "processor_id": processor_id,
            "name": name,
        }
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}

        return PreparedRequest(method="GET", url=self.links_base_url, params=params)

    def prepare_get(self, link_id: str) -> PreparedRequest:
        """Get a specific extraction link configuration.

        Args:
            link_id: ID of the extraction link

        Returns:
            Link: The extraction link configuration
        """
        return PreparedRequest(method="GET", url=f"{self.links_base_url}/{link_id}")

    def prepare_update(
        self,
        link_id: str,
        name: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_url: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
        password: str | None = PydanticUndefined,  # type: ignore[assignment]
    ) -> PreparedRequest:
        request = UpdateLinkRequest(
            name=name,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            password=password,
        )
        return PreparedRequest(method="PUT", url=f"{self.links_base_url}/{link_id}", data=request.model_dump(mode="json", exclude_unset=True, exclude_defaults=True))

    def prepare_delete(self, link_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"{self.links_base_url}/{link_id}", raise_for_status=True)


class Links(SyncAPIResource, LinksMixin):
    """Extraction Link API wrapper for managing extraction link configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    def create(
        self,
        processor_id: str,
        name: str,
        webhook_url: str,
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
        password: str | None = PydanticUndefined,  # type: ignore[assignment]
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

        request = self.prepare_create(
            processor_id=processor_id,
            name=name,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            password=password,
        )
        response = self._client._prepared_request(request)

        print(f"Link Created. Url: https://www.retab.com/dashboard/processors/automations/{response['id']}")
        return Link.model_validate(response)

    def list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        # Filtering parameters
        processor_id: Optional[str] = None,
        name: Optional[str] = None,
    ) -> ListLinks:
        """List extraction link configurations with pagination support.

        Args:
            before: Optional cursor for pagination before a specific link ID
            after: Optional cursor for pagination after a specific link ID
            limit: Optional limit on number of results (max 100)
            order: Optional sort order ("asc" or "desc")
            processor_id: Optional filter by processor ID
            name: Optional filter by link name

        Returns:
            ListLinks: Paginated list of extraction link configurations with metadata
        """
        request = self.prepare_list(before=before, after=after, limit=limit, order=order, processor_id=processor_id, name=name)
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
        name: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_url: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        password: str | None = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
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

        request = self.prepare_update(
            link_id=link_id,
            name=name,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            password=password,
            need_validation=need_validation,
        )
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


class AsyncLinks(AsyncAPIResource, LinksMixin):
    """Async Extraction Link API wrapper for managing extraction link configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    async def create(
        self,
        processor_id: str,
        name: str,
        webhook_url: str,
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
        password: str | None = PydanticUndefined,  # type: ignore[assignment]
    ) -> Link:
        request = self.prepare_create(
            processor_id=processor_id,
            name=name,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            password=password,
        )
        response = await self._client._prepared_request(request)
        print(f"Link Created. Url: https://www.retab.com/dashboard/processors/automations/{response['id']}")
        return Link.model_validate(response)

    async def list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        processor_id: Optional[str] = None,
        name: Optional[str] = None,
    ) -> ListLinks:
        request = self.prepare_list(before=before, after=after, limit=limit, order=order, processor_id=processor_id, name=name)
        response = await self._client._prepared_request(request)
        return ListLinks.model_validate(response)

    async def get(self, link_id: str) -> Link:
        request = self.prepare_get(link_id)
        response = await self._client._prepared_request(request)
        return Link.model_validate(response)

    async def update(
        self,
        link_id: str,
        name: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_url: str = PydanticUndefined,  # type: ignore[assignment]
        webhook_headers: dict[str, str] = PydanticUndefined,  # type: ignore[assignment]
        password: str | None = PydanticUndefined,  # type: ignore[assignment]
        need_validation: bool = PydanticUndefined,  # type: ignore[assignment]
    ) -> Link:
        request = self.prepare_update(
            link_id=link_id,
            name=name,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            password=password,
            need_validation=need_validation,
        )
        response = await self._client._prepared_request(request)
        return Link.model_validate(response)

    async def delete(self, link_id: str) -> None:
        request = self.prepare_delete(link_id)
        await self._client._prepared_request(request)
