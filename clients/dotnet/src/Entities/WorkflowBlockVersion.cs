namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Immutable block snapshot derived from a workflow version.</summary>
    public class WorkflowBlockVersion
    {

        /// <summary>Public content-addressed block version ID</summary>
        public string Id { get; set; } = default!;

        /// <summary>Stable logical block ID</summary>
        public string BlockId { get; set; } = default!;

        /// <summary>Source workflow ID</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>Organization ID for data isolation</summary>
        public string OrganizationId { get; set; } = default!;

        /// <summary>Customer environment ID for data isolation</summary>
        public string EnvironmentId { get; set; } = default!;

        /// <summary>Workflow version this block version belongs to</summary>
        public string WorkflowVersionId { get; set; } = default!;
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public WorkflowBlockType Type { get; set; }
        public string? Label { get; set; } = "";
        public double? PositionX { get; set; }
        public double? PositionY { get; set; }
        public double? Width { get; set; }
        public double? Height { get; set; }
        public string? ParentId { get; set; }
        public Dictionary<string, object>? Config { get; set; }
        public Dictionary<string, string>? FieldRefSnapshot { get; set; }
        public Dictionary<string, object>? ResolvedSchemas { get; set; }

        /// <summary>Stable SHA-256 hash of the executable block config</summary>
        public string? ConfigHash { get; set; } = "";
        public DateTimeOffset? CreatedAt { get; set; }

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
