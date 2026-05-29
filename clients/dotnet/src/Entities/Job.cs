namespace Retab
{
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Core Job object following OpenAI-style specification.</summary>
    /// <remarks>
    /// Represents a single asynchronous job that can be polled for status
    /// and result retrieval.
    /// </remarks>
    public class Job
    {

        /// <summary>Opaque job id (server-generated ``job_&lt;nanoid&gt;``).</summary>
        public string Id { get; set; } = default!;
        public string? Object { get; set; }
        public JobStatus? Status { get; set; }
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public CreateJobRequestEndpoint Endpoint { get; set; }
        public JobError? Error { get; set; }
        public List<JobError>? Warnings { get; set; }
        public string? CreatedAt { get; set; }
        public string? StartedAt { get; set; }
        public string? CompletedAt { get; set; }
        public string? ExpiresAt { get; set; }
        public Dictionary<string, string>? Metadata { get; set; }
        public bool? Cancelled { get; set; }
        public long? AttemptCount { get; set; }
        public string? LastAttemptAt { get; set; }
        public string? LastFailureCode { get; set; }
        public Dictionary<string, object>? Request { get; set; }
        public JobResponse? Response { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

        /// <summary>
        /// Typed accessor for <see cref="Request"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetRequestAttribute<T>(string key)
        {
            if (this.Request == null)
            {
                return default;
            }

            if (!this.Request.TryGetValue(key, out var value))
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
