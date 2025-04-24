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
from ...types.automations.links import Link, ListLinks, UpdateLinkRequest
from ...types.documents.extractions import UiParsedChatCompletion
from ...types.image_settings import ImageSettings
from ...types.logs import AutomationLog, ExternalRequestLog, ListLogs
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
            "password": password,
            "image_settings": image_settings or ImageSettings(),
            "modality": modality,
            "model": model,
            "temperature": temperature,
            "reasoning_effort": reasoning_effort,
        }

        request = Link.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/automations/links", data=request.model_dump(mode='json'))

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

        return PreparedRequest(method="GET", url="/v1/automations", params=params)

    def prepare_get(self, link_id: str) -> PreparedRequest:
        """Get a specific extraction link configuration.

        Args:
            link_id: ID of the extraction link

        Returns:
            Link: The extraction link configuration
        """
        return PreparedRequest(method="GET", url=f"/v1/automations/{link_id}")

    def prepare_update(
        self,
        link_id: str,
        name: Optional[str] = None,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        password: Optional[str] = None,
        image_settings: Optional[Dict[str, Any]] = None,
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
        if image_settings is not None:
            data["image_settings"] = image_settings
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
        return PreparedRequest(method="PUT", url=f"/v1/automations/{link_id}", data=request.model_dump(mode='json'))

    def prepare_delete(self, link_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/automations/links/{link_id}", raise_for_status=True)

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

        return PreparedRequest(method="GET", url="/v1/automations/logs", params=params)


class Links(SyncAPIResource, LinksMixin):
    """Extraction Link API wrapper for managing extraction link configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.tests = TestLinks(client=client)

    def create(
        self,
        name: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        webhook_headers: Optional[Dict[str, str]] = None,
        password: str | None = None,
        # DocumentExtraction Config
        image_settings: Optional[Dict[str, Any]] = None,
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
            image_settings: Optional image preprocessing operations
            modality: Processing modality (currently only "native" supported)
            model: AI model to use for processing
            temperature: Model temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
        Returns:
            Link: The created extraction link configuration
        """

        request = self.prepare_create(name, json_schema, webhook_url, webhook_headers, password, image_settings, modality, model, temperature, reasoning_effort)
        response = self._client._prepared_request(request)
        print(f"Extraction Link Created. Link available at https://uiform.com/links/{response['id']}")
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
        image_settings: Optional[Dict[str, Any]] = None,
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
            image_settings: New image preprocessing operations
            modality: New processing modality
            model: New AI model
            temperature: New temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
            json_schema: New JSON schema

        Returns:
            Link: The updated extraction link configuration
        """

        request = self.prepare_update(link_id, name, webhook_url, webhook_headers, password, image_settings, modality, model, temperature, reasoning_effort, json_schema)
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
        print(f"Extraction Link Deleted. Link https://uiform.com/links/{link_id} is no longer available.")

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
        self.tests = AsyncTestLinks(client=client)

    async def create(
        self,
        name: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        webhook_headers: Optional[Dict[str, str]] = None,
        password: str | None = None,
        image_settings: Optional[Dict[str, Any]] = None,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
    ) -> Link:
        request = self.prepare_create(name, json_schema, webhook_url, webhook_headers, password, image_settings, modality, model, temperature, reasoning_effort)
        response = await self._client._prepared_request(request)
        print(f"Extraction Link Created. Link available at https://uiform.com/links/{response['id']}")
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
        image_settings: Optional[Dict[str, Any]] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        reasoning_effort: Optional[ChatCompletionReasoningEffort] = None,
        json_schema: Optional[Dict[str, Any]] = None,
    ) -> Link:
        request = self.prepare_update(link_id, name, webhook_url, webhook_headers, password, image_settings, modality, model, temperature, reasoning_effort, json_schema)
        response = await self._client._prepared_request(request)
        return Link.model_validate(response)

    async def delete(self, link_id: str) -> None:
        request = self.prepare_delete(link_id)
        await self._client._prepared_request(request)
        print(f"Extraction Link Deleted. Link https://uiform.com/links/{link_id} is no longer available.")

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


### Test Links ###


class TestLinksMixin:
    def prepare_upload(self, link_id: str, document: Path | str | IOBase | HttpUrl | Image | MIMEData) -> PreparedRequest:
        mime_document = prepare_mime_document(document)
        return PreparedRequest(method="POST", url=f"/v1/automations/links/tests/upload/{link_id}", data={"document": mime_document.model_dump(mode='json')})

    def prepare_webhook(self, link_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"/v1/automations/links/tests/webhook/{link_id}", data=None)

    def print_upload_verbose(self, log: AutomationLog) -> None:
        if log.external_request_log:
            print(f"\nTEST FILE UPLOAD RESULTS:")
            print(f"\n#########################")
            print(f"Status Code: {log.external_request_log.status_code}")
            print(f"Duration: {log.external_request_log.duration_ms:.2f}ms")

            if log.external_request_log.error:
                print(f"\nERROR: {log.external_request_log.error}")

            if log.external_request_log.response_body:
                print("\n--------------")
                print("RESPONSE BODY:")
                print("--------------")

                print(json.dumps(log.external_request_log.response_body, indent=2))
            if log.external_request_log.response_headers:
                print("\n--------------")
                print("RESPONSE HEADERS:")
                print("--------------")
                print(json.dumps(log.external_request_log.response_headers, indent=2))

    def print_webhook_verbose(self, log: AutomationLog) -> None:
        if log.external_request_log:
            print(f"\nTEST WEBHOOK RESULTS:")
            print(f"\n#########################")
            print(f"Status Code: {log.external_request_log.status_code}")
            print(f"Duration: {log.external_request_log.duration_ms:.2f}ms")

            if log.external_request_log.error:
                print(f"\nERROR: {log.external_request_log.error}")

            if log.external_request_log.response_body:
                print("\n--------------")
                print("RESPONSE BODY:")
                print("--------------")

                print(json.dumps(log.external_request_log.response_body, indent=2))
            if log.external_request_log.response_headers:
                print("\n--------------")
                print("RESPONSE HEADERS:")
                print("--------------")
                print(json.dumps(log.external_request_log.response_headers, indent=2))


clean_response = UiParsedChatCompletion.model_validate(
    {
        'id': 'chatcmpl-xxxxxxxxxx',
        'choices': [
            {
                'finish_reason': 'stop',
                'index': 0,
                'logprobs': None,
                'message': {
                    'content': '{"name": "ACME Corporation", "type": "corporate", "relationship": "client"}',
                    'refusal': None,
                    'role': 'assistant',
                    'audio': None,
                    'function_call': None,
                    'tool_calls': [],
                    'parsed': {},
                },
            }
        ],
        'created': 1738084776,
        'model': 'gpt-4o-mini-2024-07-18',
        'object': 'chat.completion',
        'service_tier': 'default',
        'system_fingerprint': 'fp_xxxxxx',
        'usage': {
            'completion_tokens': 17,
            'prompt_tokens': 663,
            'total_tokens': 680,
            'completion_tokens_details': {'accepted_prediction_tokens': 0, 'audio_tokens': 0, 'reasoning_tokens': 0, 'rejected_prediction_tokens': 0},
            'prompt_tokens_details': {'audio_tokens': 0, 'cached_tokens': 0},
        },
        'likelihoods': {'name': 1.0, 'type': 0.9999993295729247, 'relationship': 0.9999920581810364},
    }
)


invoice_file_content = b"""
            INVOICE
            
            Invoice Number: INV-2024-001
            Date: March 15, 2024
            
            Bill To:
            ACME Corporation
            123 Business Street
            Business City, BC 12345
            
            Items:
            1. Consulting Services - $1,000
            2. Software License - $500
            3. Support Package - $250
            
            Subtotal: $1,750
            Tax (10%): $175
            Total: $1,925
            
            Payment Terms: Net 30
            Due Date: April 14, 2024
            """


class TestLinks(SyncAPIResource, TestLinksMixin):
    """Test Extraction Link API wrapper for testing extraction link configurations"""

    def upload(self, link_id: str, document: Path | str | IOBase | HttpUrl | Image | MIMEData, verbose: bool = True) -> AutomationLog:
        """Mock endpoint that simulates the complete extraction process with sample data.

        Args:
            link_id: ID of the extraction link to mock

        Returns:
            DocumentExtractResponse: The simulated extraction response
        """

        request = self.prepare_upload(link_id, document)
        response = self._client._prepared_request(request)

        log = AutomationLog.model_validate(response)

        if verbose:
            self.print_upload_verbose(log)

        return log

    def webhook(self, link_id: str, webhook_url: HttpUrl | None = None, verbose: bool = True) -> AutomationLog:
        """Mock endpoint that simulates the complete webhook process with sample data.

        Args:
            link_id: ID of the extraction link to mock
            webhook_url: Optional URL to send webhook to for local testing
            verbose: Whether to print verbose output

        Returns:
            AutomationLog: The simulated webhook response
        """
        if webhook_url:
            # Step 1: get automation object
            link = self._client.automations.links.get(link_id)

            # Create sample invoice content

            # Create MIME document
            mime_document = prepare_mime_document(invoice_file_content)

            # Create sample extraction response
            # Send webhook request
            start_time = datetime.datetime.now(datetime.timezone.utc)
            webhook_response = None
            error_message = None
            status_code = None
            response_body: Dict[str, Any] = {}
            response_headers: Dict[str, str] = {}

            try:
                with httpx.Client() as client:
                    webhook_response = client.post(
                        str(webhook_url),
                        json={"completion": clean_response.model_dump(mode="json"), "file_payload": BaseMIMEData.model_validate(mime_document).model_dump(mode="json")},
                    )
                    webhook_response.raise_for_status()
                    status_code = webhook_response.status_code
                    response_body = webhook_response.json() if webhook_response.text else {}
                    response_headers = dict(webhook_response.headers)
            except Exception as e:
                error_message = f"Error sending webhook: {str(e)}"
                status_code = 500
                response_body = {"error": str(e)}

            end_time = datetime.datetime.now(datetime.timezone.utc)

            # Create log entry
            log = AutomationLog(
                user_email=None,
                automation_snapshot=link,
                file_metadata=BaseMIMEData.model_validate(mime_document),
                completion=clean_response,
                organization_id="",
                external_request_log=ExternalRequestLog(
                    webhook_url=webhook_url,
                    request_body={"completion": clean_response.model_dump(mode="json")},
                    request_headers={},
                    request_at=start_time,
                    response_body=response_body,
                    response_headers=response_headers,
                    response_at=end_time,
                    status_code=status_code or 0,
                    error=error_message,
                    duration_ms=(end_time - start_time).total_seconds() * 1000,
                ),
            )

            if verbose:
                self.print_webhook_verbose(log)

            return log

        request = self.prepare_webhook(link_id)
        response = self._client._prepared_request(request)
        log = AutomationLog.model_validate(response)
        if verbose:
            self.print_webhook_verbose(log)
        return log


class AsyncTestLinks(AsyncAPIResource, TestLinksMixin):
    """Async Test Extraction Link API wrapper for testing extraction link configurations"""

    async def upload(self, link_id: str, document: Path | str | IOBase | HttpUrl | Image | MIMEData, verbose: bool = True) -> AutomationLog:
        request = self.prepare_upload(link_id, document)
        response = await self._client._prepared_request(request)
        log = AutomationLog.model_validate(response)
        if verbose:
            self.print_upload_verbose(log)
        return log

    async def webhook(self, link_id: str, webhook_url: HttpUrl | None = None, verbose: bool = True) -> AutomationLog:
        # If webhook_url is provided, this is often to test the webhook endpoint locally (even if it does not correspond to the webhook url configured in the link)
        # If webhook_url is not provided, the webhook url configured in the link will be used (it has to be a real url and not a local url)

        if webhook_url:
            # Step 1: get automation object
            link = await self._client.automations.links.get(link_id)

            # Create sample invoice content

            # Create MIME document
            mime_document = prepare_mime_document(invoice_file_content)

            # Send webhook request
            start_time = datetime.datetime.now(datetime.timezone.utc)
            webhook_response = None
            error_message = None
            status_code = None
            response_body: Dict[str, Any] = {}
            response_headers: Dict[str, str] = {}

            try:
                async with httpx.AsyncClient() as client:
                    webhook_response = await client.post(
                        str(webhook_url),
                        json={"completion": clean_response.model_dump(mode="json"), "file_payload": BaseMIMEData.model_validate(mime_document).model_dump(mode="json")},
                    )
                    webhook_response.raise_for_status()
                    status_code = webhook_response.status_code
                    response_body = webhook_response.json() if webhook_response.text else {}
                    response_headers = dict(webhook_response.headers)
            except Exception as e:
                error_message = f"Error sending webhook: {str(e)}"
                status_code = 500
                response_body = {"error": str(e)}

            end_time = datetime.datetime.now(datetime.timezone.utc)

            # Create log entry
            log = AutomationLog(
                user_email=None,
                automation_snapshot=link,
                file_metadata=BaseMIMEData.model_validate(mime_document),
                completion=clean_response,
                organization_id="",
                external_request_log=ExternalRequestLog(
                    webhook_url=webhook_url,
                    request_body={"completion": clean_response.model_dump(mode="json")},
                    request_headers={},
                    request_at=start_time,
                    response_body=response_body,
                    response_headers=response_headers,
                    response_at=end_time,
                    status_code=status_code or 0,
                    error=error_message,
                    duration_ms=(end_time - start_time).total_seconds() * 1000,
                ),
            )

            if verbose:
                self.print_webhook_verbose(log)

            return log

        request = self.prepare_webhook(link_id)
        response = await self._client._prepared_request(request)
        log = AutomationLog.model_validate(response)
        if verbose:
            self.print_webhook_verbose(log)
        return log
