import base64
from io import IOBase
from pathlib import Path
from typing import Any, List

import PIL.Image
from pydantic import HttpUrl
from ..._resource import AsyncAPIResource, SyncAPIResource
from ...utils.mime import MIMEData, prepare_mime_document
from ...types.documents.extract import RetabParsedChatCompletion
from ...types.standards import PreparedRequest


class DeploymentsMixin:
    def prepare_extract(
        self,
        project_id: str,
        iteration_id: str,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: list[Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl] | None = None,
        temperature: float | None = None,
        seed: int | None = None,
        store: bool = True,
    ) -> PreparedRequest:
        """Prepare a request to extract documents from a deployment.

        Args:
            project_id: ID of the project
            iteration_id: ID of the iteration
            document: Single document to process (mutually exclusive with documents)
            documents: List of documents to process (mutually exclusive with document)
            temperature: Optional temperature override
            seed: Optional seed for reproducibility
            store: Whether to store the results

        Returns:
            PreparedRequest: The prepared request
        """
        # Validate that either document or documents is provided, but not both
        if not document and not documents:
            raise ValueError("Either 'document' or 'documents' must be provided")

        if document and documents:
            raise ValueError("Provide either 'document' (single) or 'documents' (multiple), not both")

        # Prepare form data parameters
        form_data = {
            "temperature": temperature,
            "seed": seed,
            "store": store,
        }
        # Remove None values
        form_data = {k: v for k, v in form_data.items() if v is not None}

        # Prepare files for upload
        files = {}
        if document:
            # Convert document to MIMEData if needed
            mime_document = prepare_mime_document(document)
            # Single document upload
            files["document"] = (mime_document.filename, base64.b64decode(mime_document.content), mime_document.mime_type)
        elif documents:
            # Multiple documents upload - httpx supports multiple files with same field name using a list
            files_list = []
            for doc in documents:
                # Convert each document to MIMEData if needed
                mime_doc = prepare_mime_document(doc)
                files_list.append(
                    (
                        "documents",  # field name
                        (mime_doc.filename, base64.b64decode(mime_doc.content), mime_doc.mime_type),
                    )
                )
            files = files_list

        url = f"/v1/deployments/extract/{project_id}/{iteration_id}"

        return PreparedRequest(method="POST", url=url, form_data=form_data, files=files)


class Deployments(SyncAPIResource, DeploymentsMixin):
    """Deployments API wrapper for managing deployment configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    def extract(
        self,
        project_id: str,
        iteration_id: str,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: List[Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl] | None = None,
        temperature: float | None = None,
        seed: int | None = None,
        store: bool = True,
    ) -> RetabParsedChatCompletion:
        """Extract documents from a deployment.

        Args:
            project_id: ID of the project
            iteration_id: ID of the iteration
            document: Single document to process (mutually exclusive with documents)
            documents: List of documents to process (mutually exclusive with document)
            temperature: Optional temperature override
            seed: Optional seed for reproducibility
            store: Whether to store the results

        Returns:
            RetabParsedChatCompletion: The processing result
        """
        request = self.prepare_extract(project_id=project_id, iteration_id=iteration_id, document=document, documents=documents, temperature=temperature, seed=seed, store=store)
        response = self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)


class AsyncDeployments(AsyncAPIResource, DeploymentsMixin):
    """Async Deployments API wrapper for managing deployment configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    async def extract(
        self,
        project_id: str,
        iteration_id: str,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: List[Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl] | None = None,
        temperature: float | None = None,
        seed: int | None = None,
        store: bool = True,
    ) -> RetabParsedChatCompletion:
        """Extract documents from a deployment.

        Args:
            project_id: ID of the project
            iteration_id: ID of the iteration
            document: Single document to process (mutually exclusive with documents)
            documents: List of documents to process (mutually exclusive with document)
            temperature: Optional temperature override
            seed: Optional seed for reproducibility
            store: Whether to store the results

        Returns:
            RetabParsedChatCompletion: The processing result
        """
        request = self.prepare_extract(project_id=project_id, iteration_id=iteration_id, document=document, documents=documents, temperature=temperature, seed=seed, store=store)
        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)
