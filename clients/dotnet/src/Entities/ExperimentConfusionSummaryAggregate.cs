namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Split/classifier diagnostics attached to the summary response.</summary>
    public class ExperimentConfusionSummaryAggregate
    {
        public Dictionary<string, double>? Diag { get; set; }
        public List<ExperimentConfusionFlowMetric>? Flows { get; set; }
    }
}
