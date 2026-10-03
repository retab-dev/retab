namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents a source job.</summary>
    public class SourceJob
    {
        public SourceJobError? Error { get; set; }
        public string Id { get; set; } = default!;

        /// <summary>Processing stopped before all requested work finished. A missing match alone is not a processing failure.</summary>
        public bool IsPartial { get; set; }
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public CreateSourcesRequestMode Mode { get; set; }
        public string Object { get; internal set; } = "extraction.sources.job";
        public long Revision { get; set; }
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public SourceJobStatus Status { get; set; }
        public DateTimeOffset UpdatedAt { get; set; }

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
