namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Gate when a conditional-style result chose a branch in `branches`.</summary>
    /// <remarks>
    /// Conditional blocks are intentionally not reviewable. The predicate remains in
    /// the global union so old review/evaluator payloads can still be parsed.
    /// </remarks>
    public class ReviewBranchIn
    {
        public string? Kind { get; set; } = "branch_in";
        public List<string> Branches { get; set; } = default!;

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
