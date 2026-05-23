namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a validation error.</summary>
    public class ValidationError
    {
        public List<OneOf.OneOf<string, long>> Loc { get; set; } = default!;
        public string Msg { get; set; } = default!;
        public string Type { get; set; } = default!;
        public object? Input { get; set; }
        public Dictionary<string, object>? Ctx { get; set; }

        /// <summary>
        /// Typed accessor for <see cref="Ctx"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetCtxAttribute<T>(string key)
        {
            if (this.Ctx == null)
            {
                return default;
            }

            if (!this.Ctx.TryGetValue(key, out var value))
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
