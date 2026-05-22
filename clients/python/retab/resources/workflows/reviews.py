"""Actor-neutral client for workflow reviews."""

from __future__ import annotations

from typing import Any, Dict

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.pagination import PaginatedList
from ...types.standards import PreparedRequest
from ...types.workflows.reviews import (
    Review,
    ReviewDecisionStatus,
    ReviewVersion,
    ReviewVersionListResponse,
    SubmitDecisionResponse,
)

__all__ = [
    # Re-export so `from retab.resources.workflows.reviews.client
    # import ReviewDecisionStatus` keeps working for any existing call
    # site that imported it from this module before the type was moved
    # to its idiomatic home at `retab.types.workflows.reviews`.
    "ReviewDecisionStatus",
    "WorkflowReviews",
    "AsyncWorkflowReviews",
    "WorkflowReviewsMixin",
]


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


class WorkflowReviewVersionsMixin:
    """Mixin providing shared prepare methods for review version operations."""

    def prepare_create(
        self,
        *,
        review_id: str,
        snapshot: dict,
        parent_id: str,
        note: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to create one immutable review version."""
        if not parent_id:
            raise ValueError("parent_id is required when creating a review version")
        data: Dict[str, Any] = {
            "review_id": review_id,
            "snapshot": snapshot,
            "parent_id": parent_id,
        }
        if note is not None:
            data["note"] = note
        return PreparedRequest(
            method="POST",
            url="/workflows/reviews/versions",
            data=data,
        )

    def prepare_get(self, version_id: str) -> PreparedRequest:
        """Prepare a request to fetch one review version."""
        return PreparedRequest(
            method="GET",
            url=f"/workflows/reviews/versions/{version_id}",
        )

    def prepare_list(
        self,
        *,
        review_id: str,
        limit: int = 50,
        before: str | None = None,
        after: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to list versions for one review."""
        params: Dict[str, Any] = {"review_id": review_id, "limit": limit}
        if before is not None:
            params["before"] = before
        if after is not None:
            params["after"] = after
        return PreparedRequest(
            method="GET",
            url="/workflows/reviews/versions",
            params=params,
        )


def _parse_review_version_response(response: Any) -> ReviewVersion:
    return ReviewVersion.model_validate(response)


class WorkflowReviewVersions(SyncAPIResource, WorkflowReviewVersionsMixin):
    """Workflow review versions API wrapper for synchronous operations."""

    def create(
        self,
        *,
        review_id: str,
        snapshot: dict,
        parent_id: str,
        note: str | None = None,
    ) -> ReviewVersion:
        """Create one immutable correction version of a review.

        ``parent_id`` is required — every SDK-created version is a
        correction of an existing version under ``review_id``. Seed
        versions (``parent_id is None``) are created by the workflow
        runtime only; the SDK does not expose that path.
        """
        request = self.prepare_create(
            review_id=review_id,
            snapshot=snapshot,
            parent_id=parent_id,
            note=note,
        )
        response = self._client._prepared_request(request)
        return _parse_review_version_response(response)

    def get(self, version_id: str) -> ReviewVersion:
        """Fetch one review version by id."""
        request = self.prepare_get(version_id)
        response = self._client._prepared_request(request)
        return ReviewVersion.model_validate(response)

    def list(
        self,
        *,
        review_id: str,
        limit: int = 50,
        before: str | None = None,
        after: str | None = None,
    ) -> ReviewVersionListResponse:
        """List versions for one review."""
        request = self.prepare_list(
            review_id=review_id,
            limit=limit,
            before=before,
            after=after,
        )
        response = self._client._prepared_request(request)
        return ReviewVersionListResponse.model_validate(response)


class AsyncWorkflowReviewVersions(AsyncAPIResource, WorkflowReviewVersionsMixin):
    """Workflow review versions API wrapper for asynchronous operations."""

    async def create(
        self,
        *,
        review_id: str,
        snapshot: dict,
        parent_id: str,
        note: str | None = None,
    ) -> ReviewVersion:
        """Create one immutable correction version of a review.

        ``parent_id`` is required — every SDK-created version is a
        correction of an existing version under ``review_id``. Seed
        versions (``parent_id is None``) are created by the workflow
        runtime only; the SDK does not expose that path.
        """
        request = self.prepare_create(
            review_id=review_id,
            snapshot=snapshot,
            parent_id=parent_id,
            note=note,
        )
        response = await self._client._prepared_request(request)
        return _parse_review_version_response(response)

    async def get(self, version_id: str) -> ReviewVersion:
        """Fetch one review version by id."""
        request = self.prepare_get(version_id)
        response = await self._client._prepared_request(request)
        return ReviewVersion.model_validate(response)

    async def list(
        self,
        *,
        review_id: str,
        limit: int = 50,
        before: str | None = None,
        after: str | None = None,
    ) -> ReviewVersionListResponse:
        """List versions for one review."""
        request = self.prepare_list(
            review_id=review_id,
            limit=limit,
            before=before,
            after=after,
        )
        response = await self._client._prepared_request(request)
        return ReviewVersionListResponse.model_validate(response)


class WorkflowReviews(SyncAPIResource, WorkflowReviewsMixin):
    """Workflow reviews API wrapper for synchronous operations."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.versions = WorkflowReviewVersions(client=self._client)

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
    ) -> PaginatedList[Review]:
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
        return PaginatedList[Review].model_validate(response)

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


class AsyncWorkflowReviews(AsyncAPIResource, WorkflowReviewsMixin):
    """Workflow reviews API wrapper for asynchronous operations."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.versions = AsyncWorkflowReviewVersions(client=self._client)

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
    ) -> PaginatedList[Review]:
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
        return PaginatedList[Review].model_validate(response)

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
