namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>An extraction produced by a workflow run, tagged with its artifact `operation` and creation time.</summary>
    public class ExtractionWorkflowArtifact
    {

        /// <summary>Unique identifier of the extraction</summary>
        public string Id { get; set; } = default!;

        /// <summary>Information about the extracted file</summary>
        public FileRef File { get; set; } = default!;

        /// <summary>Model used for the extraction</summary>
        public string Model { get; set; } = default!;

        /// <summary>JSON schema used for the extraction</summary>
        public Dictionary<string, object> JsonSchema { get; set; } = default!;

        /// <summary>Number of consensus votes used</summary>
        public long? NConsensus { get; set; }

        /// <summary>DPI used to render document images</summary>
        public long? ImageResolutionDpi { get; set; }

        /// <summary>Free-form instructions supplied with the extraction request.</summary>
        public string? Instructions { get; set; }

        /// <summary>The extracted structured data</summary>
        public Dictionary<string, object> Output { get; set; } = default!;

        /// <summary>Lifecycle status. The synchronous path returns 'completed'. Background runs progress pending -&gt; queued -&gt; in_progress -&gt; completed | failed | cancelled.</summary>
        public ClassificationStatus? Status { get; set; }

        /// <summary>Error details when a background run fails; null otherwise. Always present so consumers can read it without an existence check.</summary>
        public PrimitiveError? Error { get; set; }

        /// <summary>Consensus metadata for multi-vote extraction runs</summary>
        public ExtractionConsensus? Consensus { get; set; }
        public Dictionary<string, string>? Metadata { get; set; }

        /// <summary>Usage information for the extraction</summary>
        public RetabUsage? Usage { get; set; }

        /// <summary>Timestamp when this artifact was created.</summary>
        public DateTimeOffset? CreatedAt { get; set; }

        /// <summary>The operation that produced this artifact</summary>
        public string? Operation { get; set; } = "extraction";

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

        /// <summary>
        /// Typed accessor for <see cref="JsonSchema"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetJsonSchemaAttribute<T>(string key)
        {
            if (this.JsonSchema == null)
            {
                return default;
            }

            if (!this.JsonSchema.TryGetValue(key, out var value))
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
        /// Typed accessor for <see cref="Output"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetOutputAttribute<T>(string key)
        {
            if (this.Output == null)
            {
                return default;
            }

            if (!this.Output.TryGetValue(key, out var value))
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
