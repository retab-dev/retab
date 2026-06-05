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

        /// <summary>Filter by decision state: pending, approved, rejected, decided, or all.</summary>
        public ReviewDecisionStatus? DecisionStatus { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowReviewsService.ApproveAsync"/>: Approve Review</summary>
    public class WorkflowReviewsApproveOptions : BaseOptions
    {
        /// <summary>Exact content-addressed key of the version to approve.</summary>
        public string VersionId { get; set; } = default!;

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
