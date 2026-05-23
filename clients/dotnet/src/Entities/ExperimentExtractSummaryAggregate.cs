namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Extract-only diagnostics attached to the summary response.</summary>
    public class ExperimentExtractSummaryAggregate
    {
        public Dictionary<string, double>? Likelihoods { get; set; }
    }
}
