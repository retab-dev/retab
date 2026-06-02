namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents an artifact freshness.</summary>
    public class ArtifactFreshness
    {
        public ArtifactFreshnessStatus? Status { get; set; }
        public List<ArtifactFreshnessReasons>? Reasons { get; set; }
        public string? ValidityFingerprint { get; set; }
        public string? InputFingerprint { get; set; }
        public string? BaselineRunId { get; set; }

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
