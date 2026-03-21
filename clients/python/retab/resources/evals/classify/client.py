from io import IOBase
from pathlib import Path
from typing import Any, Dict

import PIL.Image
from pydantic import HttpUrl

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.documents.classify import ClassifyResponse
from ....types.standards import PreparedRequest
from ....utils.mime import MIMEData
from .._multipart import build_process_request


class ClassifyMixin:
    BASE = "/evals/classify"

    def prepare_process(
        self,
        eval_id: str,
        iteration_id: str | None = None,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str | None = None,
        metadata: Dict[str, str] | None = None,
        **extra_form: Any,
    ) -> PreparedRequest:
        if "documents" in extra_form:
            raise TypeError("client.evals.classify.process(...) accepts only 'document'.")
        return build_process_request(
            base=self.BASE,
            eval_id=eval_id,
            iteration_id=iteration_id,
            document=document,
            model=model,
            metadata=metadata,
            **extra_form,
        )


class Classify(SyncAPIResource, ClassifyMixin):
    """Classify evals process API."""

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


class AsyncClassify(AsyncAPIResource, ClassifyMixin):
    """Async classify evals process API."""

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
