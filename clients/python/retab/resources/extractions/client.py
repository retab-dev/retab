import json
from datetime import datetime
from typing import Any, Dict

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.extractions import Extraction, ExtractionRequest, SourcesResponse
from ...types.pagination import PaginatedList, PaginationOrder
from ...types.standards import PreparedRequest


class ExtractionsMixin:
    def prepare_create(self, payload: ExtractionRequest) -> PreparedRequest:
        return PreparedRequest(method="POST", url="/extractions", data=payload.model_dump(mode="json", exclude_none=True))

    def prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        origin_type: str | None = None,
        origin_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
        metadata: Dict[str, str] | None = None,
        **extra_params: Any,
    ) -> PreparedRequest:
        """Prepare a request to list extractions with pagination and filtering."""
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "origin_type": origin_type,
            "origin_id": origin_id,
            "from_date": from_date.isoformat() if from_date else None,
            "to_date": to_date.isoformat() if to_date else None,
            # Note: metadata must be JSON-serialized as the backend expects a JSON string
            "metadata": json.dumps(metadata) if metadata else None,
        }
        if extra_params:
            params.update(extra_params)
        params = {k: v for k, v in params.items() if v is not None}
        return PreparedRequest(method="GET", url="/extractions", params=params)

    def prepare_get(self, extraction_id: str) -> PreparedRequest:
        """Prepare a request to get an extraction by ID."""
        return PreparedRequest(method="GET", url=f"/extractions/{extraction_id}")

    def prepare_sources(self, extraction_id: str) -> PreparedRequest:
        """Prepare a request to get sourced extraction with per-leaf provenance."""
        return PreparedRequest(method="GET", url=f"/extractions/{extraction_id}/sources")

    def prepare_delete(self, extraction_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/extractions/{extraction_id}")

class Extractions(SyncAPIResource, ExtractionsMixin):
    """Extractions API wrapper."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        origin_type: str | None = None,
        origin_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
        metadata: Dict[str, str] | None = None,
        **extra_params: Any,
    ) -> PaginatedList:
        """List extractions with pagination and filtering."""
        request = self.prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            origin_type=origin_type,
            origin_id=origin_id,
            from_date=from_date,
            to_date=to_date,
            metadata=metadata,
            **extra_params,
        )
        response = self._client._prepared_request(request)
        result = PaginatedList(**response)

        # Enable auto-pagination
        def fetch_next(after: str) -> PaginatedList:
            return self.list(
                before=before,
                after=after,
                limit=limit,
                order=order,
                origin_type=origin_type,
                origin_id=origin_id,
                from_date=from_date,
                to_date=to_date,
                metadata=metadata,
                **extra_params,
            )

        result._fetch_next_page = fetch_next
        return result

    def get(self, extraction_id: str) -> Extraction:
        """Get an extraction by ID."""
        request = self.prepare_get(extraction_id)
        return Extraction.model_validate(self._client._prepared_request(request))

    def create(self, payload: ExtractionRequest) -> Extraction:
        """Create an extraction using the modern /v1/extractions endpoint."""
        request = self.prepare_create(payload)
        return Extraction.model_validate(self._client._prepared_request(request))

    def sources(self, extraction_id: str) -> SourcesResponse:
        """Get extraction result enriched with per-leaf source provenance.

        Args:
            extraction_id: ID of the extraction to source.

        Returns:
            SourcesResponse with `extraction` (original) and
            `sources` (leaves wrapped as {value, source}).
        """
        request = self.prepare_sources(extraction_id)
        return SourcesResponse.model_validate(self._client._prepared_request(request))

    def delete(self, extraction_id: str) -> None:
        request = self.prepare_delete(extraction_id)
        self._client._prepared_request(request)

class AsyncExtractions(AsyncAPIResource, ExtractionsMixin):
    """Async Extractions API wrapper."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        origin_type: str | None = None,
        origin_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
        metadata: Dict[str, str] | None = None,
        **extra_params: Any,
    ) -> PaginatedList:
        """List extractions with pagination and filtering."""
        request = self.prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            origin_type=origin_type,
            origin_id=origin_id,
            from_date=from_date,
            to_date=to_date,
            metadata=metadata,
            **extra_params,
        )
        response = await self._client._prepared_request(request)
        return PaginatedList(**response)

    async def get(self, extraction_id: str) -> Extraction:
        """Get an extraction by ID."""
        request = self.prepare_get(extraction_id)
        return Extraction.model_validate(await self._client._prepared_request(request))

    async def create(self, payload: ExtractionRequest) -> Extraction:
        """Create an extraction using the modern /v1/extractions endpoint."""
        request = self.prepare_create(payload)
        return Extraction.model_validate(await self._client._prepared_request(request))

    async def sources(self, extraction_id: str) -> SourcesResponse:
        """Get extraction result enriched with per-leaf source provenance.

        Args:
            extraction_id: ID of the extraction to source.

        Returns:
            SourcesResponse with `extraction` (original) and
            `sources` (leaves wrapped as {value, source}).
        """
        request = self.prepare_sources(extraction_id)
        return SourcesResponse.model_validate(await self._client._prepared_request(request))

    async def delete(self, extraction_id: str) -> None:
        request = self.prepare_delete(extraction_id)
        await self._client._prepared_request(request)
