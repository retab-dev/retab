from io import IOBase
from pathlib import Path
from typing import Any, AsyncGenerator, Generator, Optional, get_args

from ..extractions import maybe_parse_to_pydantic
from ...._resource import AsyncAPIResource, SyncAPIResource
from ...._utils.ai_model import assert_valid_model_extraction
from ...._utils.mime import prepare_mime_document
from ...._utils.stream_context_managers import as_async_context_manager, as_context_manager
from ....types.documents.create_messages import ChatCompletionUiformMessage
from ....types.documents.parse import DocumentExtractRequest, DocumentExtractResponse
from ....types.mime import MIMEData
from ....types.modalities import Modality
from typing import Literal

DocumentAITemplate = Literal['bank_statement', 'contract', 'driver_license', 'expense', 'identity_proofing', 'invoice', 'passport', 'pay_slip', 'w2']

from pydantic import BaseModel
from .schemas import Invoice, Contract, BankStatement, W2, Expense, PaySlip, DriverLicense, Passport, IdentityProofing

def load_template_schema(template: DocumentAITemplate) -> type[BaseModel]:
    if template == "bank_statement":
        return BankStatement
    elif template == "contract":
        return Contract
    elif template == "driver_license":
        return DriverLicense
    elif template == "expense":
        return Expense
    elif template == "identity_proofing":
        return IdentityProofing
    elif template == "invoice":
        return Invoice
    elif template == "passport":
        return Passport
    elif template == "pay_slip":
        return PaySlip
    elif template == "w2":
        return W2


class BaseDocumentAIMixin:
    def prepare_extraction(
        self,
        template: DocumentAITemplate,
        document: Path | str | IOBase | MIMEData | None,
        text_operations: Optional[dict[str, Any]],
        image_operations: Optional[dict[str, Any]],
        model: str,
        temperature: float | None,
        messages: list[ChatCompletionUiformMessage],
        modality: Modality,
        stream: bool,
        store: bool = False,
    ) -> DocumentExtractRequest:
        assert_valid_model_extraction(model)

        assert template in get_args(DocumentAITemplate), "Invalid template, template must be one of the following: " + ", ".join(get_args(DocumentAITemplate))
        template_pydantic_model = load_template_schema(template)

        data = {
            "json_schema": template_pydantic_model.model_json_schema(),
            "document": prepare_mime_document(document).model_dump() if document is not None else None,
            "model": model,
            "temperature": temperature,
            "stream": stream,
            "modality": modality,
            "messages": messages,
            "store": store,
        }
        if text_operations:
            data["text_operations"] = text_operations
        if image_operations:
            data["image_operations"] = image_operations

        # Validate DocumentAPIRequest data (raises exception if invalid)
        return DocumentExtractRequest.model_validate(data)

    def prepare_parse(
        self,
        template: DocumentAITemplate,
        document: Path | str | IOBase | MIMEData | None,
        text_operations: Optional[dict[str, Any]],
        image_operations: Optional[dict[str, Any]],
        model: str,
        temperature: float,
        messages: list[ChatCompletionUiformMessage],
        modality: Modality,
        store: bool = False,
    ) -> DocumentExtractRequest:
        stream = False
        return self.prepare_extraction(
            template, document, text_operations, image_operations, model, temperature, messages, modality, stream, store
        )
    
    def prepare_stream(
        self,
        template: DocumentAITemplate,
        document: Path | str | IOBase | MIMEData | None,
        text_operations: Optional[dict[str, Any]],
        image_operations: Optional[dict[str, Any]],
        model: str,
        temperature: float,
        messages: list[ChatCompletionUiformMessage],
        modality: Modality,
        store: bool = False,
    ) -> DocumentExtractRequest:
        stream = True
        return self.prepare_extraction(
            template, document, text_operations, image_operations, model, temperature, messages, modality, stream, store
        )


