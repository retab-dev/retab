namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Directional confusion slice for one split/classifier target.</summary>
    public class ExperimentTargetConfusionMetric
    {
        public double? Self { get; set; }
        public Dictionary<string, double>? FlowFrom { get; set; }
        public Dictionary<string, double>? FlowTo { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();
    }
}
