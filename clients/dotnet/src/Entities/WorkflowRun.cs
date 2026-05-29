namespace Retab
{

    /// <summary>A single execution of a workflow.</summary>
    public class WorkflowRun
    {

        /// <summary>Unique ID for this run</summary>
        public string Id { get; set; } = default!;

        /// <summary>Workflow + version reference</summary>
        public WorkflowSnapshotRef Workflow { get; set; } = default!;

        /// <summary>What started this run</summary>
        [Newtonsoft.Json.JsonConverter(typeof(ManualTriggerDiscriminatorConverter))]
        public object Trigger { get; set; } = default!;

        /// <summary>Lifecycle state of the run.</summary>
        [Newtonsoft.Json.JsonConverter(typeof(PendingRunDiscriminatorConverter))]
        public object Lifecycle { get; set; } = default!;

        /// <summary>All timing information</summary>
        public RunTiming Timing { get; set; } = default!;

        /// <summary>Input payloads supplied at run creation time</summary>
        public RunInputs? Inputs { get; set; }

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
