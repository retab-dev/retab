namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="ExperimentRunMetricsService.GetAsync"/>: Get Experiment Metrics For Run</summary>
    public class ExperimentRunMetricsGetOptions : BaseOptions
    {
        public string RunId { get; set; } = default!;

        public ExperimentRunMetricsView? View { get; set; }

        public string? DocumentId { get; set; }

        public string? TargetPath { get; set; }

        public bool? IncludePrior { get; set; }

        public string? PriorRunId { get; set; }

    }
}
