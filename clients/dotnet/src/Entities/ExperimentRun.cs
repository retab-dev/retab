namespace Retab
{
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>A single execution of an experiment, identified by `id`.</summary>
    public class ExperimentRun
    {
        public string Id { get; set; } = default!;
        public string WorkflowId { get; set; } = default!;
        public string WorkflowVersionId { get; set; } = default!;
        public ExperimentRunTrigger Trigger { get; set; } = default!;
        public string ExperimentId { get; set; } = default!;
        public string BlockId { get; set; } = default!;
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public ExperimentBlockType BlockType { get; set; }
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public CreateExperimentRequestNConsensus NConsensus { get; set; }
        [Newtonsoft.Json.JsonConverter(typeof(PendingWorkflowExperimentRunDiscriminatorConverter))]
        public object Lifecycle { get; set; } = default!;
        public ExperimentRunTiming Timing { get; set; } = default!;
        public string? ParentRunId { get; set; }
        public string? BlockVersionId { get; set; }
        public string? MetricsValidityFingerprint { get; set; }
        public long? MetricsValidityFingerprintVersion { get; set; }
        public string DefinitionFingerprint { get; set; } = default!;
        public string DocumentsFingerprint { get; set; } = default!;
        public double? Score { get; set; }
        public long? TotalDocumentCount { get; set; }
        public long? CompletedDocumentCount { get; set; }
        public long? DocumentCount { get; set; }
        public long? ErrorCount { get; set; }

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
