namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>A partition produced by a workflow run, tagged with its artifact `operation` and creation time.</summary>
    public class PartitionWorkflowArtifact
    {

        /// <summary>Unique identifier of the partition</summary>
        public string Id { get; set; } = default!;

        /// <summary>Information about the partitioned file</summary>
        public FileRef File { get; set; } = default!;

        /// <summary>Model used for the partition operation</summary>
        public string Model { get; set; } = default!;

        /// <summary>Partition key used for the run</summary>
        public string Key { get; set; } = default!;

        /// <summary>Free-form instructions supplied with the partition request</summary>
        public string? Instructions { get; set; }

        /// <summary>Number of consensus votes used</summary>
        public long? NConsensus { get; set; }

        /// <summary>Whether pages were allowed to appear in more than one partition chunk</summary>
        public bool? AllowOverlap { get; set; } = true;

        /// <summary>The list of partition chunks with their assigned pages</summary>
        public List<PartitionChunk>? Output { get; set; }

        /// <summary>Consensus metadata for multi-vote partition runs</summary>
        public PartitionConsensus? Consensus { get; set; }

        /// <summary>Usage information for the partition operation</summary>
        public RetabUsage? Usage { get; set; }

        /// <summary>Timestamp when this artifact was created.</summary>
        public DateTimeOffset? CreatedAt { get; set; }

        /// <summary>The operation that produced this artifact</summary>
        public string? Operation { get; set; } = "partition";

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
