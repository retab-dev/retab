from io import IOBase
from pathlib import Path
from typing import Any

import PIL.Image
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.mime import convert_mime_data_to_pil_image, prepare_mime_document
from ...types.documents.create_messages import DocumentCreateMessageRequest, DocumentMessage, DocumentCreateInputRequest
from ...types.mime import MIMEData
from ...types.modalities import Modality
from ...types.standards import PreparedRequest
from .extractions import AsyncExtractions, Extractions


class BaseDocumentsMixin:
    def _prepare_create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        modality: Modality = "native",
        image_settings: dict[str, Any] | None = None,
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)
        data: dict[str, Any] = {
            "document": mime_document.model_dump(),
            "modality": modality,
        }
        if image_settings:
            data["image_settings"] = image_settings

        loading_request = DocumentCreateMessageRequest.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/documents/create_messages", data=loading_request.model_dump(), idempotency_key=idempotency_key)

    def _prepare_create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any],
        modality: Modality = "native",
        image_settings: dict[str, Any] | None = None,
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)
        data: dict[str, Any] = {
            "document": mime_document.model_dump(),
            "modality": modality,
            "json_schema": json_schema,
        }
        if image_settings:
            data["image_settings"] = image_settings

        loading_request = DocumentCreateInputRequest.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/documents/create_inputs", data=loading_request.model_dump(), idempotency_key=idempotency_key)

    def _prepare_correct_image_orientation(self, document: Path | str | IOBase | MIMEData | PIL.Image.Image) -> PreparedRequest:
        mime_document = prepare_mime_document(document)

        if not mime_document.mime_type.startswith("image/"):
            raise ValueError("Image is not a valid image")

        return PreparedRequest(
            method="POST",
            url="/v1/documents/correct_image_orientation",
            data={"document": mime_document.model_dump()},
        )


class Documents(SyncAPIResource, BaseDocumentsMixin):
    """Documents API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.extractions = Extractions(client=client)
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
        request = self._prepare_correct_image_orientation(document)
        response = self._client._prepared_request(request)
        mime_response = MIMEData.model_validate(response['document'])
        return convert_mime_data_to_pil_image(mime_response)

    def create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        modality: Modality = "native",
        image_settings: dict[str, Any] | None = None,
        idempotency_key: str | None = None,
    ) -> DocumentMessage:
        """
        Create document messages from a file using the UiForm API.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
            modality: The processing modality to use. Defaults to "native".
            image_settings: Optional dictionary of image processing operations to apply.
                It has to be a dictionary with the following keys:
                {
                    "correct_image_orientation": True,  # Whether to auto-correct image orientation
                    "dpi": 72,                         # DPI for image processing
                    "image_to_text": "ocr",            # OCR engine to use
                    "browser_canvas": "A4"             # Output page size
                }
            idempotency_key: Optional idempotency key for the request
        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            UiformAPIError: If the API request fails.
        """
        request = self._prepare_create_messages(document, modality, image_settings, idempotency_key)
        response = self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    def create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any],
        modality: Modality = "native",
        image_settings: dict[str, Any] | None = None,
        idempotency_key: str | None = None,
    ) -> DocumentMessage:
        """
        Create document inputs (messages with schema) from a file using the UiForm API.

        Args:
            document: The document to process. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            json_schema: The JSON schema to use for structuring the document content.
            modality: The processing modality to use. Defaults to "native".
            image_settings: Optional dictionary of image processing operations to apply.
                It has to be a dictionary with the following keys:
                {
                    "correct_image_orientation": True,  # Whether to auto-correct image orientation
                    "dpi": 72,                         # DPI for image processing
                    "image_to_text": "ocr",            # OCR engine to use
                    "browser_canvas": "A4"             # Output page size
                }
            idempotency_key: Optional idempotency key for the request
        Returns:
            DocumentMessage: The processed document message containing extracted content with schema context.

        Raises:
            UiformAPIError: If the API request fails.
        """
        request = self._prepare_create_inputs(document, json_schema, modality, image_settings, idempotency_key)
        response = self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)


class AsyncDocuments(AsyncAPIResource, BaseDocumentsMixin):
    """Documents API wrapper for asynchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.extractions = AsyncExtractions(client=client)

    async def create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image,
        modality: Modality = "native",
        image_settings: dict[str, Any] | None = None,
        idempotency_key: str | None = None,
    ) -> DocumentMessage:
        """
        Create document messages from a file using the UiForm API asynchronously.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
            modality: The processing modality to use. Defaults to "native".
            idempotency_key: Idempotency key for request
        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            UiformAPIError: If the API request fails.
        """
        request = self._prepare_create_messages(document, modality, image_settings, idempotency_key)
        assert request.data is not None
        print(request.data.keys())
        print(request.data)
        response = await self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    async def create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any],
        modality: Modality = "native",
        image_settings: dict[str, Any] | None = None,
        idempotency_key: str | None = None,
    ) -> DocumentMessage:
        """
        Create document inputs (messages with schema) from a file using the UiForm API asynchronously.

        Args:
            document: The document to process. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            json_schema: The JSON schema to use for structuring the document content.
            modality: The processing modality to use. Defaults to "native".
            image_settings: Optional dictionary of image processing operations to apply.
            idempotency_key: Idempotency key for request
        Returns:
            DocumentMessage: The processed document message containing extracted content with schema context.

        Raises:
            UiformAPIError: If the API request fails.
        """
        request = self._prepare_create_inputs(document, json_schema, modality, image_settings, idempotency_key)
        response = await self._client._prepared_request(request)
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
        request = self._prepare_correct_image_orientation(document)
        response = await self._client._prepared_request(request)
        mime_response = MIMEData.model_validate(response['document'])
        return convert_mime_data_to_pil_image(mime_response)
