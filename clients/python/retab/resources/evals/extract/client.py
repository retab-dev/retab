import base64
import json
from io import IOBase
from pathlib import Path
from typing import Any, Dict, List, Optional, Sequence

import PIL.Image
from pydantic import HttpUrl
from ...._resource import AsyncAPIResource, SyncAPIResource
from ....utils.mime import MIMEData, prepare_mime_document
from ....types.documents.extract import RetabParsedChatCompletion
from ....types.projects import Project, PatchProjectRequest, CreateProjectRequest
from ....types.standards import PreparedRequest, DeleteResponse, UNSET, _Unset
from .datasets import Datasets, AsyncDatasets

BASE = "/evals/extract"


class ExtractProjectsMixin:
    def prepare_create(
        self,
        name: str,
        json_schema: dict[str, Any],
        **extra_body: Any,
    ) -> PreparedRequest:
        eval_dict: dict[str, Any] = {
            "name": name,
            "json_schema": json_schema,
        }
        if extra_body:
            eval_dict.update(extra_body)

        eval_data = CreateProjectRequest(**eval_dict)
        return PreparedRequest(method="POST", url=BASE, data=eval_data.model_dump(exclude_unset=True, mode="json"))

    def prepare_get(self, project_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"{BASE}/{project_id}")

    def prepare_update(
        self,
        project_id: str,
        name: str | _Unset = UNSET,
        json_schema: dict[str, Any] | _Unset = UNSET,
        **extra_body: Any,
    ) -> PreparedRequest:
        update_dict: dict[str, Any] = {}
        if name is not UNSET:
            update_dict["name"] = name
        if json_schema is not UNSET:
            update_dict["json_schema"] = json_schema
        if extra_body:
            update_dict.update(extra_body)

        data = PatchProjectRequest(**update_dict).model_dump(exclude_unset=True, mode="json")
        return PreparedRequest(method="PATCH", url=f"{BASE}/{project_id}", data=data)

    def prepare_list(self, **extra_params: Any) -> PreparedRequest:
        params = extra_params or None
        return PreparedRequest(method="GET", url=BASE, params=params)

    def prepare_delete(self, id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"{BASE}/{id}")

    def prepare_publish(self, project_id: str, origin: Optional[str] = None) -> PreparedRequest:
        params = {"origin": origin} if origin else None
        return PreparedRequest(method="POST", url=f"{BASE}/{project_id}/publish", params=params)

    def prepare_extract(
        self,
        project_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: Sequence[Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl] | None = None,
        model: str | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
    ) -> PreparedRequest:
        if not document and not documents:
            raise ValueError("Either 'document' or 'documents' must be provided")
        if document and documents:
            raise ValueError("Provide either 'document' (single) or 'documents' (multiple), not both")

        form_data = {
            "model": model,
            "image_resolution_dpi": image_resolution_dpi,
            "n_consensus": n_consensus,
            "metadata": json.dumps(metadata) if metadata else None,
            "extraction_id": extraction_id,
        }
        if extra_form:
            form_data.update(extra_form)
        form_data = {k: v for k, v in form_data.items() if v is not None}

        files = {}
        if document:
            mime_document = prepare_mime_document(document)
            files["document"] = (mime_document.filename, base64.b64decode(mime_document.content), mime_document.mime_type)
        elif documents:
            files_list = []
            for doc in documents:
                mime_doc = prepare_mime_document(doc)
                files_list.append(
                    (
                        "documents",
                        (mime_doc.filename, base64.b64decode(mime_doc.content), mime_doc.mime_type),
                    )
                )
            files = files_list

        url = f"{BASE}/extract/{project_id}" if iteration_id is None else f"{BASE}/extract/{project_id}/{iteration_id}"
        return PreparedRequest(method="POST", url=url, form_data=form_data, files=files)


class ExtractProjects(SyncAPIResource, ExtractProjectsMixin):
    """Evals Extract Projects API — manages extraction projects and their evaluation pipelines."""

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.datasets = Datasets(client=client)

    def create(
        self,
        name: str,
        json_schema: dict[str, Any],
        **extra_body: Any,
    ) -> Project:
        request = self.prepare_create(name, json_schema, **extra_body)
        response = self._client._prepared_request(request)
        return Project(**response)

    def get(self, project_id: str) -> Project:
        request = self.prepare_get(project_id)
        response = self._client._prepared_request(request)
        return Project(**response)

    def list(self, **extra_params: Any) -> List[Project]:
        request = self.prepare_list(**extra_params)
        response = self._client._prepared_request(request)
        return [Project(**item) for item in response.get("data", [])]

    def delete(self, project_id: str) -> DeleteResponse:
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
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        **extra_form: Any,
    ) -> RetabParsedChatCompletion:
        request = self.prepare_extract(
            project_id=project_id,
            iteration_id=iteration_id,
            document=document,
            documents=documents,
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            **extra_form,
        )
        response = self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)


class AsyncExtractProjects(AsyncAPIResource, ExtractProjectsMixin):
    """Async Evals Extract Projects API."""

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.datasets = AsyncDatasets(client=client)

    async def create(self, name: str, json_schema: dict[str, Any], **extra_body: Any) -> Project:
        request = self.prepare_create(name, json_schema, **extra_body)
        response = await self._client._prepared_request(request)
        return Project(**response)

    async def get(self, project_id: str) -> Project:
        request = self.prepare_get(project_id)
        response = await self._client._prepared_request(request)
        return Project(**response)

    async def list(self, **extra_params: Any) -> List[Project]:
        request = self.prepare_list(**extra_params)
        response = await self._client._prepared_request(request)
        return [Project(**item) for item in response.get("data", [])]

    async def delete(self, project_id: str) -> DeleteResponse:
        request = self.prepare_delete(project_id)
        return await self._client._prepared_request(request)

    async def publish(self, project_id: str, origin: Optional[str] = None) -> Project:
        """Publish a project's draft configuration."""
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
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        **extra_form: Any,
    ) -> RetabParsedChatCompletion:
        request = self.prepare_extract(project_id=project_id, iteration_id=iteration_id, document=document, documents=documents, model=model, image_resolution_dpi=image_resolution_dpi, n_consensus=n_consensus)
        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)
