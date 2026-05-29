// @oagen-ignore-file
// Hand-maintained — Newtonsoft.Json converter for integer-valued Retab enums
// (e.g. n_consensus = 3|5|7). These must serialize as JSON numbers, not quoted
// strings. The emitter attaches this to integer enums via
// [JsonConverter(typeof(RetabNewtonsoftNumberEnumConverter))]. Each enum member
// carries its wire integer as its underlying value; the Unknown sentinel uses
// int.MinValue so it never collides with a real wire value.

using System;
using Newtonsoft.Json;

namespace Retab
{
    /// <summary>Newtonsoft.Json converter for integer-valued Retab enums.</summary>
    public class RetabNewtonsoftNumberEnumConverter : JsonConverter
    {
        public override bool CanConvert(Type objectType)
        {
            var t = Nullable.GetUnderlyingType(objectType) ?? objectType;
            return t.IsEnum;
        }

        public override void WriteJson(JsonWriter writer, object? value, JsonSerializer serializer)
        {
            if (value == null) { writer.WriteNull(); return; }
            writer.WriteValue(Convert.ToInt64(value));
        }

        public override object? ReadJson(JsonReader reader, Type objectType, object? existingValue, JsonSerializer serializer)
        {
            var t = Nullable.GetUnderlyingType(objectType) ?? objectType;
            if (reader.TokenType == JsonToken.Null) return null;
            var raw = Convert.ToInt64(reader.Value);
            foreach (var v in Enum.GetValues(t))
            {
                if (Convert.ToInt64(v) == raw) return v;
            }
            // Unknown numeric value → first declared member (the Unknown sentinel,
            // which uses int.MinValue and therefore sorts first).
            return Enum.GetValues(t).GetValue(0)!;
        }
    }
}
