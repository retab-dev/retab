import json
from io import IOBase
from pathlib import Path
from typing import Any, AsyncGenerator, Generator

import PIL.Image
from pydantic import HttpUrl
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from openai.types.chat.parsed_chat_completion import ParsedChatCompletionMessage

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...utils.mime import prepare_mime_document
from ...utils.stream_context_managers import as_async_context_manager, as_context_manager
from ...types.documents.create_messages import DocumentCreateInputRequest, DocumentCreateMessageRequest, DocumentMessage
from ...types.documents.extract import DocumentExtractRequest, RetabParsedChatCompletion, RetabParsedChatCompletionChunk, RetabParsedChoice, maybe_parse_to_pydantic
from ...types.documents.parse import ParseRequest, ParseResult, TableParsingFormat
from ...types.browser_canvas import BrowserCanvas
from ...types.mime import MIMEData
from ...types.standards import PreparedRequest, FieldUnset
from ...utils.json_schema import load_json_schema, unflatten_dict

class BaseDocumentsMixin:
    def _prepare_create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        idempotency_key: str | None = None,
        **extra_body: Any,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)

        loading_request_dict: dict[str, Any] = {
            "document": mime_document
        }
        if image_resolution_dpi is not FieldUnset:
            loading_request_dict["image_resolution_dpi"] = image_resolution_dpi
        if browser_canvas is not FieldUnset:
            loading_request_dict["browser_canvas"] = browser_canvas

        # Merge any extra fields provided by the caller
        if extra_body:
            loading_request_dict.update(extra_body)

        loading_request = DocumentCreateMessageRequest(**loading_request_dict)
        return PreparedRequest(
            method="POST", url="/v1/documents/create_messages", data=loading_request.model_dump(mode="json", exclude_unset=True), idempotency_key=idempotency_key
        )

    def _prepare_create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any] | Path | str,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        idempotency_key: str | None = None,
        **extra_body: Any,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)
        loaded_schema = load_json_schema(json_schema)

        loading_request_dict = {
            "document": mime_document,
            "json_schema": loaded_schema,
        }
        if image_resolution_dpi is not FieldUnset:
            loading_request_dict["image_resolution_dpi"] = image_resolution_dpi
        if browser_canvas is not FieldUnset:
            loading_request_dict["browser_canvas"] = browser_canvas

        # Merge any extra fields provided by the caller
        if extra_body:
            loading_request_dict.update(extra_body)

        loading_request = DocumentCreateInputRequest(**loading_request_dict)
        return PreparedRequest(method="POST", url="/v1/documents/create_inputs", data=loading_request.model_dump(mode="json", exclude_unset=True), idempotency_key=idempotency_key)

    def _prepare_correct_image_orientation(self, document: Path | str | IOBase | MIMEData | PIL.Image.Image) -> PreparedRequest:
        mime_document = prepare_mime_document(document)

        if not mime_document.mime_type.startswith("image/"):
            raise ValueError("Image is not a valid image")

        return PreparedRequest(
            method="POST",
            url="/v1/documents/correct_image_orientation",
            data={"document": mime_document.model_dump()},
        )

    def _prepare_parse(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str,
        table_parsing_format: TableParsingFormat = FieldUnset,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        idempotency_key: str | None = None,
        **extra_body: Any,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)

        request_dict = {
            "document": mime_document,
            "model": model,
        }
        if table_parsing_format is not FieldUnset:
            request_dict["table_parsing_format"] = table_parsing_format
        if image_resolution_dpi is not FieldUnset:
            request_dict["image_resolution_dpi"] = image_resolution_dpi
        if browser_canvas is not FieldUnset:
            request_dict["browser_canvas"] = browser_canvas

        # Merge any extra fields provided by the caller
        if extra_body:
            request_dict.update(extra_body)

        parse_request = ParseRequest(**request_dict)
        return PreparedRequest(method="POST", url="/v1/documents/parse", data=parse_request.model_dump(mode="json", exclude_unset=True), idempotency_key=idempotency_key)

    def _prepare_extract(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        temperature: float = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        n_consensus: int = FieldUnset,
        stream: bool = FieldUnset,
        store: bool = FieldUnset,
        idempotency_key: str | None = None,
        **extra_body: Any,
    ) -> PreparedRequest:

        loaded_schema = load_json_schema(json_schema)

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

        # Build request dictionary with only provided fields
        request_dict = {
            "json_schema": loaded_schema,
            "documents": processed_documents,
            "model": model,
        }
        if stream is not FieldUnset:
            request_dict["stream"] = stream
        if store is not FieldUnset:
            request_dict["store"] = store
        if temperature is not FieldUnset:
            request_dict["temperature"] = temperature
        if reasoning_effort is not FieldUnset:
            request_dict["reasoning_effort"] = reasoning_effort
        if n_consensus is not FieldUnset:
            request_dict["n_consensus"] = n_consensus
        if image_resolution_dpi is not FieldUnset:
            request_dict["image_resolution_dpi"] = image_resolution_dpi
        if browser_canvas is not FieldUnset:
            request_dict["browser_canvas"] = browser_canvas

        # Merge any extra fields provided by the caller
        if extra_body:
            request_dict.update(extra_body)

        # Validate DocumentAPIRequest data (raises exception if invalid)
        extract_request = DocumentExtractRequest(**request_dict)

        # Use the same URL as extractions.py for consistency when streaming
        url = "/v1/documents/extractions" if stream else "/v1/documents/extract"
        return PreparedRequest(
            method="POST", url=url, data=extract_request.model_dump(mode="json", exclude_unset=True, exclude_defaults=True), idempotency_key=idempotency_key
        )


