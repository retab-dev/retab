namespace Retab
{

    /// <summary>Represents a reconciliation field.</summary>
    public class ReconciliationField
    {
        public double Likelihood { get; set; }
        public string Path { get; set; } = default!;
        public long SupportingInputCount { get; set; }
        public long TotalInputCount { get; set; }
        public object Value { get; set; } = default!;

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
