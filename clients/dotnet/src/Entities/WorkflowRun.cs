namespace Retab
{

    /// <summary>Public workflow run response without tenant isolation fields.</summary>
    /// <remarks>
    /// This is the API response shape — distinct from the internal storage
    /// model (:class:`StoredWorkflowRun`), which carries persistence-only
    /// fields that never appear in API responses. Routes call
    /// :func:`serialize_workflow_run_response` to convert the storage shape
    /// into this response shape; constructing this class directly for
    /// persistence would drop those storage-only fields silently. The two
    /// classes were deliberately renamed to avoid the prior name collision.
    /// </remarks>
    public class WorkflowRun
    {

        /// <summary>Unique ID for this run</summary>
        public string Id { get; set; } = default!;

        /// <summary>Workflow + version reference</summary>
        public WorkflowSnapshotRef Workflow { get; set; } = default!;

        /// <summary>What started this run</summary>
        [Newtonsoft.Json.JsonConverter(typeof(ManualTriggerDiscriminatorConverter))]
        public object Trigger { get; set; } = default!;

        /// <summary>Discriminated lifecycle state.</summary>
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
