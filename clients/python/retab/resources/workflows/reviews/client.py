"""Actor-neutral client for workflow reviews."""

from __future__ import annotations

import asyncio
import time
from typing import Any, Dict, Literal

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....exceptions import NotFoundError
from ....types.standards import PreparedRequest
from ....types.workflows.reviews import (
    AppendVersionResponse,
    Review,
    ReviewQueueResponse,
    SubmitDecisionResponse,
)

ReviewDecisionStatus = Literal["pending", "approved", "rejected", "decided", "all"]


class WorkflowReviewsMixin:
    """Mixin providing shared prepare methods for workflow review operations."""

    def prepare_list(
        self,
        workflow_id: str | None = None,
        run_id: str | None = None,
        block_id: str | None = None,
        step_id: str | None = None,
        iteration_key: str | None = None,
        limit: int = 50,
        decision_status: ReviewDecisionStatus = "pending",
        before: str | None = None,
        after: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to list review summaries."""
        params: Dict[str, Any] = {"limit": limit, "decision_status": decision_status}
        if workflow_id is not None:
            params["workflow_id"] = workflow_id
        if run_id is not None:
            params["run_id"] = run_id
        if block_id is not None:
            params["block_id"] = block_id
        if step_id is not None:
            params["step_id"] = step_id
        if iteration_key is not None:
            params["iteration_key"] = iteration_key
        if before is not None:
            params["before"] = before
        if after is not None:
            params["after"] = after
        return PreparedRequest(method="GET", url="/workflows/reviews", params=params)

    def prepare_get(self, review_id: str) -> PreparedRequest:
        """Prepare a request to fetch one review."""
        return PreparedRequest(method="GET", url=f"/workflows/reviews/{review_id}")

    def prepare_append_version(
        self,
        review_id: str,
        *,
        snapshot: dict,
        parent_version_id: str,
        note: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to append a new output version to the review."""
        data: Dict[str, Any] = {
            "snapshot": snapshot,
            "parent_version_id": parent_version_id,
        }
        if note is not None:
            data["note"] = note
        return PreparedRequest(
            method="POST",
            url=f"/workflows/reviews/{review_id}/versions",
            data=data,
        )

    def prepare_approve(self, review_id: str, *, version_id: str) -> PreparedRequest:
        """Prepare a request to approve one version."""
        return PreparedRequest(
            method="POST",
            url=f"/workflows/reviews/{review_id}/approve",
            data={"version_id": version_id},
        )

    def prepare_reject(self, review_id: str, *, version_id: str, reason: str) -> PreparedRequest:
        """Prepare a request to reject one version."""
        return PreparedRequest(
            method="POST",
            url=f"/workflows/reviews/{review_id}/reject",
            data={"version_id": version_id, "reason": reason},
        )


class WorkflowReviews(SyncAPIResource, WorkflowReviewsMixin):
    """Workflow reviews API wrapper for synchronous operations."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def list(
        self,
        workflow_id: str | None = None,
        run_id: str | None = None,
        block_id: str | None = None,
        step_id: str | None = None,
        iteration_key: str | None = None,
        limit: int = 50,
        decision_status: ReviewDecisionStatus = "pending",
        before: str | None = None,
        after: str | None = None,
    ) -> ReviewQueueResponse:
        """List review summaries."""
        request = self.prepare_list(
            workflow_id=workflow_id,
            run_id=run_id,
            block_id=block_id,
            step_id=step_id,
            iteration_key=iteration_key,
            limit=limit,
            decision_status=decision_status,
            before=before,
            after=after,
        )
        response = self._client._prepared_request(request)
        return ReviewQueueResponse.model_validate(response)

    def get(self, review_id: str) -> Review:
        """Fetch the full review by id."""
        request = self.prepare_get(review_id)
        response = self._client._prepared_request(request)
        return Review.model_validate(response)

    def approve(self, review_id: str, *, version_id: str) -> SubmitDecisionResponse:
        """Approve one exact output version."""
        request = self.prepare_approve(review_id, version_id=version_id)
        response = self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    def reject(self, review_id: str, *, version_id: str, reason: str) -> SubmitDecisionResponse:
        """Reject one exact output version."""
        request = self.prepare_reject(review_id, version_id=version_id, reason=reason)
        response = self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    def append_version(
        self,
        review_id: str,
        *,
        snapshot: dict,
        parent_version_id: str,
        note: str | None = None,
    ) -> AppendVersionResponse:
        """Append a new output version to the review's version graph."""
        request = self.prepare_append_version(
            review_id,
            snapshot=snapshot,
            parent_version_id=parent_version_id,
            note=note,
        )
        response = self._client._prepared_request(request)
        return AppendVersionResponse.model_validate(response)

    def wait_for(
        self,
        review_id: str,
        *,
        timeout: float = 120.0,
        poll_interval: float = 2.0,
    ) -> Review:
        """Poll until the review exists and is pending."""
        deadline = time.monotonic() + timeout
        while True:
            try:
                review = self.get(review_id)
                if review.decision is None:
                    return review
            except NotFoundError:
                pass
            if time.monotonic() >= deadline:
                raise TimeoutError(f"Review {review_id!r} was not pending within {timeout}s")
            time.sleep(poll_interval)


class AsyncWorkflowReviews(AsyncAPIResource, WorkflowReviewsMixin):
    """Workflow reviews API wrapper for asynchronous operations."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def list(
        self,
        workflow_id: str | None = None,
        run_id: str | None = None,
        block_id: str | None = None,
        step_id: str | None = None,
        iteration_key: str | None = None,
        limit: int = 50,
        decision_status: ReviewDecisionStatus = "pending",
        before: str | None = None,
        after: str | None = None,
    ) -> ReviewQueueResponse:
        """List review summaries."""
        request = self.prepare_list(
            workflow_id=workflow_id,
            run_id=run_id,
            block_id=block_id,
            step_id=step_id,
            iteration_key=iteration_key,
            limit=limit,
            decision_status=decision_status,
            before=before,
            after=after,
        )
        response = await self._client._prepared_request(request)
        return ReviewQueueResponse.model_validate(response)

    async def get(self, review_id: str) -> Review:
        """Fetch the full review by id."""
        request = self.prepare_get(review_id)
        response = await self._client._prepared_request(request)
        return Review.model_validate(response)

    async def approve(self, review_id: str, *, version_id: str) -> SubmitDecisionResponse:
        """Approve one exact output version."""
        request = self.prepare_approve(review_id, version_id=version_id)
        response = await self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    async def reject(self, review_id: str, *, version_id: str, reason: str) -> SubmitDecisionResponse:
        """Reject one exact output version."""
        request = self.prepare_reject(review_id, version_id=version_id, reason=reason)
        response = await self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    async def append_version(
        self,
        review_id: str,
        *,
        snapshot: dict,
        parent_version_id: str,
        note: str | None = None,
    ) -> AppendVersionResponse:
        """Append a new output version to the review's version graph."""
        request = self.prepare_append_version(
            review_id,
            snapshot=snapshot,
            parent_version_id=parent_version_id,
            note=note,
        )
        response = await self._client._prepared_request(request)
        return AppendVersionResponse.model_validate(response)

    async def wait_for(
        self,
        review_id: str,
        *,
        timeout: float = 120.0,
        poll_interval: float = 2.0,
    ) -> Review:
        """Poll until the review exists and is pending."""
        deadline = time.monotonic() + timeout
        while True:
            try:
                review = await self.get(review_id)
                if review.decision is None:
                    return review
            except NotFoundError:
                pass
            if time.monotonic() >= deadline:
                raise TimeoutError(f"Review {review_id!r} was not pending within {timeout}s")
            await asyncio.sleep(poll_interval)
