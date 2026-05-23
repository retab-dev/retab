namespace Retab
{
    using System.Collections.Generic;

    /// <summary>One target's compact per-document distribution.</summary>
    public class ExperimentByTargetMetricsResponse
    {
        public string RunId { get; set; } = default!;
        public string? Kind { get; set; }
        public string? View { get; set; }
        public string Target { get; set; } = default!;
        public double? Score { get; set; }
        public double? PriorScore { get; set; }
        public ExperimentTargetConfusionMetric? Confusion { get; set; }
        public List<ExperimentByTargetDocumentMetric>? Documents { get; set; }
    }
}
