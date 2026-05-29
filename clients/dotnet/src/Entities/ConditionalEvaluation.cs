namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Record of how a conditional block routed during a workflow run.</summary>
    /// <remarks>
    /// Captures each condition that was evaluated (`evaluations`), which output
    /// branches were chosen (`selected_handles`), and the branch and condition
    /// IDs that matched (`matched_branch_id`, `matched_condition_ids`).
    /// </remarks>
    public class ConditionalEvaluation
    {

        /// <summary>The operation that produced this artifact</summary>
        public string? Operation { get; set; } = "conditional_evaluation";
        public string Id { get; set; } = default!;
        public string RunId { get; set; } = default!;
        public string StepId { get; set; } = default!;
        public List<ConditionEvaluationResult>? Evaluations { get; set; }
        public List<string>? SelectedHandles { get; set; }
        public string? MatchedBranchId { get; set; }
        public List<string>? MatchedConditionIds { get; set; }

        /// <summary>Timestamp when this artifact was created.</summary>
        public DateTimeOffset? CreatedAt { get; set; }

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
