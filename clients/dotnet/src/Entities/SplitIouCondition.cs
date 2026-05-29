namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Intersection-over-Union for split page assignments.</summary>
    /// <remarks>
    /// `expected` uses the split payload shape:
    /// `{"splits": [{"name", "pages"}]}`
    /// </remarks>
    public class SplitIouCondition
    {
        public string? Kind { get; set; } = "split_iou_gte";
        public Dictionary<string, object> Expected { get; set; } = default!;
        public double? Threshold { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

        /// <summary>
        /// Typed accessor for <see cref="Expected"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetExpectedAttribute<T>(string key)
        {
            if (this.Expected == null)
            {
                return default;
            }

            if (!this.Expected.TryGetValue(key, out var value))
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
