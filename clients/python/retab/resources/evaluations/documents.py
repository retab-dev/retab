from io import IOBase
from pathlib import Path
from typing import Any, Dict, List, Union

import PIL.Image
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...utils.mime import prepare_mime_document
from ...types.evaluations import DocumentItem, EvaluationDocument, PatchEvaluationDocumentRequest
from ...types.mime import MIMEData
from ...types.standards import PreparedRequest, DeleteResponse, FieldUnset
from ...types.documents.extractions import RetabParsedChatCompletion


class DocumentsMixin:
    def prepare_get(self, evaluation_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/evaluations/{evaluation_id}/documents/{document_id}")

    def prepare_create(self, evaluation_id: str, document: MIMEData, annotation: dict[str, Any]) -> PreparedRequest:
        # Serialize the MIMEData
        document_item = DocumentItem(mime_data=document, annotation=annotation, annotation_metadata=None)
        return PreparedRequest(method="POST", url=f"/v1/evaluations/{evaluation_id}/documents", data=document_item.model_dump(mode="json"))

    def prepare_list(self, evaluation_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/evaluations/{evaluation_id}/documents")

    def prepare_update(self, evaluation_id: str, document_id: str, annotation: dict[str, Any] = FieldUnset) -> PreparedRequest:
        update_request = PatchEvaluationDocumentRequest(annotation=annotation)
        return PreparedRequest(
            method="PATCH", url=f"/v1/evaluations/{evaluation_id}/documents/{document_id}", data=update_request.model_dump(mode="json", exclude_unset=True, exclude_defaults=True)
        )

    def prepare_delete(self, evaluation_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/evaluations/{evaluation_id}/documents/{document_id}")

    def prepare_llm_annotate(self, evaluation_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"/v1/evaluations/{evaluation_id}/documents/{document_id}/llm-annotate", data={"stream": False})


class Documents(SyncAPIResource, DocumentsMixin):
    """Documents API wrapper for evaluations"""

    def create(self, evaluation_id: str, document: Union[Path, str, IOBase, MIMEData, PIL.Image.Image, HttpUrl], annotation: Dict[str, Any]) -> EvaluationDocument:
        """
        Create a document for an evaluation.

        Args:
            evaluation_id: The ID of the evaluation
            document: The document to process. Can be:
                - A file path (Path or str)
                - A file-like object (IOBase)
                - A MIMEData object
                - A PIL Image object
                - A URL (HttpUrl)
            annotation: The ground truth for the document

        Returns:
            EvaluationDocument: The created document
        Raises:
            HTTPException if the request fails
        """
        # Convert document to MIME data format
        mime_document: MIMEData = prepare_mime_document(document)

        # Let prepare_create handle the serialization
        request = self.prepare_create(evaluation_id, mime_document, annotation)
        response = self._client._prepared_request(request)
        return EvaluationDocument(**response)

    def list(self, evaluation_id: str) -> List[EvaluationDocument]:
        """
        List documents for an evaluation.

        Args:
            evaluation_id: The ID of the evaluation

        Returns:
            List[EvaluationDocument]: List of documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(evaluation_id)
        response = self._client._prepared_request(request)
        return [EvaluationDocument(**item) for item in response.get("data", [])]

    def get(self, evaluation_id: str, document_id: str) -> EvaluationDocument:
        """
        Get a document by ID.

        Args:
            evaluation_id: The ID of the evaluation
            document_id: The ID of the document

        Returns:
            EvaluationDocument: The document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(evaluation_id, document_id)
        response = self._client._prepared_request(request)
        return EvaluationDocument(**response)

    def update(self, evaluation_id: str, document_id: str, annotation: dict[str, Any] = FieldUnset) -> EvaluationDocument:
        """
        Update a document.

        Args:
            evaluation_id: The ID of the evaluation
            document_id: The ID of the document
            annotation: The ground truth for the document
        Returns:
            EvaluationDocument: The updated document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(evaluation_id, document_id, annotation=annotation)
        response = self._client._prepared_request(request)
        return EvaluationDocument(**response)

    def delete(self, evaluation_id: str, document_id: str) -> DeleteResponse:
        """
        Delete a document.

        Args:
            evaluation_id: The ID of the evaluation
            document_id: The ID of the document

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(evaluation_id, document_id)
        return self._client._prepared_request(request)

    def llm_annotate(self, evaluation_id: str, document_id: str) -> RetabParsedChatCompletion:
        """
        Annotate a document with an LLM. This method updates the document (within the evaluation) with the latest extraction.
        """
        request = self.prepare_llm_annotate(evaluation_id, document_id)
        response = self._client._prepared_request(request)
        return RetabParsedChatCompletion(**response)


class AsyncDocuments(AsyncAPIResource, DocumentsMixin):
    """Async Documents API wrapper for evaluations"""

    async def create(self, evaluation_id: str, document: Union[Path, str, IOBase, MIMEData, PIL.Image.Image, HttpUrl], annotation: Dict[str, Any]) -> EvaluationDocument:
        """
        Create a document for an evaluation.

        Args:
            evaluation_id: The ID of the evaluation
            document: The document to process. Can be:
                - A file path (Path or str)
                - A file-like object (IOBase)
                - A MIMEData object
                - A PIL Image object
                - A URL (HttpUrl)
            annotation: The ground truth for the document

        Returns:
            EvaluationDocument: The created document
        Raises:
            HTTPException if the request fails
        """
        # Convert document to MIME data format
        mime_document: MIMEData = prepare_mime_document(document)

        # Let prepare_create handle the serialization
        request = self.prepare_create(evaluation_id, mime_document, annotation)
        response = await self._client._prepared_request(request)
        return EvaluationDocument(**response)

    async def list(self, evaluation_id: str) -> List[EvaluationDocument]:
        """
        List documents for an evaluation.

        Args:
            evaluation_id: The ID of the evaluation

        Returns:
            List[EvaluationDocument]: List of documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(evaluation_id)
        response = await self._client._prepared_request(request)
        return [EvaluationDocument(**item) for item in response.get("data", [])]

    async def update(self, evaluation_id: str, document_id: str, annotation: dict[str, Any] = FieldUnset) -> EvaluationDocument:
        """
        Update a document.

        Args:
            evaluation_id: The ID of the evaluation
            document_id: The ID of the document
            annotation: The ground truth for the document

        Returns:
            EvaluationDocument: The updated document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(evaluation_id, document_id, annotation)
        response = await self._client._prepared_request(request)
        return EvaluationDocument(**response)

    async def delete(self, evaluation_id: str, document_id: str) -> DeleteResponse:
        """
        Delete a document.

        Args:
            evaluation_id: The ID of the evaluation
            document_id: The ID of the document

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(evaluation_id, document_id)
        return await self._client._prepared_request(request)

    async def llm_annotate(self, evaluation_id: str, document_id: str) -> RetabParsedChatCompletion:
        """
        Annotate a document with an LLM.
        This method updates the document (within the evaluation) with the latest extraction.
        """
        request = self.prepare_llm_annotate(evaluation_id, document_id)
        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion(**response)
