from io import IOBase
from pathlib import Path
from typing import Any, Dict

import PIL.Image
from pydantic import HttpUrl

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.documents.extract import RetabParsedChatCompletion
from ....types.standards import PreparedRequest
from ....utils.mime import MIMEData
from .._multipart import build_process_request


class SplitMixin:
    BASE = "/evals/split"

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
            raise TypeError("client.evals.split.process(...) accepts only 'document'.")
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


class Split(SyncAPIResource, SplitMixin):
    """Split evals process API."""

    def process(
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


class AsyncSplit(AsyncAPIResource, SplitMixin):
    """Async split evals process API."""

    async def process(
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
