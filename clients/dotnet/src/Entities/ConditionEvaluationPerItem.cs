namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Per-item evaluation result for wildcard array conditions.</summary>
    /// <remarks>
    /// When a condition path contains .*, each array element is evaluated
    /// individually with implicit AND logic (all must match).
    /// </remarks>
    public class ConditionEvaluationPerItem
    {

        /// <summary>Index of this item in the array</summary>
        public long Index { get; set; }

        /// <summary>Hierarchical indices for nested arrays (e.g., [0, 2, 1] for items[0].subitems[2].field[1])</summary>
        public List<long>? Indices { get; set; }

        /// <summary>Actual value at this index</summary>
        public object? Actual { get; set; }

        /// <summary>Whether this item matched the condition</summary>
        public bool? Matched { get; set; } = false;

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
