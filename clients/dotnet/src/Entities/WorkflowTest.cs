namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>A saved workflow test: a target block, an input `source`, and the `assertion` evaluated against its output.</summary>
    public class WorkflowTest
    {
        public string Id { get; set; } = default!;
        public string WorkflowId { get; set; } = default!;
        public WorkflowTestBlockTarget Target { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(ManualWorkflowTestSourceDiscriminatorConverter))]
        public object Source { get; set; } = default!;
        public string? Name { get; set; }
        public AssertionSpec? Assertion { get; set; }
        public AssertionSchemaDep? AssertionSchemaDep { get; set; }
        public AssertionDriftStatus? AssertionDriftStatus { get; set; }
        public ExperimentSchemaDriftStatus? SchemaDrift { get; set; }
        public string? SchemaDriftDetail { get; set; }
        public ArtifactFreshness? Freshness { get; set; }
        public ArtifactDrift? Drift { get; set; }
        public string? ValidationStatus { get; set; } = "valid";
        public List<string> ValidationIssues { get; set; } = default!;
        public LatestBlockTestRunSummary? LatestRunSummary { get; set; }
        public LatestBlockTestRunSummary? LatestPassingRunSummary { get; set; }
        public LatestBlockTestRunSummary? LatestFailingRunSummary { get; set; }

        /// <summary>When the workflow test was created</summary>
        public DateTimeOffset? CreatedAt { get; set; }

        /// <summary>When the workflow test was last updated</summary>
        public DateTimeOffset? UpdatedAt { get; set; }

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
