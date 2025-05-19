import datetime
import json
from io import IOBase
from pathlib import Path
from typing import Any, Dict, Literal, Optional

import httpx
from pydantic import HttpUrl
from PIL.Image import Image

from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.mime import prepare_mime_document
from ...types.mime import BaseMIMEData, MIMEData
from ...types.logs import DeploymentLog
from ...types.standards import PreparedRequest


class TestsMixin:
    def prepare_upload(self, id: str, document: Path | str | IOBase | HttpUrl | Image | MIMEData) -> PreparedRequest:
        mime_document = prepare_mime_document(document)
        return PreparedRequest(method="POST", url=f"/v1/deployments/tests/upload/{id}", data={"document": mime_document.model_dump(mode='json')})

    def prepare_webhook(self, id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"/v1/deployments/tests/webhook/{id}", data=None)

    def print_upload_verbose(self, log: DeploymentLog) -> None:
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

    def print_webhook_verbose(self, log: DeploymentLog) -> None:
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


class Tests(SyncAPIResource, TestsMixin):
    """Test API wrapper for testing automation configurations"""

    def upload(self, id: str, document: Path | str | IOBase | HttpUrl | Image | MIMEData, verbose: bool = True) -> DeploymentLog:
        """Test endpoint that simulates the complete extraction process with the provided document.

        Args:
            id: ID of the deployment to test
            document: Document to process
            verbose: Whether to print verbose output

        Returns:
            DeploymentLog: The automation log with extraction results
        """
        request = self.prepare_upload(id, document)
        response = self._client._prepared_request(request)
        
        log = DeploymentLog.model_validate(response)
        
        if verbose:
            self.print_upload_verbose(log)
            
        return log

    def webhook(self, id: str, verbose: bool = True) -> DeploymentLog:
        """Test endpoint that simulates the complete webhook process with sample data.

        Args:
            id: ID of the deployment to test
            verbose: Whether to print verbose output

        Returns:
            DeploymentLog: The automation log with webhook results
        """
        request = self.prepare_webhook(id)
        response = self._client._prepared_request(request)
        
        log = DeploymentLog.model_validate(response)
        
        if verbose:
            self.print_webhook_verbose(log)
            
        return log


class AsyncTests(AsyncAPIResource, TestsMixin):
    """Async Test API wrapper for testing deployment configurations"""

    async def upload(self, id: str, document: Path | str | IOBase | HttpUrl | Image | MIMEData, verbose: bool = True) -> DeploymentLog:
        """Test endpoint that simulates the complete extraction process with the provided document.

        Args:
            id: ID of the deployment to test
            document: Document to process
            verbose: Whether to print verbose output

        Returns:
            DeploymentLog: The automation log with extraction results
        """
        request = self.prepare_upload(id, document)
        response = await self._client._prepared_request(request)
        
        log = DeploymentLog.model_validate(response)
        
        if verbose:
            self.print_upload_verbose(log)
            
        return log

    async def webhook(self, id: str, verbose: bool = True) -> DeploymentLog:
        """Test endpoint that simulates the complete webhook process with sample data.

        Args:
            id: ID of the deployment to test
            verbose: Whether to print verbose output

        Returns:
            DeploymentLog: The automation log with webhook results
        """
        request = self.prepare_webhook(id)
        response = await self._client._prepared_request(request)
        
        log = DeploymentLog.model_validate(response)
        
        if verbose:
            self.print_webhook_verbose(log)
            
        return log
