namespace Retab
{
    using System.Collections.Generic;

    /// <summary>The outcome of one eval within an eval run: its `lifecycle`, `timing`, and `verdict`.</summary>
    public class WorkflowEvalResult
    {
        public string Id { get; set; } = default!;
        public string? WorkflowEvalRunId { get; set; }
        public string EvalId { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(PendingWorkflowEvalRunDiscriminatorConverter))]
        public object? Lifecycle { get; set; }
        public ExperimentRunTiming? Timing { get; set; }

        /// <summary>Verdict label populated only when the underlying eval reaches a terminal lifecycle state and the verdict could be determined. Execution-error details flow through `lifecycle`, not through this enum.</summary>
        public AssertionOutcome? Verdict { get; set; }
        public string WorkflowId { get; set; } = default!;
        public string BlockId { get; set; } = default!;
        public string BlockType { get; set; } = default!;
        public string? ExecutionFingerprint { get; set; }
        public string? HandleInputsFingerprint { get; set; }
        public string? WorkflowDraftFingerprint { get; set; }
        public string? BlockConfigFingerprint { get; set; }
        public StepArtifactRef? Artifact { get; set; }
        public Dictionary<string, object>? HandleInputs { get; set; }
        public Dictionary<string, object>? HandleOutputs { get; set; }
        public List<string>? RoutingDecisions { get; set; }
        public List<string>? Warnings { get; set; }
        public AssertionResult? AssertionResult { get; set; }
        public VerdictSummary? VerdictSummary { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

        /// <summary>
        /// Typed accessor for <see cref="HandleInputs"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetHandleInputsAttribute<T>(string key)
        {
            if (this.HandleInputs == null)
            {
                return default;
            }

            if (!this.HandleInputs.TryGetValue(key, out var value))
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

        /// <summary>
        /// Typed accessor for <see cref="HandleOutputs"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetHandleOutputsAttribute<T>(string key)
        {
            if (this.HandleOutputs == null)
            {
                return default;
            }

            if (!this.HandleOutputs.TryGetValue(key, out var value))
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
