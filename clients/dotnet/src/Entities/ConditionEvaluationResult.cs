namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Complete evaluation result for a termination condition.</summary>
    /// <remarks>
    /// This model represents the full evaluation data sent to the frontend
    /// for displaying in the Exit Trigger Evaluation dialog.
    /// The frontend expects data at both top-level and nested in 'details'
    /// for compatibility with the ConditionalEvaluationsTable component.
    /// </remarks>
    public class ConditionEvaluationResult
    {

        /// <summary>Unique identifier for this condition</summary>
        public string ConditionId { get; set; } = default!;

        /// <summary>JSON path that was evaluated</summary>
        public string? Path { get; set; } = "";

        /// <summary>Comparison operator used</summary>
        public string? Operator { get; set; } = "";

        /// <summary>Expected value</summary>
        public object? Expected { get; set; }

        /// <summary>Actual value found</summary>
        public object? Actual { get; set; }

        /// <summary>Whether the condition matched</summary>
        public bool? Matched { get; set; } = false;

        /// <summary>Branch name (always 'exit' for while-loop termination)</summary>
        public string? BranchName { get; set; } = "exit";

        /// <summary>Logical operator for compound conditions</summary>
        public ConditionEvaluationDetailsLogicalOperator? LogicalOperator { get; set; }

        /// <summary>Per-item breakdown for wildcard array conditions</summary>
        public List<ConditionEvaluationPerItem>? Items { get; set; }

        /// <summary>Sub-condition evaluations for compound conditions</summary>
        public List<ConditionEvaluationSubCondition>? SubEvaluations { get; set; }

        /// <summary>Nested details object for frontend compatibility</summary>
        public ConditionEvaluationDetails Details { get; set; } = default!;

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
