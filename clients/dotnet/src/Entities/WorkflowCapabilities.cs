namespace Retab
{

    /// <summary>Server-derived permissions for the current actor.</summary>
    /// <remarks>
    /// These fields are response-only. They should not be persisted on
    /// ``StoredWorkflow`` documents.
    /// </remarks>
    public class WorkflowCapabilities
    {
        public bool? CanView { get; set; } = false;
        public bool? CanEdit { get; set; } = false;
        public bool? CanRun { get; set; } = false;
        public bool? CanReview { get; set; } = false;
        public bool? CanPublish { get; set; } = false;
        public bool? CanManageMembers { get; set; } = false;
        public bool? CanManageSettings { get; set; } = false;
        public bool? CanDelete { get; set; } = false;

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
