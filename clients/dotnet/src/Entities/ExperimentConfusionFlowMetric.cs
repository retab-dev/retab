namespace Retab
{

    /// <summary>One non-diagonal confusion flow between two labels.</summary>
    public class ExperimentConfusionFlowMetric
    {
        public string Source { get; set; } = default!;
        public string Target { get; set; } = default!;
        public double Score { get; set; }

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
