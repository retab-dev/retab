from io import IOBase
from pathlib import Path
from typing import Any, Dict, List, Union

import PIL.Image
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...utils.mime import prepare_mime_document
from ...types.projects import DocumentItem, ProjectDocument, PatchProjectDocumentRequest
from ...types.predictions import PredictionMetadata
from ...types.mime import MIMEData
from ...types.standards import PreparedRequest, DeleteResponse, FieldUnset
from ...types.documents.extractions import RetabParsedChatCompletion


class DocumentsMixin:
    def prepare_get(self, project_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/projects/{project_id}/documents/{document_id}")

    def prepare_create(self, project_id: str, document: MIMEData, annotation: dict[str, Any], annotation_metadata: dict[str, Any] | None = None) -> PreparedRequest:
        # Serialize the MIMEData
        document_item = DocumentItem(mime_data=document, annotation=annotation, annotation_metadata=PredictionMetadata(**annotation_metadata) if annotation_metadata else None)
        return PreparedRequest(method="POST", url=f"/v1/projects/{project_id}/documents", data=document_item.model_dump(mode="json"))

    def prepare_list(self, project_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/projects/{project_id}/documents")

    def prepare_update(self, project_id: str, document_id: str, annotation: dict[str, Any]) -> PreparedRequest:
        update_request = PatchProjectDocumentRequest(annotation=annotation)
        return PreparedRequest(method="PATCH", url=f"/v1/projects/{project_id}/documents/{document_id}", data=update_request.model_dump(mode="json", exclude_unset=True))

    def prepare_delete(self, project_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/projects/{project_id}/documents/{document_id}")

    def prepare_llm_annotate(self, project_id: str, document_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"/v1/projects/{project_id}/documents/{document_id}/llm-annotate", data={"stream": False})


class Documents(SyncAPIResource, DocumentsMixin):
    """Documents API wrapper for evaluations"""

    def create(
        self,
        project_id: str,
        document: Union[Path, str, IOBase, MIMEData, PIL.Image.Image, HttpUrl],
        annotation: Dict[str, Any],
        annotation_metadata: Dict[str, Any] | None = None,
    ) -> ProjectDocument:
        """
        Create a document for an evaluation.

        Args:
            project_id: The ID of the evaluation
            document: The document to process. Can be:
                - A file path (Path or str)
                - A file-like object (IOBase)
                - A MIMEData object
                - A PIL Image object
                - A URL (HttpUrl)
            annotation: The ground truth for the document

        Returns:
            ProjectDocument: The created document
        Raises:
            HTTPException if the request fails
        """
        # Convert document to MIME data format
        mime_document: MIMEData = prepare_mime_document(document)

        # Let prepare_create handle the serialization
        request = self.prepare_create(project_id, mime_document, annotation, annotation_metadata)
        response = self._client._prepared_request(request)
        return ProjectDocument(**response)

    def list(self, project_id: str) -> List[ProjectDocument]:
        """
        List documents for an evaluation.

        Args:
            project_id: The ID of the evaluation

        Returns:
            List[ProjectDocument]: List of documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(project_id)
        response = self._client._prepared_request(request)
        return [ProjectDocument(**item) for item in response.get("data", [])]

    def get(self, project_id: str, document_id: str) -> ProjectDocument:
        """
        Get a document by ID.

        Args:
            project_id: The ID of the evaluation
            document_id: The ID of the document

        Returns:
            ProjectDocument: The document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(project_id, document_id)
        response = self._client._prepared_request(request)
        return ProjectDocument(**response)

    def update(self, project_id: str, document_id: str, annotation: dict[str, Any]) -> ProjectDocument:
        """
        Update a document.

        Args:
            project_id: The ID of the evaluation
            document_id: The ID of the document
            annotation: The ground truth for the document
        Returns:
            ProjectDocument: The updated document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(project_id, document_id, annotation=annotation)
        response = self._client._prepared_request(request)
        return ProjectDocument(**response)

    def delete(self, project_id: str, document_id: str) -> DeleteResponse:
        """
        Delete a document.

        Args:
            project_id: The ID of the evaluation
            document_id: The ID of the document

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(project_id, document_id)
        return self._client._prepared_request(request)

    def llm_annotate(self, project_id: str, document_id: str) -> RetabParsedChatCompletion:
        """
        Annotate a document with an LLM. This method updates the document (within the evaluation) with the latest extraction.
        """
        request = self.prepare_llm_annotate(project_id, document_id)
        response = self._client._prepared_request(request)
        return RetabParsedChatCompletion(**response)


class AsyncDocuments(AsyncAPIResource, DocumentsMixin):
    """Async Documents API wrapper for evaluations"""

    async def create(
        self,
        project_id: str,
        document: Union[Path, str, IOBase, MIMEData, PIL.Image.Image, HttpUrl],
        annotation: Dict[str, Any],
        annotation_metadata: Dict[str, Any] | None = None,
    ) -> ProjectDocument:
        """
        Create a document for an evaluation.

        Args:
            project_id: The ID of the evaluation
            document: The document to process. Can be:
                - A file path (Path or str)
                - A file-like object (IOBase)
                - A MIMEData object
                - A PIL Image object
                - A URL (HttpUrl)
            annotation: The ground truth for the document
            annotation_metadata: The metadata of the annotation
        Returns:
            ProjectDocument: The created document
        Raises:
            HTTPException if the request fails
        """
        # Convert document to MIME data format
        mime_document: MIMEData = prepare_mime_document(document)

        # Let prepare_create handle the serialization
        request = self.prepare_create(project_id, mime_document, annotation, annotation_metadata)
        response = await self._client._prepared_request(request)
        return ProjectDocument(**response)

    async def list(self, project_id: str) -> List[ProjectDocument]:
        """
        List documents for an evaluation.

        Args:
            project_id: The ID of the evaluation

        Returns:
            List[ProjectDocument]: List of documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(project_id)
        response = await self._client._prepared_request(request)
        return [ProjectDocument(**item) for item in response.get("data", [])]

    async def update(self, project_id: str, document_id: str, annotation: dict[str, Any]) -> ProjectDocument:
        """
        Update a document.

        Args:
            project_id: The ID of the evaluation
            document_id: The ID of the document
            annotation: The ground truth for the document

        Returns:
            ProjectDocument: The updated document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(project_id, document_id, annotation)
        response = await self._client._prepared_request(request)
        return ProjectDocument(**response)

    async def delete(self, project_id: str, document_id: str) -> DeleteResponse:
        """
        Delete a document.

        Args:
            project_id: The ID of the evaluation
            document_id: The ID of the document

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(project_id, document_id)
        return await self._client._prepared_request(request)

    async def llm_annotate(self, project_id: str, document_id: str) -> RetabParsedChatCompletion:
        """
        Annotate a document with an LLM.
        This method updates the document (within the evaluation) with the latest extraction.
        """
        request = self.prepare_llm_annotate(project_id, document_id)
        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion(**response)
