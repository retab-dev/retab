namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Represents a split.</summary>
    public class Split
    {

        /// <summary>Unique identifier of the split result</summary>
        public string Id { get; set; } = default!;

        /// <summary>Information about the split file</summary>
        public FileRef File { get; set; } = default!;

        /// <summary>Model used for the split operation</summary>
        public string Model { get; set; } = default!;

        /// <summary>Subdocuments used for the split operation</summary>
        public List<Subdocument> Subdocuments { get; set; } = default!;

        /// <summary>Number of consensus votes used</summary>
        public long? NConsensus { get; set; }

        /// <summary>Free-form instructions supplied with the split request.</summary>
        public string? Instructions { get; set; }

        /// <summary>The list of document splits with their assigned pages</summary>
        public List<SplitResult> Output { get; set; } = default!;

        /// <summary>Consensus metadata for multi-vote split runs</summary>
        public SplitConsensus? Consensus { get; set; }

        /// <summary>Usage information for the split operation</summary>
        public RetabUsage? Usage { get; set; }
        public DateTimeOffset? CreatedAt { get; set; }
    }
}
