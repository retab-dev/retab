import datetime
from typing import Optional

from pydantic import BaseModel

from .._resource import AsyncAPIResource, SyncAPIResource
from ..types.ai_models import Amount

total_cost = 0.0

from openai.types.chat import completion_create_params
from openai.types.chat.chat_completion import ChatCompletion

from ..types.logs import AutomationLog, LogCompletionRequest


class Usage(SyncAPIResource):
    def total(self, start_date: Optional[datetime.datetime] = None, end_date: Optional[datetime.datetime] = None) -> Amount:
        """Get the total usage cost for a mailbox within an optional date range.

        Returns:
            Amount: The total usage cost
        """
        return Amount(value=total_cost, currency="USD")

    def mailbox(self, email: str, start_date: Optional[datetime.datetime] = None, end_date: Optional[datetime.datetime] = None) -> Amount:
        """Get the total usage cost for a mailbox within an optional date range.

        Args:
            email: The email address of the mailbox
            start_date: Start date for usage calculation
            end_date: End date for usage calculation

        Returns:
            Amount: The total usage cost
        """
        params = {}
        if start_date:
            params["start_date"] = start_date.isoformat()
        if end_date:
            params["end_date"] = end_date.isoformat()

        response = self._client._request("GET", f"/v1/deployments/mailboxes/{email}/usage", params=params)
        return Amount.model_validate(response)

    def link(self, link_id: str, start_date: Optional[datetime.datetime] = None, end_date: Optional[datetime.datetime] = None) -> Amount:
        """Get the total usage cost for a link within an optional date range.

        Args:
            link_id: The ID of the link
            start_date: Start date for usage calculation
            end_date: End date for usage calculation

        Returns:
            Amount: The total usage cost
        """
        params = {}
        if start_date:
            params["start_date"] = start_date.isoformat()
        if end_date:
            params["end_date"] = end_date.isoformat()

        response = self._client._request("GET", f"/v1/deployments/links/{link_id}/usage", params=params)
        return Amount.model_validate(response)

    def schema(self, schema_id: str, start_date: Optional[datetime.datetime] = None, end_date: Optional[datetime.datetime] = None) -> Amount:
        """Get the total usage cost for a schema within an optional date range.

        Args:
            schema_id: The ID of the schema
            start_date: Start date for usage calculation
            end_date: End date for usage calculation

        Returns:
            Amount: The total usage cost
        """
        params = {}
        if start_date:
            params["start_date"] = start_date.isoformat()
        if end_date:
            params["end_date"] = end_date.isoformat()

        response = self._client._request("GET", f"/v1/schemas/{schema_id}/usage", params=params)
        return Amount.model_validate(response)

    def schema_data(self, schema_data_id: str, start_date: Optional[datetime.datetime] = None, end_date: Optional[datetime.datetime] = None) -> Amount:
        """Get the total usage cost for a schema within an optional date range.

        Args:
            schema_id: The ID of the schema
            start_date: Start date for usage calculation
            end_date: End date for usage calculation

        Returns:
            Amount: The total usage cost
        """
        params = {}
        if start_date:
            params["start_date"] = start_date.isoformat()
        if end_date:
            params["end_date"] = end_date.isoformat()

        response = self._client._request("GET", f"/v1/schemas/{schema_data_id}/usage_data", params=params)
        return Amount.model_validate(response)

    # TODO: Turn that into an async process
    def log(self, response_format: completion_create_params.ResponseFormat, completion: ChatCompletion) -> AutomationLog:
        """Logs an openai request completion as an automation log to make the usage calculation possible for the user

        client = OpenAI()
        completion = client.beta.chat.completions.parse(
            model="gpt-4o-2024-08-06",
            messages=[
                {"role": "developer", "content": "Extract the event information."},
                {"role": "user", "content": "Alice and Bob are going to a science fair on Friday."},
            ],
            response_format=CalendarEvent,
        )
        uiclient.usage.log(
            response_format=CalendarEvent,
            completion=completion
        )


        Args:
            response_format: The response format of the openai request
            completion: The completion of the openai request

        Returns:
            AutomationLog: The automation log
        """

        if isinstance(response_format, BaseModel):
            log_completion_request = LogCompletionRequest(json_schema=response_format.model_json_schema(), completion=completion)

        if isinstance(response_format, dict):
            if "json_schema" in response_format:
                json_schema = response_format["json_schema"]  # type: ignore
                if "schema" in json_schema:
                    log_completion_request = LogCompletionRequest(json_schema=json_schema["schema"], completion=completion)
                else:
                    raise ValueError("Invalid response format")
            else:
                raise ValueError("Invalid response format")
        else:
            raise ValueError("Invalid response format")

        response = self._client._request("POST", "/v1/usage/log", data=log_completion_request.model_dump())

        return AutomationLog.model_validate(response)
