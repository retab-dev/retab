namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Record of a review-gate evaluation during a workflow run.</summary>
    /// <remarks>
    /// Captures the conditions evaluated against the block's output
    /// (`evaluations`), whether the gate required human review
    /// (`requires_human_review`), and, once a reviewer acts, the verdict
    /// (`review_decision`), any notes, whether a revision was requested, and the
    /// reviewer and timestamp.
    /// </remarks>
    public class ReviewEvaluation
    {

        /// <summary>The operation that produced this artifact</summary>
        public string? Operation { get; set; } = "review_trigger_evaluation";
        public string Id { get; set; } = default!;
        public string RunId { get; set; } = default!;
        public string StepId { get; set; } = default!;
        public List<ConditionEvaluationResult>? Evaluations { get; set; }
        public List<string>? SelectedHandles { get; set; }
        public string? MatchedBranchId { get; set; }
        public List<string>? MatchedConditionIds { get; set; }
        public bool? RequiresHumanReview { get; set; } = false;
        public string? ReviewerId { get; set; }
        public ReviewEvaluationReviewDecision? ReviewDecision { get; set; }
        public string? ReviewNotes { get; set; }
        public bool? RequestedRevision { get; set; } = false;
        public DateTimeOffset? ReviewedAt { get; set; }

        /// <summary>Timestamp when this artifact was created.</summary>
        public DateTimeOffset CreatedAt { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();
    }
}
