namespace Retab
{
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents a workflow config block.</summary>
    public class WorkflowConfigBlock
    {
        public string? Id { get; set; }
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public WorkflowBlockType Type { get; set; }
        public WorkflowBlockPosition Position { get; set; } = default!;
        public string Label { get; set; } = default!;

        /// <summary>Block-specific configuration</summary>
        public Dictionary<string, object>? Config { get; set; }

        /// <summary>Derived schema transport sidecar for UI/runtime consumers. Not authored config.</summary>
        public Dictionary<string, object>? ResolvedSchemas { get; set; }

        /// <summary>Block width for resizable blocks</summary>
        public double? Width { get; set; }

        /// <summary>Block height for resizable blocks</summary>
        public double? Height { get; set; }

        /// <summary>ID of parent container block (while_loop, for_each)</summary>
        public string? ParentId { get; set; }

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
