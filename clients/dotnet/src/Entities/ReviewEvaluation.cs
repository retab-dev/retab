namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Represents a review evaluation.</summary>
    public class ReviewEvaluation
    {

        /// <summary>Artifact operation that determines the backing record type</summary>
        public string? Operation { get; set; }
        public string Id { get; set; } = default!;
        public string RunId { get; set; } = default!;
        public string StepId { get; set; } = default!;
        public List<ConditionEvaluationResult>? Evaluations { get; set; }
        public List<string>? SelectedHandles { get; set; }
        public string? MatchedBranchId { get; set; }
        public List<string>? MatchedConditionIds { get; set; }
        public bool? RequiresHumanReview { get; set; }
        public string? ReviewerId { get; set; }
        public ReviewEvaluationReviewDecision? ReviewDecision { get; set; }
        public string? ReviewNotes { get; set; }
        public bool? RequestedRevision { get; set; }
        public DateTimeOffset? ReviewedAt { get; set; }

        /// <summary>When this artifact was written by the orchestrator.</summary>
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
