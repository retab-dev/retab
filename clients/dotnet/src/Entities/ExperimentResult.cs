namespace Retab
{
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Run-scoped per-document experiment result.</summary>
    /// <remarks>
    /// The storage row is still named ``experiment_jobs`` internally, but the
    /// public contract is a result row addressed by ``run_id`` + ``document_id``.
    /// </remarks>
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
        public bool? IsPlaceholder { get; set; }

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
