namespace Retab
{

    /// <summary>Represents a run step workflow eval source.</summary>
    public class RunStepWorkflowEvalSource
    {
        public string? Type { get; set; } = "run_step";
        public string RunId { get; set; } = default!;
        public string? StepId { get; set; }

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
