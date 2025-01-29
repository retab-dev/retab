from typing import Any, Optional, Literal, List, Dict
from pydantic import HttpUrl, EmailStr
import json
from PIL.Image import Image
from pathlib import Path
from io import IOBase

from ..._resource import SyncAPIResource, AsyncAPIResource

from ...types.modalities import Modality
from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.documents.image_operations import ImageOperations
from ...types.documents.text_operations import TextOperations

from ..._utils.mime import prepare_mime_document
from ...types.automations.automations import ExtractionLinkConfig, UpdateExtractionLinkRequest, AutomationLog

from ...types.mime import MIMEData

class ExtractionLink(SyncAPIResource):
    """Extraction Link API wrapper for managing extraction link configurations"""
    def create(
        self,
        name: str,
        json_schema: Dict[str, Any],
        webhook_url: HttpUrl,
        webhook_headers: Optional[Dict[str, str]] = None,
        password: str | None = None,

        # DocumentExtraction Config
        text_operations: Optional[Dict[str, Any]] = None,
        image_operations: Optional[Dict[str, Any]] = None,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,

    ) -> ExtractionLinkConfig:
        """Create a new extraction link configuration.
        
        Args:
            name: Name of the extraction link
            json_schema: JSON schema to validate extracted data
            webhook_url: Webhook endpoint for forwarding processed files
            webhook_headers: Optional HTTP headers for webhook requests
            max_file_size: Optional maximum file size in MB
            file_payload: Optional flag to forward original file
            protection_type: Protection type for the link
            password: Optional password for protected links
            invitations: Optional list of authorized email addresses
            text_operations: Optional text preprocessing operations
            image_operations: Optional image preprocessing operations
            modality: Processing modality (currently only "native" supported)
            model: AI model to use for processing
            temperature: Model temperature setting
            
        Returns:
            ExtractionLinkConfig: The created extraction link configuration
        """

        data = {
            "name": name,
            "webhook_url": webhook_url,
            "webhook_headers": webhook_headers or {},
            "json_schema": json_schema,
            "password": password,
            "text_operations": text_operations or TextOperations(),
            "image_operations": image_operations or ImageOperations(),
            "modality": modality,
            "model": model,
            "temperature": temperature,
        }

        request = ExtractionLinkConfig.model_validate(data)

        response = self._client._request("POST", "/api/v1/extraction-link/extraction-link", data=request.model_dump(mode='json'))
        
        return ExtractionLinkConfig.model_validate(response)

    def list(self) -> List[ExtractionLinkConfig]:
        """List all extraction link configurations.
        
        Returns:
            List[ExtractionLinkConfig]: List of extraction link configurations
        """
        response = self._client._request("GET", "/api/v1/extraction-link/extraction-link")

        return [ExtractionLinkConfig.model_validate(link) for link in response]

    def get(self, link_id: str) -> ExtractionLinkConfig:
        """Get a specific extraction link configuration.
        
        Args:
            link_id: ID of the extraction link
            
        Returns:
            ExtractionLinkConfig: The extraction link configuration
        """
        response = self._client._request("GET", f"/api/v1/extraction-link/extraction-link/{link_id}")
        return ExtractionLinkConfig.model_validate(response)

    def update(
        self,
        link_id: str,
        name: Optional[str] = None,
        webhook_url: Optional[HttpUrl] = None,
        webhook_headers: Optional[Dict[str, str]] = None,
        password: Optional[str] = None,
        text_operations: Optional[Dict[str, Any]] = None,
        image_operations: Optional[Dict[str, Any]] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        json_schema: Optional[Dict[str, Any]] = None
    ) -> ExtractionLinkConfig:
        """Update an extraction link configuration.
        
        Args:
            link_id: ID of the extraction link to update
            name: New name for the link
            webhook_url: New webhook endpoint URL
            webhook_headers: New webhook headers
            password: New password for protected links
            text_operations: New text preprocessing operations
            image_operations: New image preprocessing operations
            modality: New processing modality
            model: New AI model
            temperature: New temperature setting
            json_schema: New JSON schema
            
        Returns:
            ExtractionLinkConfig: The updated extraction link configuration
        """
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
        if text_operations is not None:
            data["text_operations"] = text_operations
        if image_operations is not None:
            data["image_operations"] = image_operations
        if modality is not None:
            data["modality"] = modality
        if model is not None:
            data["model"] = model
        if temperature is not None:
            data["temperature"] = temperature
        if json_schema is not None:
            data["json_schema"] = json_schema

        request = UpdateExtractionLinkRequest.model_validate(data)

        response = self._client._request("PUT", f"/api/v1/extraction-link/extraction-link/{link_id}", data=request.model_dump(mode='json'))

        return ExtractionLinkConfig.model_validate(response)

    def delete(self, link_id: str) -> None:
        """Delete an extraction link configuration.
        
        Args:
            link_id: ID of the extraction link to delete
            
        Returns:
            Dict[str, str]: Response message confirming deletion
        """
        self._client._request("DELETE", f"/api/v1/extraction-link/extraction-link/{link_id}")

    

    def test_document_upload(self, 
                         link_id: str,
                         document: Path | str | IOBase | HttpUrl | Image | MIMEData,
                         verbose: bool = True
                         ) -> AutomationLog:
        """Mock endpoint that simulates the complete extraction process with sample data.
        
        Args:
            link_id: ID of the extraction link to mock
            
        Returns:
            DocumentExtractResponse: The simulated extraction response
        """

        mime_document = prepare_mime_document(document)
        response = self._client._request("POST", f"/api/v1/extraction-link/extraction-link/test-document-upload/{link_id}", data={"document": mime_document.model_dump()})

        log = AutomationLog.model_validate(response)

        if verbose:
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


        return log
    

    def test_webhook(self, 
                          link_id: str,
                          verbose: bool = True
                          ) -> AutomationLog:
        """Mock endpoint that simulates the complete webhook process with sample data.
        
        Args:
            link_id: ID of the extraction link to mock
            
        Returns:
            AutomationLog: The simulated webhook response
        """

        response = self._client._request("POST", f"/api/v1/extraction-link/extraction-link/test-webhook/{link_id}")

        log = AutomationLog.model_validate(response)

        if verbose:
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

        return log