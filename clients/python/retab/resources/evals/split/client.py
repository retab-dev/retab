from io import IOBase
from pathlib import Path
from typing import Any, Dict

import PIL.Image
from pydantic import HttpUrl

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.documents.extract import RetabParsedChatCompletion
from ....types.standards import PreparedRequest
from ....utils.mime import MIMEData
from ....types.evals.split import CreateSplitProjectRequest, PatchSplitProjectRequest, SplitProject
from .datasets import AsyncDatasets, Datasets
from .templates import AsyncTemplates, Templates
from .._base import AsyncEvalCRUDMixin, SyncEvalCRUDMixin
from .._multipart import build_process_request


class SplitMixin:
    BASE = "/evals/split"

    def prepare_create(
        self,
        name: str,
        split_config: list[dict[str, Any]] | list[Any],
    ) -> PreparedRequest:
        project = CreateSplitProjectRequest(name=name, split_config=split_config)
        return PreparedRequest(method="POST", url=self.BASE, data=project.model_dump(mode="json", exclude_unset=True, by_alias=True))

    def prepare_update(
        self,
        eval_id: str,
        name: str | None = None,
        published_config: Any | None = None,
        draft_config: Any | None = None,
        is_published: bool | None = None,
    ) -> PreparedRequest:
        data = PatchSplitProjectRequest(
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
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
    ) -> PreparedRequest:
        return build_process_request(
            base=self.BASE,
            eval_id=eval_id,
            iteration_id=iteration_id,
            document=document,
            documents=documents,
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )


class Split(SyncAPIResource, SyncEvalCRUDMixin, SplitMixin):
    """Split evals API."""
    PROJECT_MODEL = SplitProject
    CREATE_MODEL = CreateSplitProjectRequest
    PATCH_MODEL = PatchSplitProjectRequest

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.datasets = Datasets(client=client)
        self.templates = Templates(client=client)

    def prepare_create(
        self,
        name: str,
        split_config: list[dict[str, Any]] | list[Any],
    ) -> PreparedRequest:
        return SplitMixin.prepare_create(self, name=name, split_config=split_config)

    def prepare_update(
        self,
        eval_id: str,
        name: str | None = None,
        published_config: Any | None = None,
        draft_config: Any | None = None,
        is_published: bool | None = None,
    ) -> PreparedRequest:
        return SplitMixin.prepare_update(
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
        split_config: list[dict[str, Any]] | list[Any],
    ) -> SplitProject:
        request = self.prepare_create(name=name, split_config=split_config)
        response = self._client._prepared_request(request)
        return SplitProject.model_validate(response)

    def update(
        self,
        eval_id: str,
        name: str | None = None,
        published_config: Any | None = None,
        draft_config: Any | None = None,
        is_published: bool | None = None,
    ) -> SplitProject:
        request = self.prepare_update(
            eval_id=eval_id,
            name=name,
            published_config=published_config,
            draft_config=draft_config,
            is_published=is_published,
        )
        response = self._client._prepared_request(request)
        return SplitProject.model_validate(response)

    def process(
        self,
        eval_id: str,
        iteration_id: str | None = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: None = None,
        model: str | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
    ) -> RetabParsedChatCompletion:
        request = self.prepare_process(
            eval_id=eval_id,
            iteration_id=iteration_id,
            document=document,
            documents=documents,
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )
        response = self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)


class AsyncSplit(AsyncAPIResource, AsyncEvalCRUDMixin, SplitMixin):
    """Async split evals API."""
    PROJECT_MODEL = SplitProject
    CREATE_MODEL = CreateSplitProjectRequest
    PATCH_MODEL = PatchSplitProjectRequest

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.datasets = AsyncDatasets(client=client)
        self.templates = AsyncTemplates(client=client)

    def prepare_create(
        self,
        name: str,
        split_config: list[dict[str, Any]] | list[Any],
    ) -> PreparedRequest:
        return SplitMixin.prepare_create(self, name=name, split_config=split_config)

    def prepare_update(
        self,
        eval_id: str,
        name: str | None = None,
        published_config: Any | None = None,
        draft_config: Any | None = None,
        is_published: bool | None = None,
    ) -> PreparedRequest:
        return SplitMixin.prepare_update(
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
        split_config: list[dict[str, Any]] | list[Any],
    ) -> SplitProject:
        request = self.prepare_create(name=name, split_config=split_config)
        response = await self._client._prepared_request(request)
        return SplitProject.model_validate(response)

    async def update(
        self,
        eval_id: str,
        name: str | None = None,
        published_config: Any | None = None,
        draft_config: Any | None = None,
        is_published: bool | None = None,
    ) -> SplitProject:
        request = self.prepare_update(
            eval_id=eval_id,
            name=name,
            published_config=published_config,
            draft_config=draft_config,
            is_published=is_published,
        )
        response = await self._client._prepared_request(request)
        return SplitProject.model_validate(response)

    async def process(
        self,
        eval_id: str,
        iteration_id: str | None = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: None = None,
        model: str | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
    ) -> RetabParsedChatCompletion:
        request = self.prepare_process(
            eval_id=eval_id,
            iteration_id=iteration_id,
            document=document,
            documents=documents,
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )
        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)
