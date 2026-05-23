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
        public object? Lifecycle { get; set; }

        /// <summary>All timing information</summary>
        public RunTiming? Timing { get; set; }

        /// <summary>Input payloads supplied at run creation time</summary>
        public RunInputs? Inputs { get; set; }
    }
}
