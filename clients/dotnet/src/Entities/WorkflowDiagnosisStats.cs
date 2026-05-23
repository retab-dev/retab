namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a workflow diagnosis stats.</summary>
    public class WorkflowDiagnosisStats
    {

        /// <summary>Total number of blocks diagnosed</summary>
        public long? TotalBlocks { get; set; }

        /// <summary>Total number of edges diagnosed</summary>
        public long? TotalEdges { get; set; }

        /// <summary>Counts by block type</summary>
        public Dictionary<string, long>? BlockTypes { get; set; }

        /// <summary>Number of start_document blocks</summary>
        public long? StartDocumentBlocks { get; set; }
    }
}
