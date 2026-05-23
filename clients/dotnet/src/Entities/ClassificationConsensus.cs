namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a classification consensus.</summary>
    public class ClassificationConsensus
    {

        /// <summary>Alternative classification vote outputs used to build the consolidated result.</summary>
        public List<ClassificationDecision>? Choices { get; set; }

        /// <summary>Consensus likelihood score (0.0-1.0) of the winning classification.</summary>
        public double? Likelihoods { get; set; }
    }
}
