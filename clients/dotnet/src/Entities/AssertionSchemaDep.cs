namespace Retab
{

    /// <summary>Single-rule schema dependency for Level 2 drift detection.</summary>
    public class AssertionSchemaDep
    {
        public string? OutputHandleId { get; set; }
        public string SchemaPath { get; set; } = default!;
        public string SubtreeHash { get; set; } = default!;
        public bool? DependsOnRoot { get; set; }

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
