from typing import Literal, Optional

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.logs import AutomationLog, ExternalRequestLog, ListLogs
from ....types.standards import PreparedRequest


class LogsMixin:
    def prepare_get(self, processor_id: str, log_id: str) -> PreparedRequest:
        """Get a specific automation log by ID.

        Args:
            processor_id: ID of the processor
            log_id: ID of the log to retrieve

        Returns:
            PreparedRequest: The prepared request
        """
        return PreparedRequest(method="GET", url=f"/v1/processors/{processor_id}/automations/logs/{log_id}")

    def prepare_list(
        self,
        processor_id: str,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        # Filtering parameters
        status_code: Optional[int] = None,
        status_class: Optional[Literal["2xx", "3xx", "4xx", "5xx"]] = None,
        automation_id: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> PreparedRequest:
        """List automation logs with pagination support.

        Args:
            processor_id: ID of the processor
            before: Optional cursor for pagination before a specific log ID
            after: Optional cursor for pagination after a specific log ID
            limit: Optional limit on number of results (max 100)
            order: Optional sort order ("asc" or "desc")
            status_code: Optional filter by status code
            status_class: Optional filter by status_class
            automation_id: Optional filter by deployment ID
            webhook_url: Optional filter by webhook URL
            schema_id: Optional filter by schema ID
            schema_data_id: Optional filter by schema data ID

        Returns:
            PreparedRequest: The prepared request
        """
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "status_code": status_code,
            "status_class": status_class,
            "automation_id": automation_id,
            "webhook_url": webhook_url,
            "schema_id": schema_id,
            "schema_data_id": schema_data_id,
        }
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}

        return PreparedRequest(method="GET", url=f"/v1/processors/{processor_id}/automations/logs", params=params)

    def prepare_rerun(self, processor_id: str, log_id: str) -> PreparedRequest:
        """Rerun a webhook from an existing AutomationLog.

        Args:
            processor_id: ID of the processor
            log_id: ID of the log to rerun

        Returns:
            PreparedRequest: The prepared request
        """
        return PreparedRequest(method="POST", url=f"/v1/processors/{processor_id}/automations/logs/{log_id}/rerun")


class Logs(SyncAPIResource, LogsMixin):
    """Logs API wrapper for managing automation logs"""

    def get(self, processor_id: str, log_id: str) -> AutomationLog:
        """Get a specific automation log by ID.

        Args:
            processor_id: ID of the processor
            log_id: ID of the log to retrieve

        Returns:
            AutomationLog: The automation log
        """
        request = self.prepare_get(processor_id, log_id)
        response = self._client._prepared_request(request)
        return AutomationLog.model_validate(response)

    def list(
        self,
        processor_id: str,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        status_code: Optional[int] = None,
        status_class: Optional[Literal["2xx", "3xx", "4xx", "5xx"]] = None,
        automation_id: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> ListLogs:
        """List automation logs with pagination support.

        Args:
            processor_id: ID of the processor
            before: Optional cursor for pagination before a specific log ID
            after: Optional cursor for pagination after a specific log ID
            limit: Optional limit on number of results (max 100)
            order: Optional sort order ("asc" or "desc")
            status_code: Optional filter by status code
            status_class: Optional filter by status_class
            automation_id: Optional filter by deployment ID
            webhook_url: Optional filter by webhook URL
            schema_id: Optional filter by schema ID
            schema_data_id: Optional filter by schema data ID

        Returns:
            ListLogs: Paginated list of automation logs with metadata
        """
        request = self.prepare_list(processor_id, before, after, limit, order, status_code, status_class, automation_id, webhook_url, schema_id, schema_data_id)
        response = self._client._prepared_request(request)
        return ListLogs.model_validate(response)

    def rerun(self, processor_id: str, log_id: str) -> ExternalRequestLog:
        """Rerun a webhook from an existing AutomationLog.

        Args:
            processor_id: ID of the processor
            log_id: ID of the log to rerun

        Returns:
            ExternalRequestLog: The result of the rerun webhook call
        """
        request = self.prepare_rerun(processor_id, log_id)
        response = self._client._prepared_request(request)

        print(f"Webhook call run successfully. Log available at https://www.retab.com/dashboard/processors/{processor_id}/logs/{log_id}")

        return ExternalRequestLog.model_validate(response)


class AsyncLogs(AsyncAPIResource, LogsMixin):
    """Async Logs API wrapper for managing automation logs"""

    async def get(self, processor_id: str, log_id: str) -> AutomationLog:
        """Get a specific automation log by ID.

        Args:
            processor_id: ID of the processor
            log_id: ID of the log to retrieve

        Returns:
            AutomationLog: The automation log
        """
        request = self.prepare_get(processor_id, log_id)
        response = await self._client._prepared_request(request)
        return AutomationLog.model_validate(response)

    async def list(
        self,
        processor_id: str,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: Optional[int] = 10,
        order: Optional[Literal["asc", "desc"]] = "desc",
        status_code: Optional[int] = None,
        status_class: Optional[Literal["2xx", "3xx", "4xx", "5xx"]] = None,
        automation_id: Optional[str] = None,
        webhook_url: Optional[str] = None,
        schema_id: Optional[str] = None,
        schema_data_id: Optional[str] = None,
    ) -> ListLogs:
        """List automation logs with pagination support.

        Args:
            before: Optional cursor for pagination before a specific log ID
            after: Optional cursor for pagination after a specific log ID
            limit: Optional limit on number of results (max 100)
            order: Optional sort order ("asc" or "desc")
            status_code: Optional filter by status code
            status_class: Optional filter by status_class
            automation_id: Optional filter by deployment ID
            webhook_url: Optional filter by webhook URL
            schema_id: Optional filter by schema ID
            schema_data_id: Optional filter by schema data ID

        Returns:
            ListLogs: Paginated list of automation logs with metadata
        """
        request = self.prepare_list(processor_id, before, after, limit, order, status_code, status_class, automation_id, webhook_url, schema_id, schema_data_id)
        response = await self._client._prepared_request(request)
        return ListLogs.model_validate(response)

    async def rerun(self, processor_id: str, log_id: str) -> ExternalRequestLog:
        """Rerun a webhook from an existing AutomationLog.

        Args:
            processor_id: ID of the processor
            log_id: ID of the log to rerun

        Returns:
            ExternalRequestLog: The result of the rerun webhook call
        """
        request = self.prepare_rerun(processor_id, log_id)
        response = await self._client._prepared_request(request)

        print(f"Webhook call run successfully. Log available at https://www.retab.com/dashboard/processors/{processor_id}/logs/{log_id}")

        return ExternalRequestLog.model_validate(response)
