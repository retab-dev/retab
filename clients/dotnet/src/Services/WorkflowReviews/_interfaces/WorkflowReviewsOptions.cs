namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowReviewsService.ListAsync"/>: List Reviews</summary>
    public class WorkflowReviewsListOptions : ListOptions
    {
        public string? WorkflowId { get; set; }

        public string? RunId { get; set; }

        public string? BlockId { get; set; }

        public string? StepId { get; set; }

        public string? IterationKey { get; set; }

        /// <summary>Filter by decision state: pending, approved, rejected, decided, cancelled, or all.</summary>
        public ReviewDecisionStatus? DecisionStatus { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowReviewsService.ApproveAsync"/>: Approve Review</summary>
    public class WorkflowReviewsApproveOptions : BaseOptions
    {
        /// <summary>Exact content-addressed key of the version to approve.</summary>
        public string VersionId { get; set; } = default!;

        /// <summary>Version ids the caller has seen and is deliberately superseding. A review's versions form a lineage; deciding (or parenting on) a version that is not its only latest version discards every other latest version, so the server refuses that write with a 409 unless every discarded id is listed here. Leave empty unless you are intentionally rolling back to an earlier version or choosing one arm of a forked lineage. The 409 detail names the exact ids to pass.</summary>
        public List<string>? AcknowledgedSupersededVersionIds { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowReviewsService.RejectAsync"/>: Reject Review</summary>
    public class WorkflowReviewsRejectOptions : BaseOptions
    {
        /// <summary>Exact content-addressed key of the version to reject.</summary>
        public string VersionId { get; set; } = default!;

        /// <summary>Required, non-empty rejection reason.</summary>
        public string Reason { get; set; } = default!;

    }
}
