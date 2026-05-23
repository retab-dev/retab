namespace Retab
{
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Payload for a single block output handle.</summary>
    public class HandlePayload
    {

        /// <summary>Type of payload</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public HandlePayloadType Type { get; set; }

        /// <summary>For file handles: document reference</summary>
        public FileRef? Document { get; set; }

        /// <summary>For JSON handles: structured data</summary>
        public object? Data { get; set; }

        /// <summary>For json_ref handles: pointer to artifact-storage JSON body</summary>
        public Dictionary<string, object>? ArtifactRef { get; set; }

        /// <summary>For json_ref handles: lightweight preview of the body</summary>
        public Dictionary<string, object>? Preview { get; set; }

        /// <summary>
        /// Typed accessor for <see cref="ArtifactRef"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetArtifactRefAttribute<T>(string key)
        {
            if (this.ArtifactRef == null)
            {
                return default;
            }

            if (!this.ArtifactRef.TryGetValue(key, out var value))
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
        /// Typed accessor for <see cref="Preview"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetPreviewAttribute<T>(string key)
        {
            if (this.Preview == null)
            {
                return default;
            }

            if (!this.Preview.TryGetValue(key, out var value))
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
