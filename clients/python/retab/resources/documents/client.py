from io import IOBase
from pathlib import Path
from typing import Any

import PIL.Image
from pydantic import HttpUrl
from pydantic_core import PydanticUndefined
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort

from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.json_schema import load_json_schema, filter_auxiliary_fields_json
from ..._utils.mime import convert_mime_data_to_pil_image, prepare_mime_document
from ..._utils.ai_models import assert_valid_model_extraction
from ...types.documents.create_messages import DocumentCreateInputRequest, DocumentCreateMessageRequest, DocumentMessage
from ...types.documents.extractions import DocumentExtractRequest, UiParsedChatCompletion
from ...types.browser_canvas import BrowserCanvas
from ...types.mime import MIMEData
from ...types.modalities import Modality
from ...types.schemas.object import Schema
from ...types.standards import PreparedRequest
from .extractions import AsyncExtractions, Extractions


def maybe_parse_to_pydantic(schema: Schema, response: UiParsedChatCompletion, allow_partial: bool = False) -> UiParsedChatCompletion:
    if response.choices[0].message.content:
        try:
            if allow_partial:
                response.choices[0].message.parsed = schema._partial_pydantic_model.model_validate(filter_auxiliary_fields_json(response.choices[0].message.content))
            else:
                response.choices[0].message.parsed = schema.pydantic_model.model_validate(filter_auxiliary_fields_json(response.choices[0].message.content))
        except Exception:
            pass
    return response


