import json
from datetime import datetime
from typing import Any, Dict, List, Literal

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.standards import PreparedRequest
from ...types.pagination import PaginatedList, PaginationOrder
from ...types.extractions.types import HumanReviewStatus

class ExtractionsMixin:
    def prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        origin_dot_type: str | None = None,
        origin_dot_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None, 
        human_review_status: HumanReviewStatus | None = None,
        metadata: Dict[str, str] | None = None,
        **extra_params: Any,
    ) -> PreparedRequest:
        """Prepare a request to list extractions with pagination and filtering."""
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "origin_dot_type": origin_dot_type,
            "origin_dot_id": origin_dot_id,
            "from_date": from_date.isoformat() if from_date else None,
            "to_date": to_date.isoformat() if to_date else None,
            "human_review_status": human_review_status,
            # Note: metadata must be JSON-serialized as the backend expects a JSON string
            "metadata": json.dumps(metadata) if metadata else None,
        }
        if extra_params:
            params.update(extra_params)
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}
        return PreparedRequest(method="GET", url="/v1/extractions", params=params)

    def prepare_download(
        self,
        order: Literal["asc", "desc"] = "desc",
        origin_dot_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
        human_review_status: HumanReviewStatus | None = None,
        metadata: Dict[str, str] | None = None,
        format: Literal["jsonl", "csv", "xlsx"] = "jsonl",
        **extra_params: Any,
    ) -> PreparedRequest:
        """Prepare a request to download extractions in various formats."""
        params = {
            "order": order,
            "origin_dot_id": origin_dot_id,
            "from_date": from_date.isoformat() if from_date else None,
            "to_date": to_date.isoformat() if to_date else None,
            "human_review_status": human_review_status,
            # Note: metadata must be JSON-serialized as the backend expects a JSON string
            "metadata": json.dumps(metadata) if metadata else None,
            "format": format,
        }
        if extra_params:
            params.update(extra_params)
        params = {k: v for k, v in params.items() if v is not None}
        return PreparedRequest(method="GET", url="/v1/extractions/download", params=params)

    def prepare_update(
        self,
        extraction_id: str,
        predictions: dict[str, Any] | None = None,
        human_review_status: HumanReviewStatus | None = None,
        json_schema: dict[str, Any] | None = None,
        inference_settings: dict[str, Any] | None = None,
        **extra_body: Any,
    ) -> PreparedRequest:
        """Prepare a request to update an extraction."""
        data: dict[str, Any] = {}
        if predictions is not None:
            data["predictions"] = predictions
        if human_review_status is not None:
            data["human_review_status"] = human_review_status
        if json_schema is not None:
            data["json_schema"] = json_schema
        if inference_settings is not None:
            data["inference_settings"] = inference_settings
        if extra_body:
            data.update(extra_body)
        return PreparedRequest(method="PATCH", url=f"/v1/extractions/{extraction_id}", data=data)

    def prepare_get(self, extraction_id: str) -> PreparedRequest:
        """Prepare a request to get an extraction by ID."""
        return PreparedRequest(method="GET", url=f"/v1/extractions/{extraction_id}")

    def prepare_delete(self, extraction_id: str) -> PreparedRequest:
        """Prepare a request to delete an extraction by ID."""
        return PreparedRequest(method="DELETE", url=f"/v1/extractions/{extraction_id}")

