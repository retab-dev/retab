namespace Retab
{

    /// <summary>Represents a declarative plan summary.</summary>
    public class DeclarativePlanSummary
    {
        public long? Add { get; set; }
        public long? Change { get; set; }
        public long? Destroy { get; set; }
        public long? Replace { get; set; }
        public long? Noop { get; set; }
        public long? Total { get; set; }
        public bool? HasChanges { get; set; } = false;

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
