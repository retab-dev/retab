namespace Retab
{
    using System.Collections.Generic;

    /// <summary>One document's compact per-target breakdown.</summary>
    public class ExperimentByDocumentMetricsResponse
    {
        public string RunId { get; set; } = default!;
        public string? Kind { get; set; }
        public string? View { get; set; }
        public ExperimentMetricDocumentRef Document { get; set; } = default!;
        public double? Score { get; set; }
        public double? PriorScore { get; set; }
        public ExperimentConfusionSummaryAggregate? Confusion { get; set; }
        public List<ExperimentByDocumentTargetMetric>? Targets { get; set; }

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
