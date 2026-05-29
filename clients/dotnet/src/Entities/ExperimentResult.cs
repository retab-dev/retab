namespace Retab
{
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>One experiment result for a single document, addressed by `run_id` and `document_id`.</summary>
    public class ExperimentResult
    {
        public string Id { get; set; } = default!;
        public string RunId { get; set; } = default!;
        public string ExperimentId { get; set; } = default!;
        public string DocumentId { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(PendingWorkflowExperimentResultDiscriminatorConverter))]
        public object Lifecycle { get; set; } = default!;
        public ExperimentResultTiming Timing { get; set; } = default!;
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public ExperimentBlockType BlockType { get; set; }
        public Dictionary<string, object>? HandleInputs { get; set; }
        public StepArtifactRef? Artifact { get; set; }
        public long? Attempt { get; set; }

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
    }
}
