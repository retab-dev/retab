// @oagen-ignore-file
// Hand-maintained — Newtonsoft.Json converter that maps snake_case wire
// values to PascalCase C# enum members. The emitter attaches this to every
// generated enum via [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))].

using System;
using System.Reflection;
using System.Runtime.Serialization;
using Newtonsoft.Json;

namespace Retab
{
    /// <summary>Newtonsoft.Json converter for Retab enums.</summary>
    public class RetabNewtonsoftStringEnumConverter : JsonConverter
    {
        public override bool CanConvert(Type objectType)
        {
            var t = Nullable.GetUnderlyingType(objectType) ?? objectType;
            return t.IsEnum;
        }

        public override void WriteJson(JsonWriter writer, object? value, JsonSerializer serializer)
        {
            if (value == null) { writer.WriteNull(); return; }
            writer.WriteValue(ToWire(value));
        }

        public override object? ReadJson(JsonReader reader, Type objectType, object? existingValue, JsonSerializer serializer)
        {
            var t = Nullable.GetUnderlyingType(objectType) ?? objectType;
            if (reader.TokenType == JsonToken.Null) return null;
            var wire = reader.Value?.ToString();
            if (wire == null) return null;
            return FromWire(t, wire);
        }

        private static string ToWire(object enumValue)
        {
            var type = enumValue.GetType();
            var name = enumValue.ToString();
            if (name == null) return string.Empty;
            var field = type.GetField(name);
            var attr = field?.GetCustomAttribute<EnumMemberAttribute>();
            return attr?.Value ?? ToSnakeCase(name);
        }

        private static object FromWire(Type enumType, string wire)
        {
            foreach (var field in enumType.GetFields(BindingFlags.Public | BindingFlags.Static))
            {
                var attr = field.GetCustomAttribute<EnumMemberAttribute>();
                if (attr != null && attr.Value == wire) return field.GetValue(null)!;
                if (string.Equals(field.Name, wire, StringComparison.OrdinalIgnoreCase)) return field.GetValue(null)!;
                if (string.Equals(ToSnakeCase(field.Name), wire, StringComparison.OrdinalIgnoreCase))
                    return field.GetValue(null)!;
            }
            // Unknown wire value → first declared member (typically a stable sentinel).
            return Enum.GetValues(enumType).GetValue(0)!;
        }

        private static string ToSnakeCase(string name)
        {
            var sb = new System.Text.StringBuilder(name.Length + 4);
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
}
