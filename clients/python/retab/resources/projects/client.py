import base64
import json
from io import IOBase
from pathlib import Path
from typing import Any, Dict, List, Optional, Sequence

import PIL.Image
from pydantic import HttpUrl
from ..._resource import AsyncAPIResource, SyncAPIResource
from ...utils.mime import MIMEData, prepare_mime_document
from ...types.documents.extract import RetabParsedChatCompletion
from ...types.projects import Project, PatchProjectRequest, CreateProjectRequest
from ...types.standards import PreparedRequest, DeleteResponse, FieldUnset

class ProjectsMixin:
    def prepare_create(
        self,
        name: str,
        json_schema: dict[str, Any],
        **extra_body: Any,
    ) -> PreparedRequest:
        # Use BaseProject model
        eval_dict: dict[str, Any] = {
            "name": name,
            "json_schema": json_schema,
        }
        if extra_body:
            eval_dict.update(extra_body)

        eval_data = CreateProjectRequest(**eval_dict)
        return PreparedRequest(method="POST", url="/v1/projects", data=eval_data.model_dump(exclude_unset=True, mode="json"))

    def prepare_get(self, project_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/projects/{project_id}")

    def prepare_update(
        self,
        project_id: str,
        name: str = FieldUnset,
        json_schema: dict[str, Any] = FieldUnset,
        **extra_body: Any,
    ) -> PreparedRequest:
        """
        Prepare a request to update an project with partial updates.

        Only the provided fields will be updated. Fields set to None will be excluded from the update.
        """
        # Build a dictionary with only the provided fields
        update_dict: dict[str, Any] = {}
        if name is not FieldUnset:
            update_dict["name"] = name
        if json_schema is not FieldUnset:
            update_dict["json_schema"] = json_schema
        if extra_body:
            update_dict.update(extra_body)

        data = PatchProjectRequest(**update_dict).model_dump(exclude_unset=True, mode="json")

        return PreparedRequest(method="PATCH", url=f"/v1/projects/{project_id}", data=data)

    def prepare_list(self, **extra_params: Any) -> PreparedRequest:
        """
        Prepare a request to list projects.

        Usage:
        >>> client.projects.list()                          # List all projects

        Returns:
            PreparedRequest: The prepared request
        """
        params = extra_params or None
        return PreparedRequest(method="GET", url="/v1/projects", params=params)

    def prepare_delete(self, id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/projects/{id}")

    def prepare_publish(self, project_id: str, origin: Optional[str] = None) -> PreparedRequest:
        params = {"origin": origin} if origin else None
        return PreparedRequest(method="POST", url=f"/v1/projects/{project_id}/publish", params=params)

    def prepare_extract(
        self,
        project_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: Sequence[Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl] | None = None,
        model: str | None = None,
        temperature: float | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        seed: int | None = None,
        store: bool = True,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
    ) -> PreparedRequest:
        """Prepare a request to extract documents from a project.

        Args:
            project_id: ID of the project
            iteration_id: Optional ID of the iteration. If None, uses project base.
            document: Single document to process (mutually exclusive with documents)
            documents: List of documents to process (mutually exclusive with document)
            model: Optional model override
            temperature: Optional temperature override
            image_resolution_dpi: Optional image resolution DPI override
            n_consensus: Optional number of consensus extractions
            store: Whether to store the results
            seed: Optional seed for reproducibility
            metadata: User-defined metadata for the extraction

        Returns:
            PreparedRequest: The prepared request
        """
        # Validate that either document or documents is provided, but not both
        if not document and not documents:
            raise ValueError("Either 'document' or 'documents' must be provided")

        if document and documents:
            raise ValueError("Provide either 'document' (single) or 'documents' (multiple), not both")

        # Prepare form data parameters
        # Note: metadata must be JSON-serialized since httpx multipart forms only accept primitive types
        form_data = {
            "model": model,
            "temperature": temperature,
            "image_resolution_dpi": image_resolution_dpi,
            "n_consensus": n_consensus,
            "seed": seed,
            "store": store,
            "metadata": json.dumps(metadata) if metadata else None,
            "extraction_id": extraction_id,
        }
        if extra_form:
            form_data.update(extra_form)
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

        url = f"/v1/projects/extract/{project_id}" if iteration_id is None else f"/v1/projects/extract/{project_id}/{iteration_id}"

        return PreparedRequest(method="POST", url=url, form_data=form_data, files=files)


class Projects(SyncAPIResource, ProjectsMixin):
    """Projects API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def create(
        self,
        name: str,
        json_schema: dict[str, Any],
        **extra_body: Any,
    ) -> Project:
        """
        Create a new project.

        Args:
            name: The name of the project
            json_schema: The json schema of the project
        Returns:
            Project: The created project
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(name, json_schema, **extra_body)
        response = self._client._prepared_request(request)
        return Project(**response)

    def get(self, project_id: str) -> Project:
        """
        Get an project by ID.

        Args:
            project_id: The ID of the project to retrieve

        Returns:
            Project: The project
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(project_id)
        response = self._client._prepared_request(request)
        return Project(**response)

    def list(self, **extra_params: Any) -> List[Project]:
        """
        List projects for a project.

        Args:
        Returns:
            List[Project]: List of projects
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(**extra_params)
        response = self._client._prepared_request(request)
        return [Project(**item) for item in response.get("data", [])]

    def delete(self, project_id: str) -> DeleteResponse:
        """
        Delete an project.

        Args:
            project_id: The ID of the project to delete

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(project_id)
        return self._client._prepared_request(request)

    def publish(self, project_id: str, origin: Optional[str] = None) -> Project:
        """Publish a project's draft configuration.
        
        Args:
            project_id: The ID of the project to publish
            origin: Optional origin identifier (e.g., iteration ID). If an iteration ID 
                    is provided, the project will be published using that iteration's config.
        """
        request = self.prepare_publish(project_id, origin=origin)
        response = self._client._prepared_request(request)
        return Project(**response)

    def extract(
        self,
        project_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: Sequence[Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl] | None = None,
        model: str | None = None,
        temperature: float | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        seed: int | None = None,
        store: bool = True,
        **extra_form: Any,
    ) -> RetabParsedChatCompletion:
        """Extract documents from a project.

        Args:
            project_id: ID of the project
            iteration_id: Optional ID of the iteration. If None, uses project base.
            document: Single document to process (mutually exclusive with documents)
            documents: List of documents to process (mutually exclusive with document)
            model: Optional model override
            temperature: Optional temperature override
            image_resolution_dpi: Optional image resolution DPI override
            seed: Optional seed for reproducibility
            store: Whether to store the results

        Returns:
            RetabParsedChatCompletion: The processing result
        """
        request = self.prepare_extract(
            project_id=project_id,
            iteration_id=iteration_id,
            document=document,
            documents=documents,
            model=model,
            temperature=temperature,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            seed=seed,
            store=store,
            **extra_form,
        )
        response = self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)


class AsyncProjects(AsyncAPIResource, ProjectsMixin):
    """Async Projects API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def create(self, name: str, json_schema: dict[str, Any], **extra_body: Any) -> Project:
        """
        Create a new project.

        Args:
            name: The name of the project
            json_schema: The json schema of the project
        Returns:
            Project: The created project
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(name, json_schema, **extra_body)
        response = await self._client._prepared_request(request)
        return Project(**response)

    async def get(self, project_id: str) -> Project:
        """
        Get an project by ID.

        Args:
            project_id: The ID of the project to retrieve

        Returns:
            Project: The project
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(project_id)
        response = await self._client._prepared_request(request)
        return Project(**response)

    async def list(self, **extra_params: Any) -> List[Project]:
        """
        List projects for a project.

        Returns:
            List[Project]: List of projects
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(**extra_params)
        response = await self._client._prepared_request(request)
        return [Project(**item) for item in response.get("data", [])]

    async def delete(self, project_id: str) -> DeleteResponse:
        """
        Delete an project.

        Args:
            project_id: The ID of the project to delete

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(project_id)
        return await self._client._prepared_request(request)

    async def publish(self, project_id: str, origin: Optional[str] = None) -> Project:
        """Publish a project's draft configuration.
        
        Args:
            project_id: The ID of the project to publish
            origin: Optional origin identifier (e.g., iteration ID). If an iteration ID 
                    is provided, the project will be published using that iteration's config.
        """
        request = self.prepare_publish(project_id, origin=origin)
        response = await self._client._prepared_request(request)
        return Project(**response)

    async def extract(
        self,
        project_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: Sequence[Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl] | None = None,
        model: str | None = None,
        temperature: float | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        seed: int | None = None,
        store: bool = True,
        **extra_form: Any,
    ) -> RetabParsedChatCompletion:
        """Extract documents from a project.

        Args:
            project_id: ID of the project
            iteration_id: Optional ID of the iteration. If None, uses project base.
            document: Single document to process (mutually exclusive with documents)
            documents: List of documents to process (mutually exclusive with document)
            model: Optional model override
            temperature: Optional temperature override
            image_resolution_dpi: Optional image resolution DPI override
            n_consensus: Optional number of consensus extractions
            seed: Optional seed for reproducibility
            store: Whether to store the results

        Returns:
            RetabParsedChatCompletion: The processing result
        """
        request = self.prepare_extract(project_id=project_id, iteration_id=iteration_id, document=document, documents=documents, model=model, temperature=temperature, image_resolution_dpi=image_resolution_dpi, n_consensus=n_consensus, seed=seed, store=store)
        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)
