namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a verdict summary.</summary>
    public class VerdictSummary
    {
        public bool Result { get; set; }
        public long? AssertionsPassed { get; set; }
        public long? AssertionsFailed { get; set; }
        public long? BlockedAssertions { get; set; }
        public List<string>? FailedAssertionIds { get; set; }

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
