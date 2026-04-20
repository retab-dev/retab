from __future__ import annotations

from io import IOBase
from pathlib import Path
from typing import Any

import PIL.Image
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.mime import MIMEData
from ...types.partitions import PartitionRequest, PartitionResponse
from ...types.standards import PreparedRequest
from ...utils.mime import prepare_mime_document


class PartitionsMixin:
    def _prepare_create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        key: str,
        instructions: str,
        model: str,
        n_consensus: int = 1,
        bust_cache: bool = False,
    ) -> PreparedRequest:
        request_dict: dict[str, Any] = {
            "document": prepare_mime_document(document),
            "key": key,
            "instructions": instructions,
            "model": model,
            "n_consensus": n_consensus,
        }
        if bust_cache:
            request_dict["bust_cache"] = True

        partition_request = PartitionRequest(**request_dict)
        return PreparedRequest(
            method="POST",
            url="/partitions",
            data=partition_request.model_dump(mode="json", exclude_unset=True),
        )


class Partitions(SyncAPIResource, PartitionsMixin):
    def create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        key: str,
        instructions: str,
        model: str,
        n_consensus: int = 1,
        bust_cache: bool = False,
    ) -> PartitionResponse:
        request = self._prepare_create(
            document=document,
            key=key,
            instructions=instructions,
            model=model,
            n_consensus=n_consensus,
            bust_cache=bust_cache,
        )
        return PartitionResponse.model_validate(self._client._prepared_request(request))


class AsyncPartitions(AsyncAPIResource, PartitionsMixin):
    async def create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        key: str,
        instructions: str,
        model: str,
        n_consensus: int = 1,
        bust_cache: bool = False,
    ) -> PartitionResponse:
        request = self._prepare_create(
            document=document,
            key=key,
            instructions=instructions,
            model=model,
            n_consensus=n_consensus,
            bust_cache=bust_cache,
        )
        return PartitionResponse.model_validate(await self._client._prepared_request(request))