class Documents(SyncAPIResource, BaseDocumentsMixin):
    """Documents API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    def create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        idempotency_key: str | None = None,
        **extra_body: Any,
    ) -> DocumentMessage:
        """
        Create document messages from a file using the Retab API.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
            image_resolution_dpi: Optional image resolution DPI.
            browser_canvas: Optional browser canvas size.
            idempotency_key: Optional idempotency key for the request
        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            RetabAPIError: If the API request fails.
        """
        request = self._prepare_create_messages(
            document=document,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            idempotency_key=idempotency_key,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    def create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any] | Path | str,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        idempotency_key: str | None = None,
        **extra_body: Any,
    ) -> DocumentMessage:
        """
        Create document inputs (messages with schema) from a file using the Retab API.

        Args:
            document: The document to process. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            json_schema: The JSON schema to use for structuring the document content.
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
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            idempotency_key=idempotency_key,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    def extract(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        temperature: float = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        n_consensus: int = FieldUnset,
        idempotency_key: str | None = None,
        store: bool = FieldUnset,
        **extra_body: Any,
    ) -> RetabParsedChatCompletion:
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
            reasoning_effort: The effort level for the model to reason about the input data
            n_consensus: Number of consensus extractions to perform
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the Retab database

        Returns:
            RetabParsedChatCompletion: Parsed response from the API

        Raises:
            ValueError: If neither document nor documents is provided, or if both are provided
            HTTPException: If the request fails
        """
        request = self._prepare_extract(
            json_schema=json_schema,
            model=model,
            document=document,
            documents=documents,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            n_consensus=n_consensus,
            store=store,
            idempotency_key=idempotency_key,
            **extra_body,
        )
        response = self._client._prepared_request(request)

        return maybe_parse_to_pydantic(load_json_schema(json_schema), RetabParsedChatCompletion.model_validate(response))

    @as_context_manager
    def extract_stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        temperature: float = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        n_consensus: int = FieldUnset,
        idempotency_key: str | None = None,
        store: bool = FieldUnset,
        **extra_body: Any,
    ) -> Generator[RetabParsedChatCompletion, None, None]:
        """
        Process one or more documents using the Retab API with streaming enabled.

        Args:
            json_schema: JSON schema defining the expected data structure
            model: The AI model to use for processing
            document: Single document to process (use either this or documents, not both)
            documents: List of documents to process (use either this or document, not both)
            image_resolution_dpi: Optional image resolution DPI.
            browser_canvas: Optional browser canvas size.
            temperature: Model temperature setting (0-1)
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the Retab database

        Returns:
            Generator[RetabParsedChatCompletion]: Stream of parsed responses
        Raises:
            ValueError: If neither document nor documents is provided, or if both are provided
            HTTPException: If the request fails
        Usage:
        ```python
        # Single document
        with retab.documents.extract_stream(json_schema, model, document=document) as stream:
            for response in stream:
                print(response)

        # Multiple documents
        with retab.documents.extract_stream(json_schema, model, documents=[doc1, doc2]) as stream:
            for response in stream:
                print(response)
        ```
        """
        request = self._prepare_extract(
            json_schema=json_schema,
            document=document,
            documents=documents,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            stream=True,
            n_consensus=n_consensus,
            store=store,
            idempotency_key=idempotency_key,
            **extra_body,
        )
        schema = load_json_schema(json_schema)

        # Request the stream and return a context manager
        ui_parsed_chat_completion_cum_chunk: RetabParsedChatCompletionChunk | None = None
        # Initialize the RetabParsedChatCompletion object
        ui_parsed_completion: RetabParsedChatCompletion = RetabParsedChatCompletion(
            id="",
            created=0,
            model="",
            object="chat.completion",
            likelihoods={},
            choices=[
                RetabParsedChoice(
                    index=0,
                    message=ParsedChatCompletionMessage(content="", role="assistant"),
                    finish_reason=None,
                    logprobs=None,
                )
            ],
        )
        for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            ui_parsed_chat_completion_cum_chunk = RetabParsedChatCompletionChunk.model_validate(chunk_json).chunk_accumulator(ui_parsed_chat_completion_cum_chunk)
            # Basic stuff
            ui_parsed_completion.id = ui_parsed_chat_completion_cum_chunk.id
            ui_parsed_completion.created = ui_parsed_chat_completion_cum_chunk.created
            ui_parsed_completion.model = ui_parsed_chat_completion_cum_chunk.model
            # Update the ui_parsed_completion object
            parsed = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_parsed)
            likelihoods = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_likelihoods)
            ui_parsed_completion.choices[0].message.content = json.dumps(parsed)
            ui_parsed_completion.choices[0].message.parsed = parsed
            ui_parsed_completion.likelihoods = likelihoods

            yield maybe_parse_to_pydantic(schema, ui_parsed_completion, allow_partial=True)

        # change the finish_reason to stop
        ui_parsed_completion.choices[0].finish_reason = "stop"
        yield maybe_parse_to_pydantic(schema, ui_parsed_completion)

    def parse(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str,
        table_parsing_format: TableParsingFormat = FieldUnset,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        idempotency_key: str | None = None,
        **extra_body: Any,
    ) -> ParseResult:
        """
        Parse a document and extract text content from each page.

        This method processes various document types and returns structured text content
        along with usage information. Supports different parsing modes and formats.

        Args:
            document: The document to parse. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            model: The AI model to use for document parsing.
            table_parsing_format: Format for parsing tables. Options: "html", "json", "yaml", "markdown". Defaults to "html".
            image_resolution_dpi: DPI for image processing. Defaults to 72.
            browser_canvas: Canvas size for document rendering. Defaults to "A4".
            idempotency_key: Optional idempotency key for the request.

        Returns:
            ParseResult: Parsed response containing document metadata, usage information, and page text content.

        Raises:
            HTTPException: If the request fails.
        """
        request = self._prepare_parse(
            document=document,
            model=model,
            table_parsing_format=table_parsing_format,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            idempotency_key=idempotency_key,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return ParseResult.model_validate(response)


class AsyncDocuments(AsyncAPIResource, BaseDocumentsMixin):
    """Documents API wrapper for asynchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        # self.extractions_api = AsyncExtractions(client=client)

    async def create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        idempotency_key: str | None = None,
        **extra_body: Any,
    ) -> DocumentMessage:
        """
        Create document messages from a file using the Retab API asynchronously.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
            idempotency_key: Idempotency key for request
        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            RetabAPIError: If the API request fails.
        """
        request = self._prepare_create_messages(
            document=document,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            idempotency_key=idempotency_key,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    async def create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any] | Path | str,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        idempotency_key: str | None = None,
        **extra_body: Any,
    ) -> DocumentMessage:
        """
        Create document inputs (messages with schema) from a file using the Retab API asynchronously.

        Args:
            document: The document to process. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            json_schema: The JSON schema to use for structuring the document content.
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
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            idempotency_key=idempotency_key,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    async def extract(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        temperature: float = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        n_consensus: int = FieldUnset,
        idempotency_key: str | None = None,
        store: bool = FieldUnset,
        **extra_body: Any,
    ) -> RetabParsedChatCompletion:
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
            reasoning_effort: The effort level for the model to reason about the input data
            n_consensus: Number of consensus extractions to perform
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the Retab database

        Returns:
            RetabParsedChatCompletion: Parsed response from the API

        Raises:
            ValueError: If neither document nor documents is provided, or if both are provided
            HTTPException: If the request fails
        """
        request = self._prepare_extract(
            json_schema=json_schema,
            model=model,
            document=document,
            documents=documents,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            n_consensus=n_consensus,
            store=store,
            idempotency_key=idempotency_key,
            **extra_body,
        )
        response = await self._client._prepared_request(request)

        return maybe_parse_to_pydantic(load_json_schema(json_schema), RetabParsedChatCompletion.model_validate(response))

    @as_async_context_manager
    async def extract_stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        temperature: float = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        n_consensus: int = FieldUnset,
        idempotency_key: str | None = None,
        store: bool = FieldUnset,
        **extra_body: Any,
    ) -> AsyncGenerator[RetabParsedChatCompletion, None]:
        """
        Extract structured data from one or more documents asynchronously with streaming.

        Args:
            json_schema: JSON schema defining the expected data structure.
            model: The AI model to use.
            document: Single document to process (use either this or documents, not both)
            documents: List of documents to process (use either this or document, not both)
            image_resolution_dpi: Optional image resolution DPI.
            browser_canvas: Optional browser canvas size.
            temperature: Model temperature setting (0-1).
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the Retab database
        Returns:
            AsyncGenerator[RetabParsedChatCompletion, None]: Stream of parsed responses.
        Raises:
            ValueError: If neither document nor documents is provided, or if both are provided

        Usage:
        ```python
        # Single document
        async with retab.documents.extract_stream(json_schema, model, document=document) as stream:
            async for response in stream:
                print(response)

        # Multiple documents
        async with retab.documents.extract_stream(json_schema, model, documents=[doc1, doc2]) as stream:
            async for response in stream:
                print(response)
        ```
        """
        request = self._prepare_extract(
            json_schema=json_schema,
            document=document,
            documents=documents,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            stream=True,
            n_consensus=n_consensus,
            store=store,
            idempotency_key=idempotency_key,
            **extra_body,
        )
        schema = load_json_schema(json_schema)
        ui_parsed_chat_completion_cum_chunk: RetabParsedChatCompletionChunk | None = None
        # Initialize the RetabParsedChatCompletion object
        ui_parsed_completion: RetabParsedChatCompletion = RetabParsedChatCompletion(
            id="",
            created=0,
            model="",
            object="chat.completion",
            likelihoods={},
            choices=[
                RetabParsedChoice(
                    index=0,
                    message=ParsedChatCompletionMessage(content="", role="assistant"),
                    finish_reason=None,
                    logprobs=None,
                )
            ],
        )

        async for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            ui_parsed_chat_completion_cum_chunk = RetabParsedChatCompletionChunk.model_validate(chunk_json).chunk_accumulator(ui_parsed_chat_completion_cum_chunk)
            # Basic stuff
            ui_parsed_completion.id = ui_parsed_chat_completion_cum_chunk.id
            ui_parsed_completion.created = ui_parsed_chat_completion_cum_chunk.created
            ui_parsed_completion.model = ui_parsed_chat_completion_cum_chunk.model

            # Update the ui_parsed_completion object
            parsed = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_parsed)
            likelihoods = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_likelihoods)
            ui_parsed_completion.choices[0].message.content = json.dumps(parsed)
            ui_parsed_completion.choices[0].message.parsed = parsed
            ui_parsed_completion.likelihoods = likelihoods

            yield maybe_parse_to_pydantic(schema, ui_parsed_completion, allow_partial=True)

        # change the finish_reason to stop
        ui_parsed_completion.choices[0].finish_reason = "stop"
        yield maybe_parse_to_pydantic(schema, ui_parsed_completion)

    async def parse(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str,
        table_parsing_format: TableParsingFormat = FieldUnset,
        image_resolution_dpi: int = FieldUnset,
        browser_canvas: BrowserCanvas = FieldUnset,
        idempotency_key: str | None = None,
        **extra_body: Any,
    ) -> ParseResult:
        """
        Parse a document and extract text content from each page asynchronously.

        This method processes various document types and returns structured text content
        along with usage information. Supports different parsing modes and formats.

        Args:
            document: The document to parse. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            model: The AI model to use for document parsing.
            table_parsing_format: Format for parsing tables. Options: "html", "json", "yaml", "markdown". Defaults to "html".
            image_resolution_dpi: DPI for image processing. Defaults to 96.
            browser_canvas: Canvas size for document rendering. Defaults to "A4".
            idempotency_key: Optional idempotency key for the request.

        Returns:
            ParseResult: Parsed response containing document metadata, usage information, and page text content.

        Raises:
            HTTPException: If the request fails.
        """
        request = self._prepare_parse(
            document=document,
            model=model,
            table_parsing_format=table_parsing_format,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            idempotency_key=idempotency_key,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return ParseResult.model_validate(response)
