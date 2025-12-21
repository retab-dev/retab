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
from ...types.chat import ChatCompletionRetabMessage
from ...types.documents.edit import EditRequest, EditResponse
from ...types.documents.extract import DocumentExtractRequest, RetabParsedChatCompletion, RetabParsedChatCompletionChunk, RetabParsedChoice, maybe_parse_to_pydantic
from ...types.documents.parse import ParseRequest, ParseResult, TableParsingFormat
from ...types.documents.split import Category, SplitRequest, SplitResponse
from ...types.mime import MIMEData
from ...types.standards import PreparedRequest, FieldUnset
from ...utils.json_schema import load_json_schema, unflatten_dict
from ...types.mime import OCR


class BaseDocumentsMixin:
    def _prepare_create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        image_resolution_dpi: int = FieldUnset,
        **extra_body: Any,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)

        loading_request_dict: dict[str, Any] = {"document": mime_document}
        if image_resolution_dpi is not FieldUnset:
            loading_request_dict["image_resolution_dpi"] = image_resolution_dpi

        # Merge any extra fields provided by the caller
        if extra_body:
            loading_request_dict.update(extra_body)

        loading_request = DocumentCreateMessageRequest(**loading_request_dict)
        return PreparedRequest(
            method="POST", url="/v1/documents/create_messages", data=loading_request.model_dump(mode="json", exclude_unset=True)
        )

    def _prepare_get_extraction(self, extraction_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/extractions/{extraction_id}")

    def _prepare_perform_ocr_only(self, file_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url="/v1/documents/perform_ocr_only", data={"file_id": file_id})

    def _prepare_compute_field_locations(self, ocr_file_id: str, ocr_result: OCR | None = None, data: dict[str, Any] = {}) -> PreparedRequest:
        return PreparedRequest(method="POST", url="/v1/documents/compute_field_locations", data={"ocr_file_id": ocr_file_id, "ocr_result": ocr_result, "data": data})

    def _prepare_create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any] | Path | str,
        image_resolution_dpi: int = FieldUnset,
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

        # Merge any extra fields provided by the caller
        if extra_body:
            loading_request_dict.update(extra_body)

        loading_request = DocumentCreateInputRequest(**loading_request_dict)
        return PreparedRequest(method="POST", url="/v1/documents/create_inputs", data=loading_request.model_dump(mode="json", exclude_unset=True))

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

        # Merge any extra fields provided by the caller
        if extra_body:
            request_dict.update(extra_body)

        parse_request = ParseRequest(**request_dict)
        return PreparedRequest(method="POST", url="/v1/documents/parse", data=parse_request.model_dump(mode="json", exclude_unset=True))

    def _prepare_edit(
        self,
        filling_instructions: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str = FieldUnset,
        template_id: str | None = FieldUnset,
        **extra_body: Any,
    ) -> PreparedRequest:
        request_dict: dict[str, Any] = {
            "filling_instructions": filling_instructions,
        }
        
        if document is not None:
            mime_document = prepare_mime_document(document)
            request_dict["document"] = mime_document
        
        if model is not FieldUnset:
            request_dict["model"] = model
        if template_id is not FieldUnset:
            request_dict["template_id"] = template_id

        # Merge any extra fields provided by the caller
        if extra_body:
            request_dict.update(extra_body)

        edit_request = EditRequest(**request_dict)
        return PreparedRequest(method="POST", url="/v1/documents/edit", data=edit_request.model_dump(mode="json", exclude_unset=True))

    def _prepare_split(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        categories: list[Category] | list[dict[str, str]],
        model: str,
        **extra_body: Any,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)
        
        # Convert dict categories to Category objects if needed
        category_objects = [
            Category(**cat) if isinstance(cat, dict) else cat
            for cat in categories
        ]
        
        request_dict: dict[str, Any] = {
            "document": mime_document,
            "categories": category_objects,
            "model": model,
        }

        # Merge any extra fields provided by the caller
        if extra_body:
            request_dict.update(extra_body)

        split_request = SplitRequest(**request_dict)
        return PreparedRequest(method="POST", url="/v1/documents/split", data=split_request.model_dump(mode="json", exclude_unset=True))

    def _prepare_extract(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | MIMEData,
        image_resolution_dpi: int = FieldUnset,
        temperature: float = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        n_consensus: int = FieldUnset,
        stream: bool = FieldUnset,
        store: bool = FieldUnset,
        metadata: dict[str, str] = FieldUnset,
        additional_messages: list[ChatCompletionRetabMessage] = FieldUnset,
        **extra_body: Any,
    ) -> PreparedRequest:
        loaded_schema = load_json_schema(json_schema)

        # Handle document parameter
        if document is None:
            raise ValueError("Must provide 'document' parameter.")

        processed_document = prepare_mime_document(document)

        # Build request dictionary with only provided fields
        request_dict: dict[str, Any] = {
            "json_schema": loaded_schema,
            "document": processed_document,
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
        if metadata is not FieldUnset:
            request_dict["metadata"] = metadata
        if additional_messages is not FieldUnset:
            request_dict["additional_messages"] = additional_messages

        # Merge any extra fields provided by the caller
        if extra_body:
            request_dict.update(extra_body)

        # Validate DocumentAPIRequest data (raises exception if invalid)
        extract_request = DocumentExtractRequest(**request_dict)

        # Use the same URL as extractions.py for consistency when streaming
        url = "/v1/documents/extractions" if stream else "/v1/documents/extract"
        return PreparedRequest(method="POST", url=url, data=extract_request.model_dump(mode="json", exclude_unset=True, exclude_defaults=True))


class Documents(SyncAPIResource, BaseDocumentsMixin):
    """Documents API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    def create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        image_resolution_dpi: int = FieldUnset,
        **extra_body: Any,
    ) -> DocumentMessage:
        """
        Create document messages from a file using the Retab API.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
            image_resolution_dpi: Optional image resolution DPI.
        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            RetabAPIError: If the API request fails.
        """
        request = self._prepare_create_messages(
            document=document,
            image_resolution_dpi=image_resolution_dpi,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    def create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any] | Path | str,
        image_resolution_dpi: int = FieldUnset,
        **extra_body: Any,
    ) -> DocumentMessage:
        """
        Create document inputs (messages with schema) from a file using the Retab API.

        Args:
            document: The document to process. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            json_schema: The JSON schema to use for structuring the document content.
            image_resolution_dpi: Optional image resolution DPI.
        Returns:
            DocumentMessage: The processed document message containing extracted content with schema context.

        Raises:
            RetabAPIError: If the API request fails.
        """
        request = self._prepare_create_inputs(
            document=document,
            json_schema=json_schema,
            image_resolution_dpi=image_resolution_dpi,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    def extract(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | MIMEData,
        image_resolution_dpi: int = FieldUnset,
        temperature: float = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        n_consensus: int = FieldUnset,
        store: bool = FieldUnset,
        metadata: dict[str, str] = FieldUnset,
        additional_messages: list[ChatCompletionRetabMessage] = FieldUnset,
        **extra_body: Any,
    ) -> RetabParsedChatCompletion:
        """
        Process a document using the Retab API for structured data extraction.

        This method provides a direct interface to document extraction functionality.

        Args:
            json_schema: JSON schema defining the expected data structure
            model: The AI model to use for processing
            document: Document to process (file path, URL, file-like object, or MIMEData)
            image_resolution_dpi: Optional image resolution DPI
            temperature: Model temperature setting (0-1)
            reasoning_effort: The effort level for the model to reason about the input data
            n_consensus: Number of consensus extractions to perform
            store: Whether to store the document in the Retab database
            metadata: User-defined metadata to associate with this extraction
            additional_messages: Additional chat messages to append after the document content messages

        Returns:
            RetabParsedChatCompletion: Parsed response from the API

        Raises:
            ValueError: If document is not provided
            HTTPException: If the request fails
        """
        request = self._prepare_extract(
            json_schema=json_schema,
            model=model,
            document=document,
            image_resolution_dpi=image_resolution_dpi,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            n_consensus=n_consensus,
            store=store,
            metadata=metadata,
            additional_messages=additional_messages,
            **extra_body,
        )
        response = self._client._prepared_request(request)

        return maybe_parse_to_pydantic(load_json_schema(json_schema), RetabParsedChatCompletion.model_validate(response))
        # returns a RetabParsedChatCompletion

    def get_sources(
        self,
        extraction_id: str,
        file: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        file_id: str | None = None,
    ) -> dict[str, Any]:
        # 1) Get extraction → pick structured data
        request = self._prepare_get_extraction(extraction_id)
        extraction_resp = self._client._prepared_request(request)

        # Access completion safely whether dict-like or model-like
        if isinstance(extraction_resp, dict):
            completion = extraction_resp.get("completion", {})
            choices = completion.get("choices", []) or []
            message = (choices[0].get("message") if choices else {}) or {}
            parsed = message.get("parsed")
            if parsed is None:
                content = message.get("content") or "{}"
                try:
                    data: dict[str, Any] = json.loads(content)
                except Exception:
                    data = {}
            else:
                data = parsed
        else:
            # Fallback for attribute-style objects
            completion = getattr(extraction_resp, "completion", None)
            choices = getattr(completion, "choices", []) if completion is not None else []
            message = choices[0].message if choices else None  # type: ignore[attr-defined]
            parsed = getattr(message, "parsed", None) if message is not None else None
            if parsed is None and message is not None:
                try:
                    data = json.loads(getattr(message, "content", "{}"))
                except Exception:
                    data = {}
            else:
                data = parsed or {}

        # 2) Determine file_id from provided file or from extraction metadata
        if not file_id:
            if file is not None:
                try:
                    mime = prepare_mime_document(file)
                    file_id = mime.id
                except Exception:
                    file_id = None
        if not file_id:
            candidate_file_ids: list[str] = []
            if isinstance(extraction_resp, dict):
                candidate_file_ids = extraction_resp.get("file_ids") or []
                if not candidate_file_ids:
                    legacy = extraction_resp.get("file_id")
                    if legacy:
                        candidate_file_ids = [legacy]
            else:
                candidate_file_ids = getattr(extraction_resp, "file_ids", []) or []
                if not candidate_file_ids:
                    legacy = getattr(extraction_resp, "file_id", None)
                    if legacy:
                        candidate_file_ids = [legacy]
            if len(candidate_file_ids) == 1:
                file_id = candidate_file_ids[0]

        if not isinstance(file_id, str) or not file_id:
            raise ValueError("Unable to infer file_id. Provide a file path or explicit file_id, or use an extraction with a single file.")

        # 3) OCR the file_id
        request = self._prepare_perform_ocr_only(file_id)
        ocr_resp = self._client._prepared_request(request)

        # Support dict or model-like
        if isinstance(ocr_resp, dict):
            ocr_file_id = ocr_resp.get("ocr_file_id")
            ocr_result = ocr_resp.get("ocr_result")
        else:
            ocr_file_id = getattr(ocr_resp, "ocr_file_id", None)
            ocr_result = getattr(ocr_resp, "ocr_result", None)

        # 4) Compute field locations (pass both ocr_file_id and ocr_result to avoid refetch)
        if not isinstance(ocr_file_id, str) or not ocr_file_id:
            raise ValueError("perform_ocr_only did not return a valid ocr_file_id")
        request = self._prepare_compute_field_locations(ocr_file_id, ocr_result, data)
        field_locations_resp = self._client._prepared_request(request)

        # Caller will decide how to format the return
        return field_locations_resp

    @as_context_manager
    def extract_stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | MIMEData,
        image_resolution_dpi: int = FieldUnset,
        temperature: float = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        n_consensus: int = FieldUnset,
        store: bool = FieldUnset,
        metadata: dict[str, str] = FieldUnset,
        additional_messages: list[ChatCompletionRetabMessage] = FieldUnset,
        **extra_body: Any,
    ) -> Generator[RetabParsedChatCompletion, None, None]:
        """
        Process a document using the Retab API with streaming enabled.

        Args:
            json_schema: JSON schema defining the expected data structure
            model: The AI model to use for processing
            document: Document to process (file path, URL, or file-like object)
            image_resolution_dpi: Optional image resolution DPI.
            temperature: Model temperature setting (0-1)
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            store: Whether to store the document in the Retab database
            metadata: User-defined metadata to associate with this extraction
            additional_messages: Additional chat messages to append after the document content messages

        Returns:
            Generator[RetabParsedChatCompletion]: Stream of parsed responses
        Raises:
            ValueError: If document is not provided
            HTTPException: If the request fails
        Usage:
        ```python
        with retab.documents.extract_stream(json_schema, model, document=document) as stream:
            for response in stream:
                print(response)
        ```
        """
        request = self._prepare_extract(
            json_schema=json_schema,
            document=document,
            image_resolution_dpi=image_resolution_dpi,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            stream=True,
            n_consensus=n_consensus,
            store=store,
            metadata=metadata,
            additional_messages=additional_messages,
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
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return ParseResult.model_validate(response)

    def edit(
        self,
        filling_instructions: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str = FieldUnset,
        template_id: str | None = FieldUnset,
        **extra_body: Any,
    ) -> EditResponse:
        """
        Edit a PDF document by inferring form fields and filling them with provided instructions.

        This method performs:
        1. OCR on the PDF to extract text elements with bounding boxes
        2. LLM inference to identify fillable form fields
        3. LLM-based form filling using the provided instructions
        4. Returns the filled PDF with form field values populated

        Either `document` OR `template_id` must be provided, but not both.

        Args:
            filling_instructions: Instructions describing how to fill the form fields.
            document: The document to edit. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
                Mutually exclusive with template_id.
            model: The LLM model to use for inference. Defaults to "retab-small".
            template_id: Template ID to use for filling. When provided, uses the template's pre-defined form fields
                and empty PDF. Only works for PDF documents. Mutually exclusive with document.

        Returns:
            EditResponse: Response containing:
                - form_data: List of form fields with filled values
                - filled_document: PDF with filled form values (MIMEData)

        Raises:
            HTTPException: If the request fails.
        """
        request = self._prepare_edit(
            filling_instructions=filling_instructions,
            document=document,
            model=model,
            template_id=template_id,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return EditResponse.model_validate(response)

    def split(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        categories: list[Category] | list[dict[str, str]],
        model: str,
        **extra_body: Any,
    ) -> SplitResponse:
        """
        Split a document into sections based on provided categories.

        This method analyzes a multi-page document and classifies pages into 
        user-defined categories, returning the page ranges for each section.

        Args:
            document: The document to split. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            categories: List of categories to split the document into. Each category should have a 'name' and 'description'.
                Can be Category objects or dicts with 'name' and 'description' keys.
            model: The AI model to use for document splitting (e.g., "gemini-2.5-flash").

        Returns:
            SplitResponse: Response containing:
                - splits: List of SplitResult objects with name, start_page, and end_page for each section.

        Raises:
            HTTPException: If the request fails.

        Example:
            ```python
            response = retab.documents.split(
                document="invoice_batch.pdf",
                model="gemini-2.5-flash",
                categories=[
                    {"name": "invoice", "description": "Invoice documents with billing information"},
                    {"name": "receipt", "description": "Receipt documents for payments"},
                    {"name": "contract", "description": "Legal contract documents"},
                ]
            )
            for split in response.splits:
                print(f"{split.name}: pages {split.start_page}-{split.end_page}")
            ```
        """
        request = self._prepare_split(
            document=document,
            categories=categories,
            model=model,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return SplitResponse.model_validate(response)


class AsyncDocuments(AsyncAPIResource, BaseDocumentsMixin):
    """Documents API wrapper for asynchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        # self.extractions_api = AsyncExtractions(client=client)

    async def create_messages(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image,
        image_resolution_dpi: int = FieldUnset,
        **extra_body: Any,
    ) -> DocumentMessage:
        """
        Create document messages from a file using the Retab API asynchronously.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            RetabAPIError: If the API request fails.
        """
        request = self._prepare_create_messages(
            document=document,
            image_resolution_dpi=image_resolution_dpi,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    async def create_inputs(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        json_schema: dict[str, Any] | Path | str,
        image_resolution_dpi: int = FieldUnset,
        **extra_body: Any,
    ) -> DocumentMessage:
        """
        Create document inputs (messages with schema) from a file using the Retab API asynchronously.

        Args:
            document: The document to process. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            json_schema: The JSON schema to use for structuring the document content.
            image_resolution_dpi: Optional image resolution DPI.
        Returns:
            DocumentMessage: The processed document message containing extracted content with schema context.

        Raises:
            RetabAPIError: If the API request fails.
        """
        request = self._prepare_create_inputs(
            document=document,
            json_schema=json_schema,
            image_resolution_dpi=image_resolution_dpi,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return DocumentMessage.model_validate(response)

    async def extract(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | MIMEData,
        image_resolution_dpi: int = FieldUnset,
        temperature: float = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        n_consensus: int = FieldUnset,
        store: bool = FieldUnset,
        metadata: dict[str, str] = FieldUnset,
        additional_messages: list[ChatCompletionRetabMessage] = FieldUnset,
        **extra_body: Any,
    ) -> RetabParsedChatCompletion:
        """
        Process a document using the Retab API for structured data extraction asynchronously.

        This method provides a direct interface to document extraction functionality.

        Args:
            json_schema: JSON schema defining the expected data structure
            model: The AI model to use for processing
            document: Document to process (file path, URL, file-like object, or MIMEData)
            image_resolution_dpi: Optional image resolution DPI
            temperature: Model temperature setting (0-1)
            reasoning_effort: The effort level for the model to reason about the input data
            n_consensus: Number of consensus extractions to perform
            store: Whether to store the document in the Retab database
            metadata: User-defined metadata to associate with this extraction
            additional_messages: Additional chat messages to append after the document content messages

        Returns:
            RetabParsedChatCompletion: Parsed response from the API

        Raises:
            ValueError: If document is not provided
            HTTPException: If the request fails
        """
        request = self._prepare_extract(
            json_schema=json_schema,
            model=model,
            document=document,
            image_resolution_dpi=image_resolution_dpi,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            n_consensus=n_consensus,
            store=store,
            metadata=metadata,
            additional_messages=additional_messages,
            **extra_body,
        )
        response = await self._client._prepared_request(request)

        return maybe_parse_to_pydantic(load_json_schema(json_schema), RetabParsedChatCompletion.model_validate(response))

    @as_async_context_manager
    async def extract_stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | MIMEData,
        image_resolution_dpi: int = FieldUnset,
        temperature: float = FieldUnset,
        reasoning_effort: ChatCompletionReasoningEffort = FieldUnset,
        n_consensus: int = FieldUnset,
        store: bool = FieldUnset,
        metadata: dict[str, str] = FieldUnset,
        additional_messages: list[ChatCompletionRetabMessage] = FieldUnset,
        **extra_body: Any,
    ) -> AsyncGenerator[RetabParsedChatCompletion, None]:
        """
        Extract structured data from a document asynchronously with streaming.

        Args:
            json_schema: JSON schema defining the expected data structure.
            model: The AI model to use.
            document: Document to process (file path, URL, file-like object, or MIMEData)
            image_resolution_dpi: Optional image resolution DPI.
            temperature: Model temperature setting (0-1).
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            store: Whether to store the document in the Retab database
            metadata: User-defined metadata to associate with this extraction
            additional_messages: Additional chat messages to append after the document content messages
        Returns:
            AsyncGenerator[RetabParsedChatCompletion, None]: Stream of parsed responses.
        Raises:
            ValueError: If document is not provided

        Usage:
        ```python
        async with retab.documents.extract_stream(json_schema, model, document=document) as stream:
            async for response in stream:
                print(response)
        ```
        """
        request = self._prepare_extract(
            json_schema=json_schema,
            document=document,
            image_resolution_dpi=image_resolution_dpi,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            stream=True,
            n_consensus=n_consensus,
            store=store,
            metadata=metadata,
            additional_messages=additional_messages,
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
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return ParseResult.model_validate(response)

    async def edit(
        self,
        filling_instructions: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str = FieldUnset,
        template_id: str | None = FieldUnset,
        **extra_body: Any,
    ) -> EditResponse:
        """
        Edit a PDF document by inferring form fields and filling them asynchronously.

        This method performs:
        1. OCR on the PDF to extract text elements with bounding boxes
        2. LLM inference to identify fillable form fields
        3. LLM-based form filling using the provided instructions
        4. Returns the filled PDF with form field values populated

        Either `document` OR `template_id` must be provided, but not both.

        Args:
            filling_instructions: Instructions describing how to fill the form fields.
            document: The document to edit. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
                Mutually exclusive with template_id.
            model: The LLM model to use for inference. Defaults to "gemini-2.5-pro".
            template_id: Template ID to use for filling. When provided, uses the template's pre-defined form fields
                and empty PDF. Only works for PDF documents. Mutually exclusive with document.

        Returns:
            EditResponse: Response containing:
                - form_data: List of form fields with filled values
                - filled_document: PDF with filled form values (MIMEData)

        Raises:
            HTTPException: If the request fails.
        """
        request = self._prepare_edit(
            filling_instructions=filling_instructions,
            document=document,
            model=model,
            template_id=template_id,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return EditResponse.model_validate(response)

    async def split(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        categories: list[Category] | list[dict[str, str]],
        model: str,
        **extra_body: Any,
    ) -> SplitResponse:
        """
        Split a document into sections based on provided categories asynchronously.

        This method analyzes a multi-page document and classifies pages into 
        user-defined categories, returning the page ranges for each section.

        Args:
            document: The document to split. Can be a file path (Path or str), file-like object, MIMEData, PIL Image, or URL.
            categories: List of categories to split the document into. Each category should have a 'name' and 'description'.
                Can be Category objects or dicts with 'name' and 'description' keys.
            model: The AI model to use for document splitting (e.g., "gemini-2.5-flash").

        Returns:
            SplitResponse: Response containing:
                - splits: List of SplitResult objects with name, start_page, and end_page for each section.

        Raises:
            HTTPException: If the request fails.

        Example:
            ```python
            response = await retab.documents.split(
                document="invoice_batch.pdf",
                model="gemini-2.5-flash",
                categories=[
                    {"name": "invoice", "description": "Invoice documents with billing information"},
                    {"name": "receipt", "description": "Receipt documents for payments"},
                    {"name": "contract", "description": "Legal contract documents"},
                ]
            )
            for split in response.splits:
                print(f"{split.name}: pages {split.start_page}-{split.end_page}")
            ```
        """
        request = self._prepare_split(
            document=document,
            categories=categories,
            model=model,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return SplitResponse.model_validate(response)
