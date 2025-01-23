from typing import Any
from pathlib import Path
from io import IOBase
import PIL.Image
from pydantic import HttpUrl


from ...types.modalities import Modality
from ...types.mime import MIMEData
from ..._utils.mime import prepare_mime_document, convert_mime_data_to_pil_image
from ..._resource import SyncAPIResource, AsyncAPIResource
from ...types.documents.create_messages import DocumentCreateMessageRequest, DocumentMessage
from .extractions import Extractions, AsyncExtractions

from .templates.templates import Templates, AsyncTemplates

class BaseDocumentsMixin:
    def _prepare_create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        modality: Modality = "native", 
        text_operations: dict[str, Any] | None = None,
        image_operations: dict[str, Any] | None = None
    ) -> DocumentCreateMessageRequest:
        
        mime_document = prepare_mime_document(document)
        data: dict[str, Any] = {
            "document": mime_document.model_dump(),
            "modality": modality,
        }
        if text_operations:
            data["text_operations"] = text_operations
        if image_operations:
            data["image_operations"] = image_operations

        return DocumentCreateMessageRequest.model_validate(data)


class Documents(SyncAPIResource, BaseDocumentsMixin): 
    """Documents API wrapper"""

    # TODO: Add batch methods
    #client.documents.batch.extract()
    #client.documents.batch.preprocess()

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.extractions = Extractions(client=client)
        self.templates = Templates(client=client)
        # self.batch = Batch(client=client)


    def correct_image_orientation(self, document: Path | str | IOBase | MIMEData | PIL.Image.Image) -> PIL.Image.Image:
        """Corrects the orientation of an image using the UiForm API.

        This method takes an image in various formats and returns a PIL Image with corrected orientation.
        Useful for handling images from mobile devices or cameras that may have incorrect EXIF orientation.

        Args:
            image: The input image to correct. Can be:
                - A file path (Path or str)
                - A file-like object (IOBase)
                - A MIMEData object
                - A PIL Image object

        Returns:
            PIL.Image.Image: The orientation-corrected image as a PIL Image object

        Raises:
            ValueError: If the input is not a valid image
            UiformAPIError: If the API request fails
        """
        mime_document = prepare_mime_document(document)

        if not mime_document.mime_type.startswith("image/"):
            raise ValueError("Image is not a valid image")

        response = self._client._request("POST", "/api/v1/documents/correct_image_orientation", data={"document": mime_document.model_dump()})
        mime_response = MIMEData.model_validate(response['document'])
        return convert_mime_data_to_pil_image(mime_response)

    def create_messages(self, 
            document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
            modality: Modality = "native", 
            text_operations: dict[str, Any] | None = None,
            image_operations: dict[str, Any] | None = None,
            idempotency_key: str | None = None) -> DocumentMessage:
        """
        Create document messages from a file using the UiForm API.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
            modality: The processing modality to use. Defaults to "native".
            text_operations: Optional dictionary of text processing operations to apply.
            idempotency_key: Idempotency key for request
        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            UiformAPIError: If the API request fails.
        """
        loading_request = self._prepare_create_messages(document, modality, text_operations, image_operations)
        response = self._client._request("POST", "/api/v1/documents/create_messages", data=loading_request.model_dump(), idempotency_key=idempotency_key)
        return DocumentMessage.model_validate(response)



class AsyncDocuments(AsyncAPIResource, BaseDocumentsMixin):
    """Documents API wrapper for asynchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.extractions = AsyncExtractions(client=client)
        self.templates = AsyncTemplates(client=client)

    async def create_messages(self, 
            document: Path | str | IOBase | MIMEData | PIL.Image.Image, 
            modality: Modality = "native",
            text_operations: dict[str, Any] | None = None,
            image_operations: dict[str, Any] | None = None,
            idempotency_key: str | None = None) -> DocumentMessage:
        """
        Create document messages from a file using the UiForm API asynchronously.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
            modality: The processing modality to use. Defaults to "native".
            text_operations: Optional dictionary of text processing operations to apply.
            idempotency_key: Idempotency key for request
        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            UiformAPIError: If the API request fails.
        """
        loading_request = self._prepare_create_messages(document, modality, text_operations, image_operations)
        response = await self._client._request("POST", "/api/v1/documents/create_messages", data=loading_request.model_dump(), idempotency_key=idempotency_key)
        return DocumentMessage.model_validate(response)

    async def correct_image_orientation(self, document: Path | str | IOBase | MIMEData | PIL.Image.Image) -> PIL.Image.Image:
        """Corrects the orientation of an image using the UiForm API asynchronously.

        This method takes an image in various formats and returns a PIL Image with corrected orientation.
        Useful for handling images from mobile devices or cameras that may have incorrect EXIF orientation.

        Args:
            image: The input image to correct. Can be:
                - A file path (Path or str)
                - A file-like object (IOBase)
                - A MIMEData object
                - A PIL Image object

        Returns:
            PIL.Image.Image: The orientation-corrected image as a PIL Image object

        Raises:
            ValueError: If the input is not a valid image
            UiformAPIError: If the API request fails
        """
        mime_document = prepare_mime_document(document)

        if not mime_document.mime_type.startswith("image/"):
            raise ValueError("Image is not a valid image")

        response = await self._client._request("POST", "/api/v1/documents/correct_image_orientation", data={"document": mime_document.model_dump()})
        mime_response = MIMEData.model_validate(response['document'])
        return convert_mime_data_to_pil_image(mime_response)