class DocumentAIs(SyncAPIResource, BaseDocumentAIMixin):
    """Extraction API wrapper"""

    def parse(
        self,
        template: DocumentAITemplate,
        document: Path | str | IOBase | MIMEData | None,
        text_operations: Optional[dict[str, Any]] = None,
        image_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        messages: list[ChatCompletionUiformMessage] = [],
        modality: Modality = "native",
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> DocumentExtractResponse:
        """
        Process a document using the UiForm API.

        Args:
            json_schema: JSON schema defining the expected data structure
            document: Single document (as MIMEData) to process
            model: The AI model to use for processing
            temperature: Model temperature setting (0-1)
            text_operations: Optional context with regex instructions or other metadata
            idempotency_key: Idempotency key for request
        Returns:
            DocumentAPIResponse
        Raises:
            HTTPException if the request fails
        """

        assert (document is not None) or (messages is not None), "Either document or messages must be provided"

        # Validate DocumentAPIRequest data (raises exception if invalid)
        request = self.prepare_parse(template, document, text_operations, image_operations, model, temperature, messages, modality, store)
        response = self._client._request("POST", "/api/v1/documents/extractions", data=request.model_dump(), idempotency_key=idempotency_key)
        return maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(response))

    @as_context_manager
    def stream(
        self,
        template: DocumentAITemplate,
        document: Path | str | IOBase | MIMEData | None,
        text_operations: Optional[dict[str, Any]] = None,
        image_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        messages: list[ChatCompletionUiformMessage] = [],
        modality: Modality = "native",
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> Generator[DocumentExtractResponse, None, None]:
        """
        Process a document using the UiForm API with streaming enabled.

        Args:
            json_schema: JSON schema defining the expected data structure
            document: Single document (as MIMEData) to process
            text_operations: Optional context with regex instructions or other metadata
            model: The AI model to use for processing
            temperature: Model temperature setting (0-1)
            modality: Modality of the document (e.g., native)
            messages: List of chat completion messages for context

        Returns:
            Generator[DocumentExtractResponse]: Stream of parsed responses
        Raises:
            HTTPException if the request fails
        Usage:
        ```python
        with uiform.documents.extractions.stream(json_schema, document, text_operations, model, temperature, messages, modality) as stream:
            for response in stream:
                print(response)
        ```
        """
        request = self.prepare_stream(template, document, text_operations, image_operations, model, temperature, messages, modality, store)

        # Request the stream and return a context manager
        chunk_json: Any = None
        for chunk_json in self._client._request_stream("POST", "/api/v1/documents/extractions", data=request.model_dump(), idempotency_key=idempotency_key):
            if not chunk_json:
                continue
            yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json), allow_partial=True)
        yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json))


class AsyncDocumentAIs(AsyncAPIResource, BaseDocumentAIMixin):
    """Extraction API wrapper for asynchronous usage."""

    async def parse(
        self,
        template: DocumentAITemplate,
        document: Path | str | IOBase | MIMEData | None,
        text_operations: Optional[dict[str, Any]] = None,
        image_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        messages: list[ChatCompletionUiformMessage] = [],
        modality: Modality = "native",
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> DocumentExtractResponse:
        """
        Extract structured data from a document asynchronously.

        Args:
            json_schema: JSON schema defining the expected data structure.
            document: Path, string, or file-like object representing the document.
            text_operations: Optional additional context for the model.
            model: The AI model to use.
            temperature: Model temperature setting (0-1).
            modality: Modality of the document (e.g., native).

        Returns:
            DocumentExtractResponse: Parsed response from the API.
        """
        request = self.prepare_parse(template, document, text_operations, image_operations, model, temperature, messages, modality, store)
        response = await self._client._request("POST", "/api/v1/documents/extractions", data=request.model_dump(), idempotency_key=idempotency_key)
        return maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(response))

    @as_async_context_manager
    async def stream(
        self,
        template: DocumentAITemplate,
        document: Path | str | IOBase | MIMEData | None,
        text_operations: Optional[dict[str, Any]] = None,
        image_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        messages: list[ChatCompletionUiformMessage] = [],
        modality: Modality = "native",
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> AsyncGenerator[DocumentExtractResponse, None]:
        """
        Extract structured data from a document asynchronously with streaming.

        Args:
            json_schema: JSON schema defining the expected data structure.
            document: Path, string, or file-like object representing the document.
            text_operations: Optional additional context for the model.
            model: The AI model to use.
            temperature: Model temperature setting (0-1).
            modality: Modality of the document (e.g., native).
            idempotency_key: Idempotency key for request
        Returns:
            AsyncGenerator[DocumentExtractResponse, None]: Stream of parsed responses.

        Usage:
        ```python
        async with uiform.documents.extractions.stream(json_schema, document, text_operations, model, temperature, messages, modality) as stream:
            async for response in stream:
                print(response)
        ```
        """
        request = self.prepare_stream(template, document, text_operations, image_operations, model, temperature, messages, modality, store)
        chunk_json: Any = None
        async for chunk_json in self._client._request_stream("POST", "/api/v1/documents/extractions", data=request.model_dump(), idempotency_key=idempotency_key):
            if not chunk_json:
                continue
            
            yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json), allow_partial=True)
        # Last chunk with full parsed response
        yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json))  
