namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Returned when the latest run's config or document set has drifted</summary>
    /// <remarks>
    /// from the current draft, so its metrics no longer reflect the
    /// experiment's definition.
    /// </remarks>
    public class ExperimentMetricsStaleError
    {
        public string? Kind { get; set; } = "stale_metrics";
        public string? Error { get; set; } = "stale_metrics";
        public string ExperimentId { get; set; } = default!;
        public List<string>? StaleReasons { get; set; }
        public MetricsStaleErrorLastRun LastRun { get; set; } = default!;
        public string? CurrentConfigFingerprint { get; set; }
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
