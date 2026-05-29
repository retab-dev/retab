namespace Retab
{

    /// <summary>A single execution of a workflow.</summary>
    public class WorkflowRun
    {

        /// <summary>Unique ID for this run</summary>
        public string Id { get; set; } = default!;

        /// <summary>ID of the workflow that was run</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>Content-addressed workflow version used for this run.</summary>
        public string WorkflowVersionId { get; set; } = default!;

        /// <summary>What started this run</summary>
        public TriggerInfo Trigger { get; set; } = default!;

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
