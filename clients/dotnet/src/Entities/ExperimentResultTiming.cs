namespace Retab
{
    using System;

    /// <summary>Represents an experiment result timing.</summary>
    public class ExperimentResultTiming
    {
        public DateTimeOffset? CreatedAt { get; set; }
        public DateTimeOffset? StartedAt { get; set; }
        public DateTimeOffset? CompletedAt { get; set; }
        public long? DurationMs { get; set; }
    }
}
