namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Public live workflow block object.</summary>
    public class WorkflowBlock
    {
        public string Id { get; set; } = default!;

        /// <summary>Foreign key to workflow</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>Block type (extract, parse, classifier, etc.)</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public WorkflowBlockType Type { get; set; }

        /// <summary>Display label for the block</summary>
        public string? Label { get; set; }

        /// <summary>X position on canvas</summary>
        public double? PositionX { get; set; }

        /// <summary>Y position on canvas</summary>
        public double? PositionY { get; set; }

        /// <summary>Block width for resizable blocks</summary>
        public double? Width { get; set; }

        /// <summary>Block height for resizable blocks</summary>
        public double? Height { get; set; }

        /// <summary>Block-specific configuration</summary>
        public Dictionary<string, object>? Config { get; set; }

        /// <summary>ID of parent container (while_loop, for_each)</summary>
        public string? ParentId { get; set; }
        public DateTimeOffset UpdatedAt { get; set; }

        /// <summary>Internal graph-derived schema sidecar.</summary>
        public Dictionary<string, object>? ResolvedSchemas { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

        /// <summary>
        /// Typed accessor for <see cref="Config"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetConfigAttribute<T>(string key)
        {
            if (this.Config == null)
            {
                return default;
            }

            if (!this.Config.TryGetValue(key, out var value))
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
        /// Typed accessor for <see cref="ResolvedSchemas"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetResolvedSchemasAttribute<T>(string key)
        {
            if (this.ResolvedSchemas == null)
            {
                return default;
            }

            if (!this.ResolvedSchemas.TryGetValue(key, out var value))
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
