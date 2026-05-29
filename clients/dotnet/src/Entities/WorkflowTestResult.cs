namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a workflow test result.</summary>
    public class WorkflowTestResult
    {
        public string Id { get; set; } = default!;
        public string? RunId { get; set; }
        public string TestId { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(PendingWorkflowTestRunDiscriminatorConverter))]
        public object? Lifecycle { get; set; }
        public ExperimentRunTiming? Timing { get; set; }

        /// <summary>Verdict label populated only when the underlying test reaches a terminal lifecycle state and the verdict could be determined. Execution-error details flow through `error` (an `ErrorDetails` envelope), not through this enum.</summary>
        public AssertionOutcome? Verdict { get; set; }
        public string WorkflowId { get; set; } = default!;
        public WorkflowTestBlockTarget Target { get; set; } = default!;
        public string? ExecutionFingerprint { get; set; }
        public string? HandleInputsFingerprint { get; set; }
        public string? WorkflowDraftFingerprint { get; set; }
        public string? BlockConfigFingerprint { get; set; }
        [Newtonsoft.Json.JsonConverter(typeof(ManualWorkflowTestSourceDiscriminatorConverter))]
        public object Source { get; set; } = default!;
        public Dictionary<string, object>? Outputs { get; set; }
        public List<string>? RoutingDecision { get; set; }
        public List<string>? Warnings { get; set; }
        public ErrorDetails? Error { get; set; }
        public bool? Skipped { get; set; }
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
        /// Typed accessor for <see cref="Outputs"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetOutputsAttribute<T>(string key)
        {
            if (this.Outputs == null)
            {
                return default;
            }

            if (!this.Outputs.TryGetValue(key, out var value))
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
