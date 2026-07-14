namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Represents an usage primitive record.</summary>
    public class UsagePrimitiveRecord
    {
        public string? BlockId { get; set; }
        public DateTimeOffset? CompletedAt { get; set; }
        public DateTimeOffset? CreatedAt { get; set; }
        public double Credits { get; set; }
        public List<UsagePrimitiveDocument>? Documents { get; set; }
        public long? DurationMs { get; set; }
        public Dictionary<string, string>? Metadata { get; set; }
        public string? Model { get; set; }
        public string Operation { get; set; } = default!;
        public long PageCount { get; set; }
        public string PrimitiveExecutionId { get; set; } = default!;
        public string? ProjectId { get; set; }
        public string? ResourceKind { get; set; }
        public string? RunId { get; set; }
        public string Status { get; set; } = default!;
        public string? WorkflowId { get; set; }

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
