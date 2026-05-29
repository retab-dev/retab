namespace Retab
{
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>The outcome of applying a workflow YAML definition: whether the workflow was `created`, the changes made, and a `rendered_plan`.</summary>
    public class DeclarativeApplyResponse
    {
        public string WorkflowId { get; set; } = default!;
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public DeclarativeApplyResponseAction Action { get; set; }
        public bool Created { get; set; }
        public long BlockCount { get; set; }
        public long EdgeCount { get; set; }
        public Dictionary<string, object> Diagnostics { get; set; } = default!;
        public string? FormatVersion { get; set; } = "workflows-plan/v1";
        public DeclarativePlanSummary? Summary { get; set; }
        public List<DeclarativePlanResourceChange>? ResourceChanges { get; set; }
        public string? RenderedPlan { get; set; } = "No changes. Workflow spec is up to date.";

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

        /// <summary>
        /// Typed accessor for <see cref="Diagnostics"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetDiagnosticsAttribute<T>(string key)
        {
            if (this.Diagnostics == null)
            {
                return default;
            }

            if (!this.Diagnostics.TryGetValue(key, out var value))
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
