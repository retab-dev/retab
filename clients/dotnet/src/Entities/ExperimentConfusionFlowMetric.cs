namespace Retab
{

    /// <summary>One non-diagonal confusion flow between two labels.</summary>
    public class ExperimentConfusionFlowMetric
    {
        public string Source { get; set; } = default!;
        public string Target { get; set; } = default!;
        public double Score { get; set; }
    }
}
