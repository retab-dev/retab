namespace Retab
{

    /// <summary>Reference to the workflow + immutable version that drove the run.</summary>
    /// <remarks>
    /// The class name is retained temporarily for compatibility with surrounding
    /// run-model code, but public API output uses ``version_id`` rather than
    /// snapshot identity.
    /// </remarks>
    public class WorkflowSnapshotRef
    {

        /// <summary>ID of the workflow that was run</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>Content-addressed workflow version used for this run.</summary>
        public string VersionId { get; set; } = default!;

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
