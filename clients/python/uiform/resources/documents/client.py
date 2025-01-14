from typing import Any
from pathlib import Path
from io import IOBase
from ...types.modalities import Modality
from ..._utils.mime import prepare_mime_document
from ..._resource import SyncAPIResource, AsyncAPIResource
from ...types.documents.create_messages import DocumentCreateMessageRequest, DocumentMessage
from .extractions import Extractions, AsyncExtractions

class BaseDocumentsMixin:
    def _prepare_create_messages(
        self,
        document: Path | str | IOBase, 
        modality: Modality = "native", 
        text_operations: dict[str, Any] | None = None
    ) -> DocumentCreateMessageRequest:
        mime_document = prepare_mime_document(document)

        data: dict[str, Any] = {
            "document": mime_document.model_dump(),
            "modality": modality,
            "text_operations": text_operations
        }
        return DocumentCreateMessageRequest.model_validate(data)


class Documents(SyncAPIResource, BaseDocumentsMixin): 
    """Documents API wrapper"""

    # TODO: Add batch methods
    #client.documents.batch.extract()
    #client.documents.batch.preprocess()

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.extractions = Extractions(client=client)
        # self.batch = Batch(client=client)

    def create_messages(self, 
            document: Path | str | IOBase, 
            modality: Modality = "native", 
            text_operations: dict[str, Any] | None = None) -> DocumentMessage:
        """
        Create document messages from a file using the UiForm API.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
            modality: The processing modality to use. Defaults to "native".
            text_operations: Optional dictionary of text processing operations to apply.

        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            UiformAPIError: If the API request fails.
        """
        loading_request = self._prepare_create_messages(document, modality, text_operations)
        response = self._client._request("POST", "/api/v1/documents/create_messages", data=loading_request.model_dump())
        return DocumentMessage.model_validate(response)



class AsyncDocuments(AsyncAPIResource, BaseDocumentsMixin):
    """Documents API wrapper for asynchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.extractions = AsyncExtractions(client=client)

    async def create_messages(self, 
            document: Path | str | IOBase, 
            modality: Modality = "native",
            text_operations: dict[str, Any] | None = None) -> DocumentMessage:
        """
        Create document messages from a file using the UiForm API asynchronously.

        Args:
            document: The document to process. Can be a file path (Path or str) or a file-like object.
            modality: The processing modality to use. Defaults to "native".
            text_operations: Optional dictionary of text processing operations to apply.

        Returns:
            DocumentMessage: The processed document message containing extracted content.

        Raises:
            UiformAPIError: If the API request fails.
        """
        loading_request = self._prepare_create_messages(document, modality, text_operations)
        response = await self._client._request("POST", "/api/v1/documents/create_messages", data=loading_request.model_dump())
        return DocumentMessage.model_validate(response)

