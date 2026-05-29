namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a partition chunk.</summary>
    public class PartitionChunk
    {

        /// <summary>The partition key value for this chunk</summary>
        public string Key { get; set; } = default!;

        /// <summary>The pages assigned to this partition chunk (1-indexed)</summary>
        public List<long>? Pages { get; set; }

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
