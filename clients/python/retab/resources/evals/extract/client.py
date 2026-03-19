import warnings
from io import IOBase
from pathlib import Path
from typing import Any, Dict, Optional

import PIL.Image
from pydantic import HttpUrl
from ...._resource import AsyncAPIResource, SyncAPIResource
from ....utils.mime import MIMEData
from ....utils.stream_context_managers import as_async_context_manager, as_context_manager
from ....types.documents.extract import RetabParsedChatCompletion, RetabParsedChatCompletionChunk
from ....types.standards import PreparedRequest
from .._base import AsyncEvalCRUDMixin, SyncEvalCRUDMixin
from .._multipart import build_process_request
from .datasets import Datasets, AsyncDatasets
from .templates import AsyncTemplates, Templates

BASE = "/evals/extract"


class ExtractMixin:
    BASE = BASE

    def prepare_process(
        self,
        eval_id: str,
        iteration_id: str | None = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
    ) -> PreparedRequest:
        if "documents" in extra_form:
            raise TypeError("client.evals.extract.process(...) accepts only 'document'.")
        return build_process_request(
            base=self.BASE,
            eval_id=eval_id,
            iteration_id=iteration_id,
            document=document,
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )

    def prepare_extract(
        self,
        project_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
        ) -> PreparedRequest:
        return self.prepare_process(
            eval_id=project_id,
            iteration_id=iteration_id,
            document=document,
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )

    def prepare_process_stream(
        self,
        eval_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
    ) -> PreparedRequest:
        if "documents" in extra_form:
            raise TypeError("client.evals.extract.process_stream(...) accepts only 'document'.")
        request = self.prepare_process(
            eval_id=eval_id,
            iteration_id=iteration_id,
            document=document,
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )
        return request.model_copy(update={"url": f"{request.url}/stream"})


class Extract(SyncAPIResource, SyncEvalCRUDMixin, ExtractMixin):
    """Extract evals API."""

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.datasets = Datasets(client=client)
        self.templates = Templates(client=client)

    def process(
        self,
        eval_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
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
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )
        response = self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)

    def extract(
        self,
        project_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
        ) -> RetabParsedChatCompletion:
        warnings.warn(
            "client.evals.extract.extract(...) is deprecated; use client.evals.extract.process(...) instead.",
            DeprecationWarning,
            stacklevel=2,
        )
        return self.process(
            eval_id=project_id,
            iteration_id=iteration_id,
            document=document,
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )

    @as_context_manager
    def process_stream(
        self,
        eval_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
    ):
        request = self.prepare_process_stream(
            eval_id=eval_id,
            iteration_id=iteration_id,
            document=document,
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )
        for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            yield RetabParsedChatCompletionChunk.model_validate(chunk_json)


class AsyncExtract(AsyncAPIResource, AsyncEvalCRUDMixin, ExtractMixin):
    """Async extract evals API."""

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.datasets = AsyncDatasets(client=client)
        self.templates = AsyncTemplates(client=client)

    async def process(
        self,
        eval_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
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
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )
        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)

    async def extract(
        self,
        project_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
        ) -> RetabParsedChatCompletion:
        warnings.warn(
            "client.evals.extract.extract(...) is deprecated; use client.evals.extract.process(...) instead.",
            DeprecationWarning,
            stacklevel=2,
        )
        return await self.process(
            eval_id=project_id,
            iteration_id=iteration_id,
            document=document,
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )

    @as_async_context_manager
    async def process_stream(
        self,
        eval_id: str,
        iteration_id: Optional[str] = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str | None = None,
        image_resolution_dpi: int | None = None,
        n_consensus: int | None = None,
        metadata: Dict[str, str] | None = None,
        extraction_id: str | None = None,
        **extra_form: Any,
    ):
        request = self.prepare_process_stream(
            eval_id=eval_id,
            iteration_id=iteration_id,
            document=document,
            model=model,
            image_resolution_dpi=image_resolution_dpi,
            n_consensus=n_consensus,
            metadata=metadata,
            extraction_id=extraction_id,
            **extra_form,
        )
        async for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            yield RetabParsedChatCompletionChunk.model_validate(chunk_json)


ExtractProjects = Extract
AsyncExtractProjects = AsyncExtract
