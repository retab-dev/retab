namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Evaluation result for a sub-condition in a compound condition.</summary>
    /// <remarks>
    /// Used when multiple conditions are combined with AND/OR operators.
    /// </remarks>
    public class ConditionEvaluationSubCondition
    {

        /// <summary>Identifier for this sub-condition</summary>
        public string? SubConditionId { get; set; }

        /// <summary>JSON path that was evaluated</summary>
        public string? Path { get; set; }

        /// <summary>Comparison operator used</summary>
        public string? Operator { get; set; }

        /// <summary>Expected value</summary>
        public object? Expected { get; set; }

        /// <summary>Actual value found</summary>
        public object? Actual { get; set; }

        /// <summary>Whether this sub-condition matched</summary>
        public bool? Matched { get; set; }

        /// <summary>Per-item breakdown if this sub-condition used a wildcard path</summary>
        public List<ConditionEvaluationPerItem>? PerItem { get; set; }
    }
}
