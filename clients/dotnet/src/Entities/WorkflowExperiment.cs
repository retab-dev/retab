namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents a workflow experiment.</summary>
    public class WorkflowExperiment
    {
        public string Id { get; set; } = default!;
        public string WorkflowId { get; set; } = default!;
        public string BlockId { get; set; } = default!;
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public CreateExperimentRequestNConsensus NConsensus { get; set; }
        public long? DocumentCount { get; set; }
        public string Name { get; set; } = default!;
        public string? LastRunId { get; set; }

        /// <summary>When the experiment was created</summary>
        public DateTimeOffset? CreatedAt { get; set; }

        /// <summary>When the experiment was last updated</summary>
        public DateTimeOffset? UpdatedAt { get; set; }
        public ExperimentPublicStatus? Status { get; set; }
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public ExperimentBlockType BlockType { get; set; }
        public double? Score { get; set; }
        public bool? IsStale { get; set; } = false;
        public ExperimentSchemaDriftStatus? SchemaDrift { get; set; }
        public string? SchemaDriftDetail { get; set; }

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
