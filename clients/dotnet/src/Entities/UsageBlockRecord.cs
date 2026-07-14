namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Represents an usage block record.</summary>
    public class UsageBlockRecord
    {
        public string BlockId { get; set; } = default!;
        public string BlockType { get; set; } = default!;
        public double Credits { get; set; }
        public long ExecutionCount { get; set; }
        public DateTimeOffset? FirstActivityAt { get; set; }
        public DateTimeOffset? LastActivityAt { get; set; }
        public long PageCount { get; set; }
        public long RunCount { get; set; }
        public Dictionary<string, long>? StatusCounts { get; set; }
        public string WorkflowId { get; set; } = default!;

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
