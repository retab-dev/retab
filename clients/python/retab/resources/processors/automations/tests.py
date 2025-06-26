import json
import base64
from io import IOBase
from pathlib import Path

from PIL.Image import Image
from pydantic import HttpUrl

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....utils.mime import prepare_mime_document
from ....types.logs import AutomationLog
from ....types.mime import MIMEData
from ....types.standards import PreparedRequest


class TestsMixin:
    def prepare_upload(self, automation_id: str, document: Path | str | IOBase | HttpUrl | Image | MIMEData) -> PreparedRequest:
        mime_document = prepare_mime_document(document)

        # Convert MIME document to file upload format (similar to processors client)
        files = {"file": (mime_document.filename, base64.b64decode(mime_document.content), mime_document.mime_type)}

        # Send as multipart form data with file upload
        return PreparedRequest(method="POST", url=f"/v1/processors/automations/tests/upload/{automation_id}", files=files)

    def prepare_webhook(self, automation_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"/v1/processors/automations/tests/webhook/{automation_id}", data=None)

    def print_upload_verbose(self, log: AutomationLog) -> None:
        if log.external_request_log:
            print("\nTEST FILE UPLOAD RESULTS:")
            print("\n#########################")
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
            print("\nTEST WEBHOOK RESULTS:")
            print("\n#########################")
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


class Tests(SyncAPIResource, TestsMixin):
    """Test API wrapper for testing automation configurations"""

    def upload(self, automation_id: str, document: Path | str | IOBase | HttpUrl | Image | MIMEData, verbose: bool = True) -> AutomationLog:
        """Test endpoint that simulates the complete extraction process with the provided document.

        Args:
            automation_id: ID of the automation to test
            document: Document to process
            verbose: Whether to print verbose output

        Returns:
            AutomationLog: The automation log with extraction results
        """
        request = self.prepare_upload(automation_id, document)
        response = self._client._prepared_request(request)

        log = AutomationLog.model_validate(response)

        if verbose:
            self.print_upload_verbose(log)

        return log

    def webhook(self, automation_id: str, verbose: bool = True) -> AutomationLog:
        """Test endpoint that simulates the complete webhook process with sample data.

        Args:
            automation_id: ID of the automation to test
            verbose: Whether to print verbose output

        Returns:
            AutomationLog: The automation log with webhook results
        """
        request = self.prepare_webhook(automation_id)
        response = self._client._prepared_request(request)

        log = AutomationLog.model_validate(response)

        if verbose:
            self.print_webhook_verbose(log)

        return log


class AsyncTests(AsyncAPIResource, TestsMixin):
    """Async Test API wrapper for testing deployment configurations"""

    async def upload(self, automation_id: str, document: Path | str | IOBase | HttpUrl | Image | MIMEData, verbose: bool = True) -> AutomationLog:
        """Test endpoint that simulates the complete extraction process with the provided document.

        Args:
            automation_id: ID of the automation to test
            document: Document to process
            verbose: Whether to print verbose output

        Returns:
            AutomationLog: The automation log with extraction results
        """
        request = self.prepare_upload(automation_id, document)
        response = await self._client._prepared_request(request)

        log = AutomationLog.model_validate(response)

        if verbose:
            self.print_upload_verbose(log)

        return log

    async def webhook(self, automation_id: str, verbose: bool = True) -> AutomationLog:
        """Test endpoint that simulates the complete webhook process with sample data.

        Args:
            automation_id: ID of the automation to test
            verbose: Whether to print verbose output

        Returns:
            AutomationLog: The automation log with webhook results
        """
        request = self.prepare_webhook(automation_id)
        response = await self._client._prepared_request(request)

        log = AutomationLog.model_validate(response)

        if verbose:
            self.print_webhook_verbose(log)

        return log
