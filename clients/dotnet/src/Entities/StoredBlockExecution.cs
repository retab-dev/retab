namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>The result of executing a single workflow block.</summary>
    /// <remarks>
    /// The execution state is carried by the `lifecycle` field.
    /// </remarks>
    public class StoredBlockExecution
    {

        /// <summary>Unique block execution ID</summary>
        public string Id { get; set; } = default!;

        /// <summary>Workflow the block belongs to</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>Workflow version whose source run supplied inputs</summary>
        public string? WorkflowVersionId { get; set; }

        /// <summary>Workflow run whose inputs were used</summary>
        public string SourceRunId { get; set; } = default!;

        /// <summary>ID of the block that was executed</summary>
        public string BlockId { get; set; } = default!;

        /// <summary>Type of the block</summary>
        public string BlockType { get; set; } = default!;

        /// <summary>Lifecycle state for this block execution.</summary>
        [Newtonsoft.Json.JsonConverter(typeof(PendingBlockExecutionLifecycleDiscriminatorConverter))]
        public object Lifecycle { get; set; } = default!;

        /// <summary>Input payloads keyed by handle ID (file metadata for files, data for json)</summary>
        public Dictionary<string, OneOf.OneOf<BlockExecJsonHandleInput, BlockExecFileHandleInput>>? HandleInputs { get; set; }

        /// <summary>Reference to the artifact produced by this block execution, if any.</summary>
        public StepArtifactRef? Artifact { get; set; }

        /// <summary>Output payloads keyed by handle ID</summary>
        public Dictionary<string, OneOf.OneOf<BlockExecJsonHandleInput, BlockExecFileHandleInput>>? HandleOutputs { get; set; }

        /// <summary>Active output handles for routing decisions</summary>
        public List<string>? RoutingDecisions { get; set; }

        /// <summary>Duration of the block execution in milliseconds</summary>
        public double? DurationMs { get; set; }

        /// <summary>When the block execution record was created</summary>
        public DateTimeOffset? CreatedAt { get; set; }

        /// <summary>When the block execution started</summary>
        public DateTimeOffset? StartedAt { get; set; }

        /// <summary>When the block execution completed</summary>
        public DateTimeOffset? CompletedAt { get; set; }
        public string? HandleInputsFingerprint { get; set; }
        public string? WorkflowDraftFingerprint { get; set; }
        public string? BlockExecutionFingerprint { get; set; }
        public string? ExecutionFingerprint { get; set; }

        /// <summary>The draft block config used for this block execution</summary>
        public Dictionary<string, object>? BlockConfig { get; set; }

        /// <summary>The step ID that was used for inputs (includes iteration prefix if applicable)</summary>
        public string? SourceStepId { get; set; }

        /// <summary>When the block has multiple iterations, lists all available ones</summary>
        public List<Dictionary<string, object>>? AvailableIterations { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

        /// <summary>
        /// Typed accessor for <see cref="BlockConfig"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetBlockConfigAttribute<T>(string key)
        {
            if (this.BlockConfig == null)
            {
                return default;
            }

            if (!this.BlockConfig.TryGetValue(key, out var value))
            {
                return default;
            }

            if (value is T typed)
            {
                return typed;
            }

            if (value is Newtonsoft.Json.Linq.JToken token)
            {
                return token.ToObject<T>();
            }

            if (value is System.Text.Json.JsonElement element)
            {
                return System.Text.Json.JsonSerializer.Deserialize<T>(element.GetRawText());
            }

            return default;
        }
    }
}
