namespace Retab
{
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents a declarative plan response.</summary>
    public class DeclarativePlanResponse
    {
        public string WorkflowId { get; set; } = default!;
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public DeclarativeApplyResponseAction Action { get; set; }
        public long BlockCount { get; set; }
        public long EdgeCount { get; set; }
        public Dictionary<string, object> Diagnostics { get; set; } = default!;
        public string? FormatVersion { get; set; }
        public DeclarativePlanSummary? Summary { get; set; }
        public List<DeclarativePlanResourceChange>? ResourceChanges { get; set; }
        public string? RenderedPlan { get; set; }

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
