namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Public API shape for one immutable review version.</summary>
    public class ReviewVersion
    {
        public string Id { get; set; } = default!;
        public string ReviewId { get; set; } = default!;
        public string? ParentId { get; set; }

        /// <summary>Actor that created the version.</summary>
        public Actor Author { get; set; } = default!;
        public Dictionary<string, object> Snapshot { get; set; } = default!;
        public string? Note { get; set; }

        /// <summary>When the review version was created.</summary>
        public DateTimeOffset CreatedAt { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

        /// <summary>
        /// Typed accessor for <see cref="Snapshot"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetSnapshotAttribute<T>(string key)
        {
            if (this.Snapshot == null)
            {
                return default;
            }

            if (!this.Snapshot.TryGetValue(key, out var value))
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