class Extractions(SyncAPIResource, ExtractionsMixin):
    """Extractions API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        origin_dot_type: str | None = None,
        origin_dot_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
        human_review_status: HumanReviewStatus | None = None,
        metadata: Dict[str, str] | None = None,
        **extra_params: Any,
    ) -> PaginatedList:
        """List extractions with pagination and filtering."""
        request = self.prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            origin_dot_type=origin_dot_type,
            origin_dot_id=origin_dot_id,
            from_date=from_date,
            to_date=to_date,
            human_review_status=human_review_status,
            metadata=metadata,
            **extra_params,
        )
        response = self._client._prepared_request(request)
        return PaginatedList(**response)

    
    def download(
        self,
        order: Literal["asc", "desc"] = "desc",
        origin_dot_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
        human_review_status: HumanReviewStatus | None = None,
        metadata: Dict[str, str] | None = None,
        format: Literal["jsonl", "csv", "xlsx"] = "jsonl",
        **extra_params: Any,
    ) -> dict[str, Any]:
        """Download extractions in various formats. Returns download_url, filename, and expires_at."""
        request = self.prepare_download(
            order=order,
            origin_dot_id=origin_dot_id,
            from_date=from_date,
            to_date=to_date,
            human_review_status=human_review_status,
            metadata=metadata,
            format=format,
            **extra_params,
        )
        return self._client._prepared_request(request)

   
    def update(
        self,
        extraction_id: str,
        predictions: dict[str, Any] | None = None,
        human_review_status: HumanReviewStatus | None = None,
        json_schema: dict[str, Any] | None = None,
        inference_settings: dict[str, Any] | None = None,
        **extra_body: Any,
    ) -> dict[str, Any]:
        """Update an extraction."""
        request = self.prepare_update(
            extraction_id=extraction_id,
            predictions=predictions,
            human_review_status=human_review_status,
            json_schema=json_schema,
            inference_settings=inference_settings,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return response

    def get(self, extraction_id: str) -> dict[str, Any]:
        """Get an extraction by ID."""
        request = self.prepare_get(extraction_id)
        return self._client._prepared_request(request)

    def delete(self, extraction_id: str) -> None:
        """Delete an extraction by ID."""
        request = self.prepare_delete(extraction_id)
        self._client._prepared_request(request)


class AsyncExtractions(AsyncAPIResource, ExtractionsMixin):
    """Async Extractions API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        origin_dot_type: str | None = None,
        origin_dot_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
        human_review_status: HumanReviewStatus | None = None,
        metadata: Dict[str, str] | None = None,
        **extra_params: Any,
    ) -> PaginatedList:
        """List extractions with pagination and filtering."""
        request = self.prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            origin_dot_type=origin_dot_type,
            origin_dot_id=origin_dot_id,
            from_date=from_date,
            to_date=to_date,
            human_review_status=human_review_status,
            metadata=metadata,
            **extra_params,
        )
        response = await self._client._prepared_request(request)
        return PaginatedList(**response)

    async def download(
        self,
        order: Literal["asc", "desc"] = "desc",
        origin_dot_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
        human_review_status: HumanReviewStatus | None = None,
        metadata: Dict[str, str] | None = None,
        format: Literal["jsonl", "csv", "xlsx"] = "jsonl",
        **extra_params: Any,
    ) -> dict[str, Any]:
        """Download extractions in various formats. Returns download_url, filename, and expires_at."""
        request = self.prepare_download(
            order=order,
            origin_dot_id=origin_dot_id,
            from_date=from_date,
            to_date=to_date,
            human_review_status=human_review_status,
            metadata=metadata,
            format=format,
            **extra_params,
        )
        return await self._client._prepared_request(request)

    async def update(
        self,
        extraction_id: str,
        predictions: dict[str, Any] | None = None,
        human_review_status: HumanReviewStatus | None = None,
        json_schema: dict[str, Any] | None = None,
        inference_settings: dict[str, Any] | None = None,
        **extra_body: Any,
    ) -> dict[str, Any]:
        """Update an extraction."""
        request = self.prepare_update(
            extraction_id=extraction_id,
            predictions=predictions,
            human_review_status=human_review_status,
            json_schema=json_schema,
            inference_settings=inference_settings,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return response

    async def get(self, extraction_id: str) -> dict[str, Any]:
        """Get an extraction by ID."""
        request = self.prepare_get(extraction_id)
        return await self._client._prepared_request(request)

    async def delete(self, extraction_id: str) -> None:
        """Delete an extraction by ID."""
        request = self.prepare_delete(extraction_id)
        await self._client._prepared_request(request)
