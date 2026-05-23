namespace Retab
{
    using System.Collections.Generic;

    /// <summary>One document's compact per-target breakdown.</summary>
    public class ExperimentByDocumentMetricsResponse
    {
        public string RunId { get; set; } = default!;
        public string? Kind { get; set; }
        public string? View { get; set; }
        public ExperimentMetricDocumentRef Document { get; set; } = default!;
        public double? Score { get; set; }
        public double? PriorScore { get; set; }
        public ExperimentConfusionSummaryAggregate? Confusion { get; set; }
        public List<ExperimentByDocumentTargetMetric>? Targets { get; set; }
    }
}
