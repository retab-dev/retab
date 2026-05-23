namespace Retab
{

    /// <summary>Returned when the experiment has no runs at all.</summary>
    public class ExperimentMetricsMissingError
    {
        public string? Kind { get; set; }
        public string? Error { get; set; }
        public string ExperimentId { get; set; } = default!;
        public string Message { get; set; } = default!;
    }
}
