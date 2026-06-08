namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Represents a workflow table.</summary>
    public class WorkflowTable
    {
        public string Id { get; set; } = default!;
        public string Name { get; set; } = default!;
        public string Filename { get; set; } = default!;

        /// <summary>Project that owns this table. Null only on legacy rows that predate the project backfill.</summary>
        public string? ProjectId { get; set; }
        public string? SourceFileId { get; set; } = "";
        public string? SnapshotFileId { get; set; } = "";
        public long RowCount { get; set; }
        public List<WorkflowTableColumn>? Columns { get; set; }
        public List<Dictionary<string, object>>? SampleRows { get; set; }
        public Dictionary<string, object>? Metadata { get; set; }
        public string? UploadedByUserId { get; set; }
        public DateTimeOffset? CreatedAt { get; set; }
        public DateTimeOffset? UpdatedAt { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

        /// <summary>
        /// Typed accessor for <see cref="Metadata"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetMetadataAttribute<T>(string key)
        {
            if (this.Metadata == null)
            {
                return default;
            }

            if (!this.Metadata.TryGetValue(key, out var value))
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
