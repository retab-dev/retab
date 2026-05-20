"""Actor-neutral client for the workflow HIL review overlay.

Drives the review loop served under ``/workflows/reviews``: list the queue,
fetch an overlay, post new output versions, and submit verdicts.

Actor symmetry is a hard rule here. A proposal authored by a model, an agent,
or a human goes through the SAME :meth:`WorkflowReviews.edit` /
:meth:`WorkflowReviews.approve` pair. ``Actor.kind`` is data carried on the
overlay — no method, parameter, or branch in this module depends on it.

Every mutating call carries a ``version_stamp`` (the overlay's ``rev``). If the
server's ``rev`` has advanced since the caller last read it, the request fails
with HTTP 409 (:class:`retab.exceptions.ConflictError`) — re-read with
:meth:`WorkflowReviews.get` and retry.
"""

from __future__ import annotations

import asyncio
import time
from typing import Any, Dict

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....exceptions import NotFoundError
from ....types.standards import PreparedRequest
from ....types.workflows.reviews import (
    EditOrigin,
    ReviewOverlay,
    ReviewQueueResponse,
    ReviewStatus,
    SubmitDecisionResponse,
)


class WorkflowReviewsMixin:
    """Mixin providing shared prepare methods for workflow review operations."""

    def prepare_list(
        self,
        workflow_id: str | None = None,
        status: ReviewStatus = "awaiting_review",
        mine: bool = False,
        limit: int = 50,
    ) -> PreparedRequest:
        """Prepare a request to list review-queue items."""
        params: Dict[str, Any] = {"status": status, "mine": mine, "limit": limit}
        if workflow_id is not None:
            params["workflow_id"] = workflow_id
        return PreparedRequest(method="GET", url="/workflows/reviews", params=params)

    def prepare_get(self, run_id: str, block_id: str) -> PreparedRequest:
        """Prepare a request to fetch one review overlay."""
        return PreparedRequest(method="GET", url=f"/workflows/reviews/{run_id}/{block_id}")

    def prepare_edit(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
        snapshot: dict | None = None,
        reviewable_value: dict | None = None,
        origin: EditOrigin = "human_edit",
        note: str | None = None,
        command_id: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to append a new output version to the overlay."""
        data: Dict[str, Any] = {
            "snapshot": snapshot,
            "reviewable_value": reviewable_value,
            "version_stamp": version_stamp,
            "origin": origin,
            "note": note,
            "command_id": command_id,
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
        version_stamp: int,
        edited_output: dict | None = None,
        reviewable_value: dict | None = None,
        on_seq: int | None = None,
        effective_seq: int | None = None,
        reason: str | None = None,
        command_id: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to submit a verdict against the overlay."""
        data: Dict[str, Any] = {
            "verdict": verdict,
            "version_stamp": version_stamp,
            "edited_output": edited_output,
            "reviewable_value": reviewable_value,
            "on_seq": on_seq,
            "effective_seq": effective_seq,
            "reason": reason,
            "command_id": command_id,
        }
        return PreparedRequest(
            method="POST",
            url=f"/workflows/reviews/{run_id}/{block_id}/decision",
            data=data,
        )

    def prepare_claim(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
        ttl_seconds: int = 900,
    ) -> PreparedRequest:
        """Prepare a request to take the soft lock on the overlay."""
        data: Dict[str, Any] = {"version_stamp": version_stamp, "ttl_seconds": ttl_seconds}
        return PreparedRequest(
            method="POST",
            url=f"/workflows/reviews/{run_id}/{block_id}/claim",
            data=data,
        )

    def prepare_release(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
    ) -> PreparedRequest:
        """Prepare a request to release the soft lock on the overlay."""
        data: Dict[str, Any] = {"version_stamp": version_stamp}
        return PreparedRequest(
            method="POST",
            url=f"/workflows/reviews/{run_id}/{block_id}/release",
            data=data,
        )


class WorkflowReviews(SyncAPIResource, WorkflowReviewsMixin):
    """Workflow HIL review overlay API wrapper for synchronous operations.

    Usage:
    - ``client.workflows.reviews.list()`` for the review queue.
    - ``client.workflows.reviews.get(run_id, block_id)`` for one overlay.
    - ``client.workflows.reviews.approve(...)`` / ``reject(...)`` to submit a verdict.
    - ``client.workflows.reviews.edit(...)`` to append a new output version.
    - ``client.workflows.reviews.claim(...)`` / ``release(...)`` for the
      advisory soft lock.
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def list(
        self,
        workflow_id: str | None = None,
        status: ReviewStatus = "awaiting_review",
        mine: bool = False,
        limit: int = 50,
    ) -> ReviewQueueResponse:
        """List review-queue items.

        Args:
            workflow_id: Restrict the queue to a single workflow.
            status: Lifecycle filter — ``awaiting_review`` / ``approved`` / ``rejected``.
            mine: When True, only items claimed by the calling actor.
            limit: Page size (1-200).

        Returns:
            ReviewQueueResponse: A page of lightweight :class:`ReviewQueueItem`
            summaries plus a ``has_more`` flag.
        """
        request = self.prepare_list(workflow_id=workflow_id, status=status, mine=mine, limit=limit)
        response = self._client._prepared_request(request)
        return ReviewQueueResponse.model_validate(response)

    def get(self, run_id: str, block_id: str) -> ReviewOverlay:
        """Fetch the full review overlay for a gated block run.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.

        Returns:
            ReviewOverlay: The full overlay, including version history,
            decisions, and audit log.

        Raises:
            NotFoundError: HTTP 404 — no overlay exists for this block yet.
        """
        request = self.prepare_get(run_id, block_id)
        response = self._client._prepared_request(request)
        return ReviewOverlay.model_validate(response)

    def approve(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
        edited_output: dict | None = None,
        reviewable_value: dict | None = None,
        on_seq: int | None = None,
        effective_seq: int | None = None,
        command_id: str | None = None,
    ) -> SubmitDecisionResponse:
        """Approve the gated block output.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_stamp: The overlay ``rev`` last observed (CAS token).
            edited_output: Optional replacement output applied with the approval.
            reviewable_value: Optional primitive-specific value compiled by the server.
            on_seq: Version sequence the decision is made against.
            effective_seq: Version that becomes effective on apply.
            command_id: Optional idempotency key for deduplicating submissions.

        Returns:
            SubmitDecisionResponse: Submission status and the updated overlay.

        Raises:
            ConflictError: HTTP 409 — ``version_stamp`` is stale; re-read and retry.
        """
        request = self.prepare_decision(
            run_id,
            block_id,
            verdict="approved",
            version_stamp=version_stamp,
            edited_output=edited_output,
            reviewable_value=reviewable_value,
            on_seq=on_seq,
            effective_seq=effective_seq,
            command_id=command_id,
        )
        response = self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    def reject(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
        reason: str,
        command_id: str | None = None,
    ) -> SubmitDecisionResponse:
        """Reject the gated block output.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_stamp: The overlay ``rev`` last observed (CAS token).
            reason: Why the output was rejected — required so every rejection
                is auditable.
            command_id: Optional idempotency key for deduplicating submissions.

        Returns:
            SubmitDecisionResponse: Submission status and the updated overlay.

        Raises:
            ConflictError: HTTP 409 — ``version_stamp`` is stale; re-read and retry.
        """
        request = self.prepare_decision(
            run_id,
            block_id,
            verdict="rejected",
            version_stamp=version_stamp,
            reason=reason,
            command_id=command_id,
        )
        response = self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    def edit(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
        snapshot: dict | None = None,
        reviewable_value: dict | None = None,
        origin: EditOrigin = "human_edit",
        note: str | None = None,
        command_id: str | None = None,
    ) -> ReviewOverlay:
        """Append a new output version to the overlay's version history.

        A proposal authored by a human or an agent uses this same call — the
        only difference is ``origin``, which is descriptive provenance, not a
        behavioral switch.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            snapshot: The new output payload to record as a version.
            reviewable_value: Primitive-specific reviewed value compiled by the server.
            version_stamp: The overlay ``rev`` last observed (CAS token).
            origin: Provenance of the snapshot — ``human_edit`` or ``agent_edit``.
            note: Optional free-text note attached to the version.
            command_id: Optional idempotency key for deduplicating submissions.

        Returns:
            ReviewOverlay: The overlay with the new version appended.

        Raises:
            ConflictError: HTTP 409 — ``version_stamp`` is stale; re-read and retry.
        """
        request = self.prepare_edit(
            run_id,
            block_id,
            snapshot=snapshot,
            reviewable_value=reviewable_value,
            version_stamp=version_stamp,
            origin=origin,
            note=note,
            command_id=command_id,
        )
        response = self._client._prepared_request(request)
        return ReviewOverlay.model_validate(response)

    def claim(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
        ttl_seconds: int = 900,
    ) -> ReviewOverlay:
        """Take the advisory soft lock on the overlay.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_stamp: The overlay ``rev`` last observed (CAS token).
            ttl_seconds: How long the claim is held before it lapses.

        Returns:
            ReviewOverlay: The overlay with the claim recorded.

        Raises:
            ConflictError: HTTP 409 — ``version_stamp`` is stale; re-read and retry.
        """
        request = self.prepare_claim(run_id, block_id, version_stamp=version_stamp, ttl_seconds=ttl_seconds)
        response = self._client._prepared_request(request)
        return ReviewOverlay.model_validate(response)

    def release(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
    ) -> ReviewOverlay:
        """Release the advisory soft lock on the overlay.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_stamp: The overlay ``rev`` last observed (CAS token).

        Returns:
            ReviewOverlay: The overlay with the claim cleared.

        Raises:
            ConflictError: HTTP 409 — ``version_stamp`` is stale; re-read and retry.
        """
        request = self.prepare_release(run_id, block_id, version_stamp=version_stamp)
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

        Calls :meth:`get` on a loop until the overlay exists and its ``status``
        is ``awaiting_review``. A 404 means the workflow has not reached the
        gate yet — that is not an error, polling continues.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            timeout: Maximum seconds to wait before giving up.
            poll_interval: Seconds to sleep between polls.

        Returns:
            ReviewOverlay: The overlay once it is awaiting review.

        Raises:
            TimeoutError: The overlay did not reach ``awaiting_review`` in time.
        """
        deadline = time.monotonic() + timeout
        while True:
            try:
                overlay = self.get(run_id, block_id)
                if overlay.status == "awaiting_review":
                    return overlay
            except NotFoundError:
                pass
            if time.monotonic() >= deadline:
                raise TimeoutError(f"Review overlay for run {run_id!r} block {block_id!r} was not awaiting_review within {timeout}s")
            time.sleep(poll_interval)


class AsyncWorkflowReviews(AsyncAPIResource, WorkflowReviewsMixin):
    """Workflow HIL review overlay API wrapper for asynchronous operations.

    Usage:
    - ``await client.workflows.reviews.list()`` for the review queue.
    - ``await client.workflows.reviews.get(run_id, block_id)`` for one overlay.
    - ``await client.workflows.reviews.approve(...)`` / ``reject(...)`` to submit a verdict.
    - ``await client.workflows.reviews.edit(...)`` to append an output version.
    - ``await client.workflows.reviews.claim(...)`` / ``release(...)`` for the
      advisory soft lock.
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def list(
        self,
        workflow_id: str | None = None,
        status: ReviewStatus = "awaiting_review",
        mine: bool = False,
        limit: int = 50,
    ) -> ReviewQueueResponse:
        """List review-queue items.

        Args:
            workflow_id: Restrict the queue to a single workflow.
            status: Lifecycle filter — ``awaiting_review`` / ``approved`` / ``rejected``.
            mine: When True, only items claimed by the calling actor.
            limit: Page size (1-200).

        Returns:
            ReviewQueueResponse: A page of lightweight :class:`ReviewQueueItem`
            summaries plus a ``has_more`` flag.
        """
        request = self.prepare_list(workflow_id=workflow_id, status=status, mine=mine, limit=limit)
        response = await self._client._prepared_request(request)
        return ReviewQueueResponse.model_validate(response)

    async def get(self, run_id: str, block_id: str) -> ReviewOverlay:
        """Fetch the full review overlay for a gated block run.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.

        Returns:
            ReviewOverlay: The full overlay, including version history,
            decisions, and audit log.

        Raises:
            NotFoundError: HTTP 404 — no overlay exists for this block yet.
        """
        request = self.prepare_get(run_id, block_id)
        response = await self._client._prepared_request(request)
        return ReviewOverlay.model_validate(response)

    async def approve(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
        edited_output: dict | None = None,
        reviewable_value: dict | None = None,
        on_seq: int | None = None,
        effective_seq: int | None = None,
        command_id: str | None = None,
    ) -> SubmitDecisionResponse:
        """Approve the gated block output.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_stamp: The overlay ``rev`` last observed (CAS token).
            edited_output: Optional replacement output applied with the approval.
            reviewable_value: Optional primitive-specific value compiled by the server.
            on_seq: Version sequence the decision is made against.
            effective_seq: Version that becomes effective on apply.
            command_id: Optional idempotency key for deduplicating submissions.

        Returns:
            SubmitDecisionResponse: Submission status and the updated overlay.

        Raises:
            ConflictError: HTTP 409 — ``version_stamp`` is stale; re-read and retry.
        """
        request = self.prepare_decision(
            run_id,
            block_id,
            verdict="approved",
            version_stamp=version_stamp,
            edited_output=edited_output,
            reviewable_value=reviewable_value,
            on_seq=on_seq,
            effective_seq=effective_seq,
            command_id=command_id,
        )
        response = await self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    async def reject(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
        reason: str,
        command_id: str | None = None,
    ) -> SubmitDecisionResponse:
        """Reject the gated block output.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_stamp: The overlay ``rev`` last observed (CAS token).
            reason: Why the output was rejected — required so every rejection
                is auditable.
            command_id: Optional idempotency key for deduplicating submissions.

        Returns:
            SubmitDecisionResponse: Submission status and the updated overlay.

        Raises:
            ConflictError: HTTP 409 — ``version_stamp`` is stale; re-read and retry.
        """
        request = self.prepare_decision(
            run_id,
            block_id,
            verdict="rejected",
            version_stamp=version_stamp,
            reason=reason,
            command_id=command_id,
        )
        response = await self._client._prepared_request(request)
        return SubmitDecisionResponse.model_validate(response)

    async def edit(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
        snapshot: dict | None = None,
        reviewable_value: dict | None = None,
        origin: EditOrigin = "human_edit",
        note: str | None = None,
        command_id: str | None = None,
    ) -> ReviewOverlay:
        """Append a new output version to the overlay's version history.

        A proposal authored by a human or an agent uses this same call — the
        only difference is ``origin``, which is descriptive provenance, not a
        behavioral switch.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            snapshot: The new output payload to record as a version.
            reviewable_value: Primitive-specific reviewed value compiled by the server.
            version_stamp: The overlay ``rev`` last observed (CAS token).
            origin: Provenance of the snapshot — ``human_edit`` or ``agent_edit``.
            note: Optional free-text note attached to the version.
            command_id: Optional idempotency key for deduplicating submissions.

        Returns:
            ReviewOverlay: The overlay with the new version appended.

        Raises:
            ConflictError: HTTP 409 — ``version_stamp`` is stale; re-read and retry.
        """
        request = self.prepare_edit(
            run_id,
            block_id,
            snapshot=snapshot,
            reviewable_value=reviewable_value,
            version_stamp=version_stamp,
            origin=origin,
            note=note,
            command_id=command_id,
        )
        response = await self._client._prepared_request(request)
        return ReviewOverlay.model_validate(response)

    async def claim(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
        ttl_seconds: int = 900,
    ) -> ReviewOverlay:
        """Take the advisory soft lock on the overlay.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_stamp: The overlay ``rev`` last observed (CAS token).
            ttl_seconds: How long the claim is held before it lapses.

        Returns:
            ReviewOverlay: The overlay with the claim recorded.

        Raises:
            ConflictError: HTTP 409 — ``version_stamp`` is stale; re-read and retry.
        """
        request = self.prepare_claim(run_id, block_id, version_stamp=version_stamp, ttl_seconds=ttl_seconds)
        response = await self._client._prepared_request(request)
        return ReviewOverlay.model_validate(response)

    async def release(
        self,
        run_id: str,
        block_id: str,
        *,
        version_stamp: int,
    ) -> ReviewOverlay:
        """Release the advisory soft lock on the overlay.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            version_stamp: The overlay ``rev`` last observed (CAS token).

        Returns:
            ReviewOverlay: The overlay with the claim cleared.

        Raises:
            ConflictError: HTTP 409 — ``version_stamp`` is stale; re-read and retry.
        """
        request = self.prepare_release(run_id, block_id, version_stamp=version_stamp)
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

        Calls :meth:`get` on a loop until the overlay exists and its ``status``
        is ``awaiting_review``. A 404 means the workflow has not reached the
        gate yet — that is not an error, polling continues.

        Args:
            run_id: The workflow run id.
            block_id: The gated block id.
            timeout: Maximum seconds to wait before giving up.
            poll_interval: Seconds to sleep between polls.

        Returns:
            ReviewOverlay: The overlay once it is awaiting review.

        Raises:
            TimeoutError: The overlay did not reach ``awaiting_review`` in time.
        """
        deadline = time.monotonic() + timeout
        while True:
            try:
                overlay = await self.get(run_id, block_id)
                if overlay.status == "awaiting_review":
                    return overlay
            except NotFoundError:
                pass
            if time.monotonic() >= deadline:
                raise TimeoutError(f"Review overlay for run {run_id!r} block {block_id!r} was not awaiting_review within {timeout}s")
            await asyncio.sleep(poll_interval)
