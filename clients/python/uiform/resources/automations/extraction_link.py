from typing import Any, Optional, Literal, List, Dict

from ..._resource import SyncAPIResource, AsyncAPIResource

from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.modalities import Modality

from typing import Any, Optional, Literal, List, Dict
from datetime import datetime
from pydantic import HttpUrl

from ..._resource import SyncAPIResource, AsyncAPIResource
from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.modalities import Modality

from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.documents.image_operations import ImageOperations
from ...types.documents.parse import DocumentExtractRequest, DocumentExtractResponse
from ...types.documents.text_operations import TextOperations

from pydantic import EmailStr

from ...types.automations.automations import ExtractionLinkConfig, UpdateExtractionLinkRequest

class ExtractionLink(SyncAPIResource):
    """Extraction Link API wrapper for managing extraction link configurations"""

    def create(
        self,
        json_schema: Dict[str, Any],
        # HTTP Config
        endpoint: HttpUrl,
        headers: Optional[Dict[str, str]] = None,
        max_file_size: Optional[int] = None,
        forward_file: Optional[bool] = None,
        # Link Config
        name: str,
        protection_type: Literal["none", "password", "invitations"] = "none",
        password: str | None = None,
        invitations: List[EmailStr] = [],


        # DocumentExtraction Config
        text_operations: Optional[Dict[str, Any]] = None,
        image_operations: Optional[Dict[str, Any]] = None,
        modality: Literal["native"] = "native",
        model: str = "gpt-4o-mini",
        temperature: float = 0,
        additional_messages: List[ChatCompletionUiformMessage] = []
    ) -> ExtractionLinkConfig:
        """Create a new extraction link configuration.
        
        Args:
            name: Name of the extraction link
            http_config: Webhook configuration for forwarding processed files
            json_schema: JSON schema to validate extracted data
            protection: Protection configuration for the link
            text_operations: Optional text preprocessing operations
            image_operations: Optional image preprocessing operations
            modality: Processing modality (currently only "native" supported)
            model: AI model to use for processing
            temperature: Model temperature setting
            additional_messages: Optional additional context messages
            
        Returns:
            ExtractionLinkConfig: The created extraction link configuration
        """
        data = {
            "name": name,
            "http_config": http_config,
            "json_schema": json_schema,
            "protection": protection or LinkProtection(),
            "text_operations": text_operations or TextOperations(),
            "image_operations": image_operations or ImageOperations(),
            "modality": modality,
            "model": model,
            "temperature": temperature,
            "additional_messages": additional_messages
        }

        request = ExtractionLinkConfig.model_validate(data)

        response = self._client._request("POST", "/api/v1/extraction-link/extraction-link", data=request.model_dump(mode='json'))
        
        return ExtractionLinkConfig.model_validate(response)

    def list(self) -> List[ExtractionLinkConfig]:
        """List all extraction link configurations.
        
        Returns:
            List[ExtractionLinkConfig]: List of extraction link configurations
        """
        response = self._client._request("GET", "/api/v1/extraction-link")

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
        http_config: Optional[Dict[str, Any]] = None,
        protection: Optional[Dict[str, Any]] = None,
        text_operations: Optional[Dict[str, Any]] = None,
        image_operations: Optional[Dict[str, Any]] = None,
        modality: Optional[Modality] = None,
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        additional_messages: Optional[List[ChatCompletionUiformMessage]] = None,
        json_schema: Optional[Dict[str, Any]] = None
    ) -> ExtractionLinkConfig:
        """Update an extraction link configuration.
        
        Args:
            link_id: ID of the extraction link to update
            name: New name for the link
            http_config: New webhook configuration
            protection: New protection configuration
            text_operations: New text preprocessing operations
            image_operations: New image preprocessing operations
            modality: New processing modality
            model: New AI model
            temperature: New temperature setting
            additional_messages: New context messages
            json_schema: New JSON schema
            
        Returns:
            ExtractionLinkConfig: The updated extraction link configuration
        """
        data: dict[str, Any] = {"id": link_id}
        if name is not None:
            data["name"] = name
        if http_config is not None:
            data["http_config"] = http_config
        if protection is not None:
            data["protection"] = protection
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
        if additional_messages is not None:
            data["additional_messages"] = additional_messages
        if json_schema is not None:
            data["json_schema"] = json_schema

        update_request = UpdateExtractionLinkRequest.model_validate(data)

        response = self._client._request("PUT", f"/api/v1/extraction-link/extraction-link/{link_id}", data=update_request.model_dump())

        return ExtractionLinkConfig.model_validate(response)

    def delete(self, link_id: str) -> Dict[str, str]:
        """Delete an extraction link configuration.
        
        Args:
            link_id: ID of the extraction link to delete
            
        Returns:
            Dict[str, str]: Response message confirming deletion
        """
        return self._client._request("DELETE", f"/api/v1/extraction-link/extraction-link/{link_id}")
    

    def test_file_upload(self, link_id: str) -> DocumentExtractResponse:
        """Mock endpoint that simulates the complete extraction process with sample data.
        
        Args:
            link_id: ID of the extraction link to mock
            
        Returns:
            DocumentExtractResponse: The simulated extraction response
        """
        response = self._client._request("POST", f"/api/v1/extraction-link/extraction-link/mock/{link_id}")
        return DocumentExtractResponse.model_validate(response)
    

    def test_http_request(self, link_id: str) -> DocumentExtractResponse:
        """Mock endpoint that simulates the complete extraction process with sample data.
        
        Args:
            link_id: ID of the extraction link to mock
            
        Returns:
            DocumentExtractResponse: The simulated extraction response
        """
        response = self._client._request("POST", f"/api/v1/extraction-link/extraction-link/mock/{link_id}")
        return DocumentExtractResponse.model_validate(response)