class BaseDocumentsMixin:
    def _prepare_create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        modality: Modality = "native",
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)

        loading_request = DocumentCreateMessageRequest(
            document=mime_document,
            modality=modality,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
        )
        return PreparedRequest(method="POST", url="/v1/documents/create_messages", data=loading_request.model_dump(), idempotency_key=idempotency_key)

    def _prepare_create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any] | Path | str,
        modality: Modality = "native",
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment],
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment],
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)
        loaded_schema = load_json_schema(json_schema)

        loading_request = DocumentCreateInputRequest(
            document=mime_document,
            modality=modality,
            json_schema=loaded_schema,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
        )
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
        #self.extractions_api = Extractions(client=client)
        # self.batch = Batch(client=client)

    def correct_image_orientation(self, document: Path | str | IOBase | MIMEData | PIL.Image.Image) -> PIL.Image.Image:
        """Corrects the orientation of an image using the Retab API.

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
            RetabAPIError: If the API request fails
        """
        request = self._prepare_correct_image_orientation(document)
        response = self._client._prepared_request(request)
        mime_response = MIMEData.model_validate(response["document"])
        return convert_mime_data_to_pil_image(mime_response)

    def create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        modality: Modality = "native",
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        idempotency_key: str | None = None,
    ) -> DocumentMessage:
        """
        Create document messages from a file using the Retab API.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
            modality: The processing modality to use. Defaults to "native".
            image_resolution_dpi: Optional image resolution DPI.
            browser_canvas: Optional browser canvas size.
            idempotency_key: Optional idempotency key for the request
        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            RetabAPIError: If the API request fails.
        """
        request = self._prepare_create_messages(
            document=document, modality=modality, image_resolution_dpi=image_resolution_dpi, browser_canvas=browser_canvas, idempotency_key=idempotency_key
        )
        response = self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    def create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any] | Path | str,
        modality: Modality = "native",
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        idempotency_key: str | None = None,
    ) -> DocumentMessage:
        """
        Create document inputs (messages with schema) from a file using the Retab API.

        Args:
            document: The document to process. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            json_schema: The JSON schema to use for structuring the document content.
            modality: The processing modality to use. Defaults to "native".
            image_resolution_dpi: Optional image resolution DPI.
            browser_canvas: Optional browser canvas size.
            idempotency_key: Optional idempotency key for the request
        Returns:
            DocumentMessage: The processed document message containing extracted content with schema context.

        Raises:
            RetabAPIError: If the API request fails.
        """
        request = self._prepare_create_inputs(
            document=document,
            json_schema=json_schema,
            modality=modality,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            idempotency_key=idempotency_key,
        )
        response = self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    def extract(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        temperature: float = PydanticUndefined,  # type: ignore[assignment]
        modality: Modality = PydanticUndefined,  # type: ignore[assignment]
        reasoning_effort: ChatCompletionReasoningEffort = PydanticUndefined,  # type: ignore[assignment]
        n_consensus: int = PydanticUndefined,  # type: ignore[assignment]
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> UiParsedChatCompletion:
        """
        Process one or more documents using the Retab API for structured data extraction.
        
        This method provides a direct interface to document extraction functionality,
        intended to replace the current `.extractions.parse()` pattern.

        Args:
            json_schema: JSON schema defining the expected data structure
            model: The AI model to use for processing
            document: Single document to process (use either this or documents, not both)
            documents: List of documents to process (use either this or document, not both)
            image_resolution_dpi: Optional image resolution DPI
            browser_canvas: Optional browser canvas size
            temperature: Model temperature setting (0-1)
            modality: Modality of the document (e.g., native)
            reasoning_effort: The effort level for the model to reason about the input data
            n_consensus: Number of consensus extractions to perform
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the Retab database
            
        Returns:
            UiParsedChatCompletion: Parsed response from the API
            
        Raises:
            ValueError: If neither document nor documents is provided, or if both are provided
            HTTPException: If the request fails
        """
        assert_valid_model_extraction(model)

        json_schema = load_json_schema(json_schema)

        # Handle both single document and multiple documents
        if document is not None and documents is not None:
            raise ValueError("Cannot provide both 'document' and 'documents' parameters. Use either one.")

        # Convert single document to documents list for consistency
        if document is not None:
            processed_documents = [prepare_mime_document(document)]
        elif documents is not None:
            processed_documents = [prepare_mime_document(doc) for doc in documents]
        else:
            raise ValueError("Must provide either 'document' or 'documents' parameter.")

        # Validate DocumentAPIRequest data (raises exception if invalid)
        request = DocumentExtractRequest(
            json_schema=json_schema,
            documents=processed_documents,
            model=model,
            temperature=temperature,
            stream=False,
            modality=modality,
            store=store,
            reasoning_effort=reasoning_effort,
            n_consensus=n_consensus,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
        )

        prepared_request = PreparedRequest(
            method="POST", 
            url="/v1/documents/extract", 
            data=request.model_dump(mode="json", exclude_unset=True, exclude_defaults=True), 
            idempotency_key=idempotency_key
        )
        
        response = self._client._prepared_request(prepared_request)

        schema = Schema(json_schema=load_json_schema(json_schema))
        return maybe_parse_to_pydantic(schema, UiParsedChatCompletion.model_validate(response))


class AsyncDocuments(AsyncAPIResource, BaseDocumentsMixin):
    """Documents API wrapper for asynchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        #self.extractions_api = AsyncExtractions(client=client)

    async def create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image,
        modality: Modality = "native",
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        idempotency_key: str | None = None,
    ) -> DocumentMessage:
        """
        Create document messages from a file using the Retab API asynchronously.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
            modality: The processing modality to use. Defaults to "native".
            idempotency_key: Idempotency key for request
        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            RetabAPIError: If the API request fails.
        """
        request = self._prepare_create_messages(
            document=document,
            modality=modality,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            idempotency_key=idempotency_key,
        )
        response = await self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    async def create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any] | Path | str,
        modality: Modality = "native",
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        idempotency_key: str | None = None,
    ) -> DocumentMessage:
        """
        Create document inputs (messages with schema) from a file using the Retab API asynchronously.

        Args:
            document: The document to process. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            json_schema: The JSON schema to use for structuring the document content.
            modality: The processing modality to use. Defaults to "native".
            image_resolution_dpi: Optional image resolution DPI.
            browser_canvas: Optional browser canvas size.
            idempotency_key: Idempotency key for request
        Returns:
            DocumentMessage: The processed document message containing extracted content with schema context.

        Raises:
            RetabAPIError: If the API request fails.
        """
        request = self._prepare_create_inputs(
            document=document,
            json_schema=json_schema,
            modality=modality,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            idempotency_key=idempotency_key,
        )
        response = await self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    async def correct_image_orientation(self, document: Path | str | IOBase | MIMEData | PIL.Image.Image) -> PIL.Image.Image:
        """Corrects the orientation of an image using the Retab API asynchronously.

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
            RetabAPIError: If the API request fails
        """
        request = self._prepare_correct_image_orientation(document)
        response = await self._client._prepared_request(request)
        mime_response = MIMEData.model_validate(response["document"])
        return convert_mime_data_to_pil_image(mime_response)

    async def extract(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        temperature: float = PydanticUndefined,  # type: ignore[assignment]
        modality: Modality = PydanticUndefined,  # type: ignore[assignment]
        reasoning_effort: ChatCompletionReasoningEffort = PydanticUndefined,  # type: ignore[assignment]
        n_consensus: int = PydanticUndefined,  # type: ignore[assignment]
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> UiParsedChatCompletion:
        """
        Process one or more documents using the Retab API for structured data extraction asynchronously.
        
        This method provides a direct interface to document extraction functionality,
        intended to replace the current `.extractions.parse()` pattern.

        Args:
            json_schema: JSON schema defining the expected data structure
            model: The AI model to use for processing
            document: Single document to process (use either this or documents, not both)
            documents: List of documents to process (use either this or document, not both)
            image_resolution_dpi: Optional image resolution DPI
            browser_canvas: Optional browser canvas size
            temperature: Model temperature setting (0-1)
            modality: Modality of the document (e.g., native)
            reasoning_effort: The effort level for the model to reason about the input data
            n_consensus: Number of consensus extractions to perform
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the Retab database
            
        Returns:
            UiParsedChatCompletion: Parsed response from the API
            
        Raises:
            ValueError: If neither document nor documents is provided, or if both are provided
            HTTPException: If the request fails
        """
        assert_valid_model_extraction(model)

        json_schema = load_json_schema(json_schema)

        # Handle both single document and multiple documents
        if document is not None and documents is not None:
            raise ValueError("Cannot provide both 'document' and 'documents' parameters. Use either one.")

        # Convert single document to documents list for consistency
        if document is not None:
            processed_documents = [prepare_mime_document(document)]
        elif documents is not None:
            processed_documents = [prepare_mime_document(doc) for doc in documents]
        else:
            raise ValueError("Must provide either 'document' or 'documents' parameter.")

        # Validate DocumentAPIRequest data (raises exception if invalid)
        request = DocumentExtractRequest(
            json_schema=json_schema,
            documents=processed_documents,
            model=model,
            temperature=temperature,
            stream=False,
            modality=modality,
            store=store,
            reasoning_effort=reasoning_effort,
            n_consensus=n_consensus,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
        )

        prepared_request = PreparedRequest(
            method="POST", 
            url="/v1/documents/extract", 
            data=request.model_dump(mode="json", exclude_unset=True, exclude_defaults=True), 
            idempotency_key=idempotency_key
        )
        
        response = await self._client._prepared_request(prepared_request)

        schema = Schema(json_schema=load_json_schema(json_schema))
        return maybe_parse_to_pydantic(schema, UiParsedChatCompletion.model_validate(response))
