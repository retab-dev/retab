namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a partition chunk likelihood.</summary>
    public class PartitionChunkLikelihood
    {

        /// <summary>Confidence that this partition key value is correct</summary>
        public double? Key { get; set; }

        /// <summary>Confidence for each page in the corresponding partition chunk.pages array</summary>
        public List<double>? Pages { get; set; }
    }
}
