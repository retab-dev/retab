from __future__ import annotations

from datetime import datetime
from io import IOBase
from pathlib import Path
from typing import Any

import PIL.Image
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.mime import MIMEData
from ...types.pagination import PaginatedList, PaginationOrder
from ...types.partitions import Partition, PartitionRequest
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

    def _prepare_get(self, partition_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/partitions/{partition_id}")

    def _prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        filename: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PreparedRequest:
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "filename": filename,
            "from_date": from_date.isoformat() if from_date else None,
            "to_date": to_date.isoformat() if to_date else None,
        }
        params = {key: value for key, value in params.items() if value is not None}
        return PreparedRequest(method="GET", url="/partitions", params=params)

    def _prepare_delete(self, partition_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/partitions/{partition_id}")


class Partitions(SyncAPIResource, PartitionsMixin):
    def create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        key: str,
        instructions: str,
        model: str,
        n_consensus: int = 1,
        bust_cache: bool = False,
    ) -> Partition:
        request = self._prepare_create(
            document=document,
            key=key,
            instructions=instructions,
            model=model,
            n_consensus=n_consensus,
            bust_cache=bust_cache,
        )
        return Partition.model_validate(self._client._prepared_request(request))

    def get(self, partition_id: str) -> Partition:
        """Retrieve a persisted partition resource by id.

        Typically used to fetch the partition referenced by a workflow step's
        ``step.artifact`` (``operation == "partition"``).
        """
        request = self._prepare_get(partition_id)
        return Partition.model_validate(self._client._prepared_request(request))

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        filename: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PaginatedList:
        request = self._prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            filename=filename,
            from_date=from_date,
            to_date=to_date,
        )
        response = self._client._prepared_request(request)
        result = PaginatedList(**response)

        def fetch_next(after_cursor: str) -> PaginatedList:
            return self.list(
                before=before,
                after=after_cursor,
                limit=limit,
                order=order,
                filename=filename,
                from_date=from_date,
                to_date=to_date,
            )

        result._fetch_next_page = fetch_next
        return result

    def delete(self, partition_id: str) -> None:
        request = self._prepare_delete(partition_id)
        self._client._prepared_request(request)


class AsyncPartitions(AsyncAPIResource, PartitionsMixin):
    async def create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        key: str,
        instructions: str,
        model: str,
        n_consensus: int = 1,
        bust_cache: bool = False,
    ) -> Partition:
        request = self._prepare_create(
            document=document,
            key=key,
            instructions=instructions,
            model=model,
            n_consensus=n_consensus,
            bust_cache=bust_cache,
        )
        return Partition.model_validate(await self._client._prepared_request(request))

    async def get(self, partition_id: str) -> Partition:
        """Retrieve a persisted partition resource by id."""
        request = self._prepare_get(partition_id)
        return Partition.model_validate(await self._client._prepared_request(request))

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        filename: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PaginatedList:
        request = self._prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            filename=filename,
            from_date=from_date,
            to_date=to_date,
        )
        response = await self._client._prepared_request(request)
        return PaginatedList(**response)

    async def delete(self, partition_id: str) -> None:
        request = self._prepare_delete(partition_id)
        await self._client._prepared_request(request)
