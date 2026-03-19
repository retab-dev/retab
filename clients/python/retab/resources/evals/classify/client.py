from io import IOBase
from pathlib import Path
from typing import Any, Dict

import PIL.Image
from pydantic import HttpUrl

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.documents.classify import ClassifyResponse
from ....types.standards import PreparedRequest
from ....utils.mime import MIMEData
from ....types.evals.classify import ClassifyProject, CreateClassifyProjectRequest, PatchClassifyProjectRequest
from .datasets import AsyncDatasets, Datasets
from .templates import AsyncTemplates, Templates
from .._base import AsyncEvalCRUDMixin, SyncEvalCRUDMixin
from .._multipart import build_process_request


class ClassifyMixin:
    BASE = "/evals/classify"

    def prepare_create(
        self,
        name: str,
        categories: list[Any],
    ) -> PreparedRequest:
        project = CreateClassifyProjectRequest(name=name, categories=categories)
        return PreparedRequest(method="POST", url=self.BASE, data=project.model_dump(mode="json", exclude_unset=True))

    def prepare_update(
        self,
        eval_id: str,
        name: str | None = None,
        published_config: Any | None = None,
        draft_config: Any | None = None,
        is_published: bool | None = None,
    ) -> PreparedRequest:
        data = PatchClassifyProjectRequest(
            name=name,
            published_config=published_config,
            draft_config=draft_config,
            is_published=is_published,
        ).model_dump(mode="json", exclude_unset=True, exclude_none=True)
        return PreparedRequest(method="PATCH", url=f"{self.BASE}/{eval_id}", data=data)

    def prepare_process(
        self,
        eval_id: str,
        iteration_id: str | None = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: None = None,
        model: str | None = None,
        metadata: Dict[str, str] | None = None,
        **extra_form: Any,
    ) -> PreparedRequest:
        return build_process_request(
            base=self.BASE,
            eval_id=eval_id,
            iteration_id=iteration_id,
            document=document,
            documents=documents,
            model=model,
            metadata=metadata,
            **extra_form,
        )


class Classify(SyncAPIResource, SyncEvalCRUDMixin, ClassifyMixin):
    """Classify evals API."""
    PROJECT_MODEL = ClassifyProject
    CREATE_MODEL = CreateClassifyProjectRequest
    PATCH_MODEL = PatchClassifyProjectRequest

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.datasets = Datasets(client=client)
        self.templates = Templates(client=client)

    def prepare_create(
        self,
        name: str,
        categories: list[Any],
    ) -> PreparedRequest:
        return ClassifyMixin.prepare_create(self, name=name, categories=categories)

    def prepare_update(
        self,
        eval_id: str,
        name: str | None = None,
        published_config: Any | None = None,
        draft_config: Any | None = None,
        is_published: bool | None = None,
    ) -> PreparedRequest:
        return ClassifyMixin.prepare_update(
            self,
            eval_id=eval_id,
            name=name,
            published_config=published_config,
            draft_config=draft_config,
            is_published=is_published,
        )

    def create(
        self,
        name: str,
        categories: list[Any],
    ) -> ClassifyProject:
        request = self.prepare_create(name=name, categories=categories)
        response = self._client._prepared_request(request)
        return ClassifyProject.model_validate(response)

    def update(
        self,
        eval_id: str,
        name: str | None = None,
        published_config: Any | None = None,
        draft_config: Any | None = None,
        is_published: bool | None = None,
    ) -> ClassifyProject:
        request = self.prepare_update(
            eval_id=eval_id,
            name=name,
            published_config=published_config,
            draft_config=draft_config,
            is_published=is_published,
        )
        response = self._client._prepared_request(request)
        return ClassifyProject.model_validate(response)

    def process(
        self,
        eval_id: str,
        iteration_id: str | None = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str | None = None,
        metadata: Dict[str, str] | None = None,
        **extra_form: Any,
    ) -> ClassifyResponse:
        request = self.prepare_process(
            eval_id=eval_id,
            iteration_id=iteration_id,
            document=document,
            model=model,
            metadata=metadata,
            **extra_form,
        )
        response = self._client._prepared_request(request)
        return ClassifyResponse.model_validate(response)


class AsyncClassify(AsyncAPIResource, AsyncEvalCRUDMixin, ClassifyMixin):
    """Async classify evals API."""
    PROJECT_MODEL = ClassifyProject
    CREATE_MODEL = CreateClassifyProjectRequest
    PATCH_MODEL = PatchClassifyProjectRequest

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.datasets = AsyncDatasets(client=client)
        self.templates = AsyncTemplates(client=client)

    def prepare_create(
        self,
        name: str,
        categories: list[Any],
    ) -> PreparedRequest:
        return ClassifyMixin.prepare_create(self, name=name, categories=categories)

    def prepare_update(
        self,
        eval_id: str,
        name: str | None = None,
        published_config: Any | None = None,
        draft_config: Any | None = None,
        is_published: bool | None = None,
    ) -> PreparedRequest:
        return ClassifyMixin.prepare_update(
            self,
            eval_id=eval_id,
            name=name,
            published_config=published_config,
            draft_config=draft_config,
            is_published=is_published,
        )

    async def create(
        self,
        name: str,
        categories: list[Any],
    ) -> ClassifyProject:
        request = self.prepare_create(name=name, categories=categories)
        response = await self._client._prepared_request(request)
        return ClassifyProject.model_validate(response)

    async def update(
        self,
        eval_id: str,
        name: str | None = None,
        published_config: Any | None = None,
        draft_config: Any | None = None,
        is_published: bool | None = None,
    ) -> ClassifyProject:
        request = self.prepare_update(
            eval_id=eval_id,
            name=name,
            published_config=published_config,
            draft_config=draft_config,
            is_published=is_published,
        )
        response = await self._client._prepared_request(request)
        return ClassifyProject.model_validate(response)

    async def process(
        self,
        eval_id: str,
        iteration_id: str | None = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str | None = None,
        metadata: Dict[str, str] | None = None,
        **extra_form: Any,
    ) -> ClassifyResponse:
        request = self.prepare_process(
            eval_id=eval_id,
            iteration_id=iteration_id,
            document=document,
            model=model,
            metadata=metadata,
            **extra_form,
        )
        response = await self._client._prepared_request(request)
        return ClassifyResponse.model_validate(response)
