namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>A document blueprint generated from an uploaded file.</summary>
    public class FileBlueprint
    {
        public string? Object { get; set; } = "file.blueprint";

        /// <summary>Unique identifier of the file blueprint.</summary>
        public string Id { get; set; } = default!;

        /// <summary>Information about the analyzed file.</summary>
        public BlockExecFileRef File { get; set; } = default!;

        /// <summary>User intent supplied with the blueprint request.</summary>
        public string? Intent { get; set; }

        /// <summary>The generated Document Blueprint payload.</summary>
        public Dictionary<string, object>? Output { get; set; }

        /// <summary>Lifecycle status. The synchronous path returns 'completed'. Background runs progress pending -&gt; queued -&gt; in_progress -&gt; completed | failed | cancelled.</summary>
        public ClassificationStatus? Status { get; set; }

        /// <summary>Error details when a background run fails; null otherwise. Always present so consumers can read it without an existence check.</summary>
        public PrimitiveError? Error { get; set; }
        public DateTimeOffset? CreatedAt { get; set; }
        public DateTimeOffset? StartedAt { get; set; }
        public DateTimeOffset? CompletedAt { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

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
