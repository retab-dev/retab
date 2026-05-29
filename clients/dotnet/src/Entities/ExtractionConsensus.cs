namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents an extraction consensus.</summary>
    public class ExtractionConsensus
    {

        /// <summary>Alternative extraction vote outputs used to build the consolidated result.</summary>
        public List<Dictionary<string, object>>? Choices { get; set; }

        /// <summary>Consensus likelihood tree mirroring the extraction output. Scalar leaves carry per-value voter-agreement in [0, 1]; list leaves carry one entry per matched list item.</summary>
        public Dictionary<string, object>? Likelihoods { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

        /// <summary>
        /// Typed accessor for <see cref="Likelihoods"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetLikelihoodsAttribute<T>(string key)
        {
            if (this.Likelihoods == null)
            {
                return default;
            }

            if (!this.Likelihoods.TryGetValue(key, out var value))
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
