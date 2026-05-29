namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Gate fires if ANY child predicate fires. Evaluated in list order;</summary>
    /// <remarks>
    /// `triggered_by` reports the first match (decision: first-match wins).
    /// </remarks>
    public class ReviewAnyOf
    {
        public string? Kind { get; set; } = "any_of";
        public List<object> Predicates { get; set; } = default!;

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
