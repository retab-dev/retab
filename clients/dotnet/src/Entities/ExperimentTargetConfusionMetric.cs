namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Directional confusion slice for one split/classifier target.</summary>
    public class ExperimentTargetConfusionMetric
    {
        public double? Self { get; set; }
        public Dictionary<string, double>? FlowFrom { get; set; }
        public Dictionary<string, double>? FlowTo { get; set; }
    }
}
