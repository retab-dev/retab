from typing import Any, Literal, List

from ....types.standards import FieldUnset

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.automations.outlook import (
    FetchParams,
    ListOutlooks,
    MatchParams,
    Outlook,
    UpdateOutlookRequest,
)
from ....types.standards import PreparedRequest


class OutlooksMixin:
    outlooks_base_url: str = "/v1/processors/automations/outlook"

    def prepare_create(
        self,
        name: str,
        processor_id: str,
        webhook_url: str,
        default_language: str = FieldUnset,
        webhook_headers: dict[str, str] = FieldUnset,
        need_validation: bool = FieldUnset,
        authorized_domains: list[str] = FieldUnset,
        authorized_emails: list[str] = FieldUnset,
        layout_schema: dict[str, Any] = FieldUnset,
        match_params: list[MatchParams] = FieldUnset,
        fetch_params: list[FetchParams] = FieldUnset,
    ) -> PreparedRequest:
        # Build outlook dictionary with only provided fields
        outlook_dict: dict[str, Any] = {
            'processor_id': processor_id,
            'name': name,
            'webhook_url': webhook_url,
        }
        if default_language is not FieldUnset:
            outlook_dict['default_language'] = default_language
        if webhook_headers is not FieldUnset:
            outlook_dict['webhook_headers'] = webhook_headers
        if need_validation is not FieldUnset:
            outlook_dict['need_validation'] = need_validation
        if authorized_domains is not FieldUnset:
            outlook_dict['authorized_domains'] = authorized_domains
        if authorized_emails is not FieldUnset:
            outlook_dict['authorized_emails'] = authorized_emails
        if layout_schema is not FieldUnset:
            outlook_dict['layout_schema'] = layout_schema
        if match_params is not FieldUnset:
            outlook_dict['match_params'] = match_params
        if fetch_params is not FieldUnset:
            outlook_dict['fetch_params'] = fetch_params

        # Validate the data
        outlook_data = Outlook(**outlook_dict)
        return PreparedRequest(
            method="POST",
            url=self.outlooks_base_url,
            data=outlook_data.model_dump(mode="json", exclude_unset=True),
        )

    def prepare_list(
        self,
        processor_id: str,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        name: str | None = None,
        webhook_url: str | None = None,
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

        return PreparedRequest(method="GET", url=self.outlooks_base_url, params=params)

    def prepare_get(self, outlook_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{self.outlooks_base_url}/{outlook_id}")

    def prepare_update(
        self,
        outlook_id: str,
        name: str = FieldUnset,
        default_language: str = FieldUnset,
        webhook_url: str = FieldUnset,
        webhook_headers: dict[str, str] = FieldUnset,
        need_validation: bool = FieldUnset,
        authorized_domains: list[str] = FieldUnset,
        authorized_emails: list[str] = FieldUnset,
        match_params: list[MatchParams] = FieldUnset,
        fetch_params: list[FetchParams] = FieldUnset,
        layout_schema: dict[str, Any] = FieldUnset,
    ) -> PreparedRequest:
        update_dict: dict[str, Any] = {}
        if name is not FieldUnset:
            update_dict['name'] = name
        if default_language is not FieldUnset:
            update_dict['default_language'] = default_language
        if webhook_url is not FieldUnset:
            update_dict['webhook_url'] = webhook_url
        if webhook_headers is not FieldUnset:
            update_dict['webhook_headers'] = webhook_headers
        if need_validation is not FieldUnset:
            update_dict['need_validation'] = need_validation
        if authorized_domains is not FieldUnset:
            update_dict['authorized_domains'] = authorized_domains
        if authorized_emails is not FieldUnset:
            update_dict['authorized_emails'] = authorized_emails
        if layout_schema is not FieldUnset:
            update_dict['layout_schema'] = layout_schema
        if match_params is not FieldUnset:
            update_dict['match_params'] = match_params
        if fetch_params is not FieldUnset:
            update_dict['fetch_params'] = fetch_params

        update_outlook_request = UpdateOutlookRequest(**update_dict)

        return PreparedRequest(
            method="PUT",
            url=f"{self.outlooks_base_url}/{outlook_id}",
            data=update_outlook_request.model_dump(mode="json", exclude_unset=True),
        )

    def prepare_delete(self, outlook_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"{self.outlooks_base_url}/{outlook_id}")


class Outlooks(SyncAPIResource, OutlooksMixin):
    """Outlook API wrapper for managing outlook automation configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    def create(
        self,
        name: str,
        processor_id: str,
        webhook_url: str,
        default_language: str = FieldUnset,
        webhook_headers: dict[str, str] = FieldUnset,
        need_validation: bool = FieldUnset,
        authorized_domains: list[str] = FieldUnset,
        authorized_emails: list[str] = FieldUnset,
        layout_schema: dict[str, Any] = FieldUnset,
        match_params: list[MatchParams] = FieldUnset,
        fetch_params: list[FetchParams] = FieldUnset,
    ) -> Outlook:
        """Create a new outlook automation configuration.

        Args:
            name: Name of the outlook plugin
            processor_id: ID of the processor to use for the automation
            webhook_url: Webhook URL to receive processed data
            webhook_headers: Webhook headers to send with processed data
            authorized_domains: List of authorized domains
            authorized_emails: List of authorized emails
            layout_schema: Layout schema to display the data
            match_params: List of match parameters for the outlook automation
            fetch_params: List of fetch parameters for the outlook automation
        Returns:
            Outlook: The created outlook plugin configuration
        """

        request = self.prepare_create(
            name=name,
            processor_id=processor_id,
            webhook_url=webhook_url,
            default_language=default_language,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            authorized_domains=authorized_domains,
            authorized_emails=authorized_emails,
            layout_schema=layout_schema,
            match_params=match_params,
            fetch_params=fetch_params,
        )
        response = self._client._prepared_request(request)

        print(f"Outlook plugin created. Url: https://www.retab.com/dashboard/processors/automations/{response['id']}")

        return Outlook.model_validate(response)

    def list(
        self,
        processor_id: str,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        name: str | None = None,
        webhook_url: str | None = None,
    ) -> ListOutlooks:
        """List all outlook automation configurations.

        Args:
            before: Optional cursor for pagination - get results before this log ID
            after: Optional cursor for pagination - get results after this log ID
            limit: Maximum number of logs to return (1-100, default 10)
            order: Sort order by creation time - "asc" or "desc" (default "desc")
            name: Optional name filter
            webhook_url: Optional webhook URL filter
        Returns:
            List[Outlook]: List of outlook plugin configurations
        """
        request = self.prepare_list(processor_id, before, after, limit, order, name, webhook_url)
        response = self._client._prepared_request(request)
        return ListOutlooks.model_validate(response)

    def get(self, outlook_id: str) -> Outlook:
        """Get a specific outlook automation configuration.

        Args:
            id: ID of the outlook plugin

        Returns:
            Outlook: The outlook plugin configuration
        """
        request = self.prepare_get(outlook_id)
        response = self._client._prepared_request(request)
        return Outlook.model_validate(response)

    def update(
        self,
        outlook_id: str,
        name: str = FieldUnset,
        default_language: str = FieldUnset,
        webhook_url: str = FieldUnset,
        webhook_headers: dict[str, str] = FieldUnset,
        need_validation: bool = FieldUnset,
        authorized_domains: List[str] = FieldUnset,
        authorized_emails: List[str] = FieldUnset,
        layout_schema: dict[str, Any] = FieldUnset,
        match_params: List[MatchParams] = FieldUnset,
        fetch_params: List[FetchParams] = FieldUnset,
    ) -> Outlook:
        """Update an outlook automation configuration.

        Args:
            outlook_id: ID of the outlook plugin to update
            name: New name for the outlook plugin
            webhook_url: New webhook URL
            webhook_headers: New webhook headers
            authorized_domains: New authorized domains
            authorized_emails: New authorized emails
            match_params: New match parameters for the outlook automation
            fetch_params: New fetch parameters for the outlook automation
            layout_schema: New layout schema for the outlook automation

        Returns:
            Outlook: The updated outlook plugin configuration
        """
        request = self.prepare_update(
            outlook_id,
            name=name,
            default_language=default_language,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            authorized_domains=authorized_domains,
            authorized_emails=authorized_emails,
            layout_schema=layout_schema,
            match_params=match_params,
            fetch_params=fetch_params,
        )
        response = self._client._prepared_request(request)
        return Outlook.model_validate(response)

    def delete(self, outlook_id: str) -> None:
        """Delete an outlook automation configuration.

        Args:
            outlook_id: ID of the outlook plugin to delete
        """
        request = self.prepare_delete(outlook_id)
        self._client._prepared_request(request)
        return None


class AsyncOutlooks(AsyncAPIResource, OutlooksMixin):
    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    async def create(
        self,
        name: str,
        processor_id: str,
        webhook_url: str,
        default_language: str = FieldUnset,
        webhook_headers: dict[str, str] = FieldUnset,
        need_validation: bool = FieldUnset,
        authorized_domains: list[str] = FieldUnset,
        authorized_emails: list[str] = FieldUnset,
        layout_schema: dict[str, Any] = FieldUnset,
        match_params: list[MatchParams] = FieldUnset,
        fetch_params: list[FetchParams] = FieldUnset,
    ) -> Outlook:
        request = self.prepare_create(
            name=name,
            processor_id=processor_id,
            webhook_url=webhook_url,
            default_language=default_language,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            authorized_domains=authorized_domains,
            authorized_emails=authorized_emails,
            layout_schema=layout_schema,
            match_params=match_params,
            fetch_params=fetch_params,
        )
        response = await self._client._prepared_request(request)
        print(f"Outlook plugin created. Url: https://www.retab.com/dashboard/processors/automations/{response['id']}")
        return Outlook.model_validate(response)

    async def list(
        self,
        processor_id: str,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc",
        name: str | None = None,
        webhook_url: str | None = None,
    ) -> ListOutlooks:
        request = self.prepare_list(processor_id, before, after, limit, order, name, webhook_url)
        response = await self._client._prepared_request(request)
        return ListOutlooks.model_validate(response)

    async def get(self, outlook_id: str) -> Outlook:
        request = self.prepare_get(outlook_id)
        response = await self._client._prepared_request(request)
        return Outlook.model_validate(response)

    async def update(
        self,
        outlook_id: str,
        name: str = FieldUnset,
        default_language: str = FieldUnset,
        webhook_url: str = FieldUnset,
        webhook_headers: dict[str, str] = FieldUnset,
        need_validation: bool = FieldUnset,
        authorized_domains: List[str] = FieldUnset,
        authorized_emails: List[str] = FieldUnset,
        layout_schema: dict[str, Any] = FieldUnset,
        match_params: List[MatchParams] = FieldUnset,
        fetch_params: List[FetchParams] = FieldUnset,
    ) -> Outlook:
        request = self.prepare_update(
            outlook_id=outlook_id,
            name=name,
            default_language=default_language,
            webhook_url=webhook_url,
            webhook_headers=webhook_headers,
            need_validation=need_validation,
            authorized_domains=authorized_domains,
            authorized_emails=authorized_emails,
            layout_schema=layout_schema,
            match_params=match_params,
            fetch_params=fetch_params,
        )
        response = await self._client._prepared_request(request)
        return Outlook.model_validate(response)

    async def delete(self, outlook_id: str) -> None:
        request = self.prepare_delete(outlook_id)
        await self._client._prepared_request(request)
        return None
