namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Public step status object.</summary>
    public class WorkflowRunStep
    {

        /// <summary>Logical ID of the block</summary>
        public string BlockId { get; set; } = default!;

        /// <summary>Full step ID with iteration context. Assigned ONCE at creation, never recomputed.</summary>
        public string StepId { get; set; } = default!;

        /// <summary>Type of the block</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public WorkflowBlockType BlockType { get; set; }

        /// <summary>Label of the block</summary>
        public string BlockLabel { get; set; } = default!;

        /// <summary>Current step lifecycle</summary>
        [Newtonsoft.Json.JsonConverter(typeof(PendingStepLifecycleDiscriminatorConverter))]
        public object Lifecycle { get; set; } = default!;

        /// <summary>When the step started executing</summary>
        public DateTimeOffset? StartedAt { get; set; }

        /// <summary>When the step finished executing</summary>
        public DateTimeOffset? CompletedAt { get; set; }

        /// <summary>LLM model used by this step, when applicable</summary>
        public string? Model { get; set; }

        /// <summary>Container hierarchy from outermost to innermost. Empty when not inside any container.</summary>
        public List<ContainerContextData>? LoopContainers { get; set; }

        /// <summary>Parent workflow run ID</summary>
        public string RunId { get; set; } = default!;

        /// <summary>When the step was created</summary>
        public DateTimeOffset? CreatedAt { get; set; }

        /// <summary>Handle input payloads consumed by this step</summary>
        public Dictionary<string, PublicHandlePayload>? HandleInputs { get; set; }

        /// <summary>Handle output payloads produced by this step</summary>
        public Dictionary<string, PublicHandlePayload>? HandleOutputs { get; set; }

        /// <summary>Reference to the result produced by this step, if any.</summary>
        public StepArtifactRef? Artifact { get; set; }

        /// <summary>Number of retry attempts</summary>
        public long? RetryCount { get; set; }

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
