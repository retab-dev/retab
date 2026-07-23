namespace Retab
{
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Run-level summary plus block-specific diagnostics.</summary>
    /// <remarks>
    /// `prior_run_id` + `prior_score` populate when the request opts into
    /// prior-comparison and a completed prior run exists.
    /// `score` and `documents` cover only the documents that produced a result.
    /// Compare `scored_document_count` against `total_document_count` to see
    /// whether any of the run's documents failed and were left out.
    /// </remarks>
    public class ExperimentSummaryMetricsResponse
    {
        public string ExperimentId { get; set; } = default!;
        public string RunId { get; set; } = default!;
        public string? Kind { get; set; } = "summary";
        public string? View { get; set; } = "summary";
        public string? BlockExecutionFingerprint { get; set; }
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public ExperimentBlockType BlockType { get; set; }
        public double? Score { get; set; }
        public double? PriorScore { get; set; }
        public List<ExperimentSummaryMetricDocument>? Documents { get; set; }
        public long? ScoredDocumentCount { get; set; }
        public long? TotalDocumentCount { get; set; }
        public OneOf.OneOf<ExperimentExtractSummaryAggregate, ExperimentConfusionSummaryAggregate>? Aggregate { get; set; }
        public string? PriorRunId { get; set; }

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
