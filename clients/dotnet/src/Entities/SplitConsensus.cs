namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a split consensus.</summary>
    public class SplitConsensus
    {

        /// <summary>Consensus likelihood tree mirroring the split output</summary>
        public List<SplitSubdocumentLikelihood>? Likelihoods { get; set; }

        /// <summary>Alternative split vote outputs used to build the consolidated result</summary>
        public List<List<SplitResult>>? Choices { get; set; }
    }
}
