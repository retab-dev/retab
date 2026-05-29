namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Record of a function block's execution during a workflow run.</summary>
    /// <remarks>
    /// Captures the `inputs` passed to the function, the `output` it returned,
    /// how long it ran (`duration_ms`), and any `error` if execution failed.
    /// </remarks>
    public class FunctionInvocation
    {

        /// <summary>The operation that produced this artifact</summary>
        public string? Operation { get; set; } = "function_invocation";
        public string Id { get; set; } = default!;
        public string RunId { get; set; } = default!;
        public string StepId { get; set; } = default!;
        public Dictionary<string, object>? Inputs { get; set; }
        public object? Output { get; set; }
        public long? DurationMs { get; set; }
        public ErrorDetails? Error { get; set; }

        /// <summary>Timestamp when this artifact was created.</summary>
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
        /// Typed accessor for <see cref="Inputs"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetInputsAttribute<T>(string key)
        {
            if (this.Inputs == null)
            {
                return default;
            }

            if (!this.Inputs.TryGetValue(key, out var value))
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
