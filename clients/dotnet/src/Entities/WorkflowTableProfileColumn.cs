namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a workflow table profile column.</summary>
    public class WorkflowTableProfileColumn
    {
        public string Name { get; set; } = default!;
        public Dictionary<string, object>? JsonSchema { get; set; }
        public long RowCount { get; set; }
        public long NullCount { get; set; }
        public long EmptyCount { get; set; }
        public long DistinctCount { get; set; }
        public object? Min { get; set; }
        public object? Max { get; set; }
        public List<string>? SampleValues { get; set; }
        public bool? IsEstimated { get; set; } = false;

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
    }
}
