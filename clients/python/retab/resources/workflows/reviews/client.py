"""Actor-neutral client for workflow reviews."""

from __future__ import annotations

import asyncio
import time
from typing import Any, Dict, Literal

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....exceptions import NotFoundError
from ....types.standards import PreparedRequest
from ....types.workflows.reviews import (
    ReviewOverlay,
    ReviewQueueResponse,
    SubmitDecisionResponse,
)


class WorkflowReviewsMixin:
    """Mixin providing shared prepare methods for workflow review operations."""

    def prepare_list(
        self,
        workflow_id: str | None = None,
        limit: int = 50,
        decision: Literal["none", "any"] = "none",
        before: str | None = None,
        after: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to list review-queue items."""
        params: Dict[str, Any] = {"limit": limit, "decision": decision}
        if workflow_id is not None:
            params["workflow_id"] = workflow_id
        if before is not None:
            params["before"] = before
        if after is not None:
            params["after"] = after
        return PreparedRequest(method="GET", url="/workflows/reviews", params=params)

    def prepare_get(self, run_id: str, block_id: str) -> PreparedRequest:
        """Prepare a request to fetch one review."""
        return PreparedRequest(method="GET", url=f"/workflows/reviews/{run_id}/{block_id}")

    def prepare_create_version(
        self,
        run_id: str,
        block_id: str,
        *,
        snapshot: dict,
        parent_id: str,
        note: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to append a new output version to the review."""
        data: Dict[str, Any] = {
            "snapshot": snapshot,
            "parent_id": parent_id,
            "note": note,
        }
        return PreparedRequest(
            method="POST",
            url=f"/workflows/reviews/{run_id}/{block_id}/versions",
            data=data,
        )

    def prepare_decision(
        self,
        run_id: str,
        block_id: str,
        *,
        verdict: str,
        version_id: str,
        reason: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to submit a verdict against the review."""
        data: Dict[str, Any] = {"verdict": verdict, "version_id": version_id}
        if reason is not None:
            data["reason"] = reason
        return PreparedRequest(
            method="POST",
            url=f"/workflows/reviews/{run_id}/{block_id}/decision",
            data=data,
        )


class WorkflowReviews(SyncAPIResource, WorkflowReviewsMixin):
    """Workflow reviews API wrapper for synchronous operations.

    Usage:
    - ``client.workflows.reviews.list()`` for the review queue.
    - ``client.workflows.reviews.get(run_id, block_id)`` for one review.
    - ``client.workflows.reviews.approve(...)`` / ``reject(...)`` to submit a verdict.
    - ``client.workflows.reviews.create_version(...)`` to append a new output version.
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def list(
        self,
        workflow_id: str | None = None,
        limit: int = 50,
        decision: Literal["none", "any"] = "none",
        before: str | None = None,
        after: str | None = None,
    ) -> ReviewQueueResponse:
        """List review-queue items.

        Args:
            workflow_id: Restrict the queue to a single workflow.
            limit: Page size (1-200).
            decision: Use ``decision='any'`` to include reviews that already have
                a terminal decision; the default ``'none'`` shows only those still
                awaiting a decision.
            before: Cursor — only return reviews that appear before this review
                id in the result order. Use ``list_metadata.before`` from the
                previous page. Mutually exclusive with ``after``.
            after: Cursor — only return reviews that appear after this review id
                in the result order. Use ``list_metadata.after`` from the
                previous page. Mutually exclusive with ``before``.

        Returns:
            ReviewQueueResponse: A page of lightweight :class:`ReviewQueueItem`
            summaries plus a ``list_metadata`` cursor (and a derived
            ``has_more`` property).
        """
        request = self.prepare_list(
            workflow_id=workflow_id,
            limit=limit,
            decision=decision,
            before=before,
            after=after,
        )
        response = self._client._prepared_request(request)
        return ReviewQueueResponse.model_validate(response)

    def get(self, run_id: str, block_id: str) -> ReviewOverlay:
        """Fetch the full review for a gated block run.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.

        Returns:
            ReviewOverlay: The full review, including version history,
            the terminal decision.

        Raises:
            NotFoundError: HTTP 404 — no review exists for this block yet.
        """
        request = self.prepare_get(run_id, block_id)
        response = self._client._prepared_request(request)
        return ReviewOverlay.model_validate(response)

    def approve(
        self,
        run_id: str,
        block_id: str,
        *,
        version_id: str,
    ) -> SubmitDecisionResponse:
        """Approve the gated block output.

        ``version_id`` must be a HEAD version — one with no child appended on
        top of it. Approving a superseded ancestor would silently discard the
        corrections layered above it, so the server rejects this with HTTP 422
        and names the current head in the error. To approve a correction,
        first call :meth:`create_version` to author it, then pass the returned
        version id here.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_id: Content-hash id of the exact version being approved.
                Must be a HEAD version (no descendant in ``versions_by_id``).

        Returns:
            SubmitDecisionResponse: Submission status and the updated review.

        Raises:
            ConflictError: HTTP 409 — the review already has a terminal decision.
            ValidationError: HTTP 422 — the version is superseded by a newer
                version. The error message names the current head.
        """
        request = self.prepare_decision(
            run_id,
            block_id,
            verdict="approved",
            version_id=version_id,
        )
        response = self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    def reject(
        self,
        run_id: str,
        block_id: str,
        *,
        version_id: str,
        reason: str,
    ) -> SubmitDecisionResponse:
        """Reject the gated block output.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_id: Content-hash id of the exact version being rejected.
            reason: Why the output was rejected — required so every rejection
                is auditable.

        Returns:
            SubmitDecisionResponse: Submission status and the updated review.

        Raises:
            ConflictError: HTTP 409 — the review already has a terminal decision.
        """
        request = self.prepare_decision(
            run_id,
            block_id,
            verdict="rejected",
            version_id=version_id,
            reason=reason,
        )
        response = self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    def create_version(
        self,
        run_id: str,
        block_id: str,
        *,
        snapshot: dict,
        parent_id: str,
        note: str | None = None,
    ) -> ReviewOverlay:
        """Append a new output version to the review's version history.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            snapshot: The new output payload to record as a version.
            parent_id: Content-hash id of the parent version.
            note: Optional free-text note attached to the version.

        Returns:
            ReviewOverlay: The review with the new version appended.

        Raises:
            ConflictError: HTTP 409 — the review already has a terminal decision.
        """
        request = self.prepare_create_version(
            run_id,
            block_id,
            snapshot=snapshot,
            parent_id=parent_id,
            note=note,
        )
        response = self._client._prepared_request(request)
        return ReviewOverlay.model_validate(response)

    def wait_for(
        self,
        run_id: str,
        block_id: str,
        *,
        timeout: float = 120.0,
        poll_interval: float = 2.0,
    ) -> ReviewOverlay:
        """Poll until the block is gated and awaiting review.

        Calls :meth:`get` on a loop until the review exists and has no
        terminal decision. A 404 means the workflow has not reached the
        gate yet — that is not an error, polling continues.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            timeout: Maximum seconds to wait before giving up.
            poll_interval: Seconds to sleep between polls.

        Returns:
            ReviewOverlay: The review once it is awaiting review.

        Raises:
            TimeoutError: The review did not reach ``awaiting_review`` in time.
        """
        deadline = time.monotonic() + timeout
        while True:
            try:
                overlay = self.get(run_id, block_id)
                if overlay.decision is None:
                    return overlay
            except NotFoundError:
                pass
            if time.monotonic() >= deadline:
                raise TimeoutError(f"Review for run {run_id!r} block {block_id!r} was not awaiting_review within {timeout}s")
            time.sleep(poll_interval)


class AsyncWorkflowReviews(AsyncAPIResource, WorkflowReviewsMixin):
    """Workflow reviews API wrapper for asynchronous operations.

    Usage:
    - ``await client.workflows.reviews.list()`` for the review queue.
    - ``await client.workflows.reviews.get(run_id, block_id)`` for one review.
    - ``await client.workflows.reviews.approve(...)`` / ``reject(...)`` to submit a verdict.
    - ``await client.workflows.reviews.create_version(...)`` to append an output version.
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def list(
        self,
        workflow_id: str | None = None,
        limit: int = 50,
        decision: Literal["none", "any"] = "none",
        before: str | None = None,
        after: str | None = None,
    ) -> ReviewQueueResponse:
        """List review-queue items.

        Args:
            workflow_id: Restrict the queue to a single workflow.
            limit: Page size (1-200).
            decision: Use ``decision='any'`` to include reviews that already have
                a terminal decision; the default ``'none'`` shows only those still
                awaiting a decision.
            before: Cursor — only return reviews that appear before this review
                id in the result order. Use ``list_metadata.before`` from the
                previous page. Mutually exclusive with ``after``.
            after: Cursor — only return reviews that appear after this review id
                in the result order. Use ``list_metadata.after`` from the
                previous page. Mutually exclusive with ``before``.

        Returns:
            ReviewQueueResponse: A page of lightweight :class:`ReviewQueueItem`
            summaries plus a ``list_metadata`` cursor (and a derived
            ``has_more`` property).
        """
        request = self.prepare_list(
            workflow_id=workflow_id,
            limit=limit,
            decision=decision,
            before=before,
            after=after,
        )
        response = await self._client._prepared_request(request)
        return ReviewQueueResponse.model_validate(response)

    async def get(self, run_id: str, block_id: str) -> ReviewOverlay:
        """Fetch the full review for a gated block run.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.

        Returns:
            ReviewOverlay: The full review, including version history,
            the terminal decision.

        Raises:
            NotFoundError: HTTP 404 — no review exists for this block yet.
        """
        request = self.prepare_get(run_id, block_id)
        response = await self._client._prepared_request(request)
        return ReviewOverlay.model_validate(response)

    async def approve(
        self,
        run_id: str,
        block_id: str,
        *,
        version_id: str,
    ) -> SubmitDecisionResponse:
        """Approve the gated block output.

        ``version_id`` must be a HEAD version — one with no child appended on
        top of it. Approving a superseded ancestor would silently discard the
        corrections layered above it, so the server rejects this with HTTP 422
        and names the current head in the error. To approve a correction,
        first call :meth:`create_version` to author it, then pass the returned
        version id here.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_id: Content-hash id of the exact version being approved.
                Must be a HEAD version (no descendant in ``versions_by_id``).

        Returns:
            SubmitDecisionResponse: Submission status and the updated review.

        Raises:
            ConflictError: HTTP 409 — the review already has a terminal decision.
            ValidationError: HTTP 422 — the version is superseded by a newer
                version. The error message names the current head.
        """
        request = self.prepare_decision(
            run_id,
            block_id,
            verdict="approved",
            version_id=version_id,
        )
        response = await self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    async def reject(
        self,
        run_id: str,
        block_id: str,
        *,
        version_id: str,
        reason: str,
    ) -> SubmitDecisionResponse:
        """Reject the gated block output.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_id: Content-hash id of the exact version being rejected.
            reason: Why the output was rejected — required so every rejection
                is auditable.

        Returns:
            SubmitDecisionResponse: Submission status and the updated review.

        Raises:
            ConflictError: HTTP 409 — the review already has a terminal decision.
        """
        request = self.prepare_decision(
            run_id,
            block_id,
            verdict="rejected",
            version_id=version_id,
            reason=reason,
        )
        response = await self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    async def create_version(
        self,
        run_id: str,
        block_id: str,
        *,
        snapshot: dict,
        parent_id: str,
        note: str | None = None,
    ) -> ReviewOverlay:
        """Append a new output version to the review's version history.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            snapshot: The new output payload to record as a version.
            parent_id: Content-hash id of the parent version.
            note: Optional free-text note attached to the version.

        Returns:
            ReviewOverlay: The review with the new version appended.

        Raises:
            ConflictError: HTTP 409 — the review already has a terminal decision.
        """
        request = self.prepare_create_version(
            run_id,
            block_id,
            snapshot=snapshot,
            parent_id=parent_id,
            note=note,
        )
        response = await self._client._prepared_request(request)
        return ReviewOverlay.model_validate(response)

    async def wait_for(
        self,
        run_id: str,
        block_id: str,
        *,
        timeout: float = 120.0,
        poll_interval: float = 2.0,
    ) -> ReviewOverlay:
        """Poll until the block is gated and awaiting review.

        Calls :meth:`get` on a loop until the review exists and has no
        terminal decision. A 404 means the workflow has not reached the
        gate yet — that is not an error, polling continues.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            timeout: Maximum seconds to wait before giving up.
            poll_interval: Seconds to sleep between polls.

        Returns:
            ReviewOverlay: The review once it is awaiting review.

        Raises:
            TimeoutError: The review did not reach ``awaiting_review`` in time.
        """
        deadline = time.monotonic() + timeout
        while True:
            try:
                overlay = await self.get(run_id, block_id)
                if overlay.decision is None:
                    return overlay
            except NotFoundError:
                pass
            if time.monotonic() >= deadline:
                raise TimeoutError(f"Review for run {run_id!r} block {block_id!r} was not awaiting_review within {timeout}s")
            await asyncio.sleep(poll_interval)
