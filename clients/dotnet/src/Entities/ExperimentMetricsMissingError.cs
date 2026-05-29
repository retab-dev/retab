namespace Retab
{

    /// <summary>Returned when the experiment has no runs at all.</summary>
    public class ExperimentMetricsMissingError
    {
        public string? Kind { get; set; } = "no_metrics";
        public string? Error { get; set; } = "no_metrics";
        public string ExperimentId { get; set; } = default!;
        public string Message { get; set; } = default!;

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
