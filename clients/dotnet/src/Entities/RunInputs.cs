namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Input payloads supplied at run creation time.</summary>
    public class RunInputs
    {

        /// <summary>start_document block ID -&gt; input document reference</summary>
        public Dictionary<string, FileRef>? Documents { get; set; }

        /// <summary>start-json block ID -&gt; input JSON data</summary>
        public Dictionary<string, object>? JsonData { get; set; }

        /// <summary>
        /// Typed accessor for <see cref="JsonData"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetJsonDataAttribute<T>(string key)
        {
            if (this.JsonData == null)
            {
                return default;
            }

            if (!this.JsonData.TryGetValue(key, out var value))
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
