// @oagen-ignore-file
// Hand-maintained — System.Text.Json converter factory matching the
// Newtonsoft.Json converter behavior for Retab enums.

using System;
using System.Reflection;
using System.Runtime.Serialization;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Retab
{
    /// <summary>System.Text.Json converter factory for Retab enums. This is
    /// registered globally in <c>Retab.JsonOptions.Converters</c>. Because
    /// converters in the options collection take precedence over a type-level
    /// <c>[JsonConverter]</c> attribute in System.Text.Json, this single
    /// factory must handle both string- and integer-valued enums: it inspects
    /// the <see cref="RetabNumberEnumAttribute"/> marker to choose between a
    /// numeric and a string converter.</summary>
    public class RetabStringEnumConverterFactory : JsonConverterFactory
    {
        public override bool CanConvert(Type typeToConvert)
        {
            return typeToConvert.IsEnum;
        }

        public override JsonConverter CreateConverter(Type typeToConvert, JsonSerializerOptions options)
        {
            var t = Nullable.GetUnderlyingType(typeToConvert) ?? typeToConvert;
            var isNumber = t.IsDefined(typeof(RetabNumberEnumAttribute), false);
            var converterType = (isNumber ? typeof(RetabNumberEnumConverter<>) : typeof(RetabEnumConverter<>))
                .MakeGenericType(t);
            return (JsonConverter)Activator.CreateInstance(converterType)!;
        }

        private class RetabEnumConverter<T> : JsonConverter<T> where T : struct, Enum
        {
            public override T Read(ref Utf8JsonReader reader, Type typeToConvert, JsonSerializerOptions options)
            {
                if (reader.TokenType == JsonTokenType.Null) return default;
                var wire = reader.GetString();
                if (wire == null) return default;
                foreach (var field in typeof(T).GetFields(BindingFlags.Public | BindingFlags.Static))
                {
                    var attr = field.GetCustomAttribute<EnumMemberAttribute>();
                    if (attr != null && attr.Value == wire) return (T)field.GetValue(null)!;
                    if (string.Equals(field.Name, wire, StringComparison.OrdinalIgnoreCase)) return (T)field.GetValue(null)!;
                    if (string.Equals(ToSnakeCase(field.Name), wire, StringComparison.OrdinalIgnoreCase))
                        return (T)field.GetValue(null)!;
                }
                return default;
            }

            public override void Write(Utf8JsonWriter writer, T value, JsonSerializerOptions options)
            {
                var name = value.ToString();
                var field = typeof(T).GetField(name);
                var attr = field?.GetCustomAttribute<EnumMemberAttribute>();
                writer.WriteStringValue(attr?.Value ?? ToSnakeCase(name));
            }

            private static string ToSnakeCase(string name)
            {
                var sb = new StringBuilder(name.Length + 4);
                for (int i = 0; i < name.Length; i++)
                {
                    var c = name[i];
                    if (char.IsUpper(c))
                    {
                        if (i > 0) sb.Append('_');
                        sb.Append(char.ToLowerInvariant(c));
                    }
                    else sb.Append(c);
                }
                return sb.ToString();
            }
        }

        // Integer-valued enums (e.g. n_consensus = 3|5|7) serialize as JSON
        // numbers. Each member carries its wire integer as its underlying value;
        // the Unknown sentinel uses int.MinValue and therefore sorts first, so
        // unrecognized wire values fall back to it.
        private class RetabNumberEnumConverter<T> : JsonConverter<T> where T : struct, Enum
        {
            public override T Read(ref Utf8JsonReader reader, Type typeToConvert, JsonSerializerOptions options)
            {
                if (reader.TokenType == JsonTokenType.Null) return default;
                long raw = reader.TokenType == JsonTokenType.Number ? reader.GetInt64() : 0;
                foreach (var v in Enum.GetValues<T>())
                {
                    if (Convert.ToInt64(v) == raw) return v;
                }
                var all = Enum.GetValues<T>();
                return all.Length > 0 ? all[0] : default;
            }

            public override void Write(Utf8JsonWriter writer, T value, JsonSerializerOptions options)
            {
                writer.WriteNumberValue(Convert.ToInt64(value));
            }
        }
    }
}
