// @oagen-ignore-file
// Hand-maintained — marks an integer-valued Retab enum (e.g. n_consensus =
// 3|5|7) so the JSON converters serialize it as a JSON number instead of a
// quoted string. RetabStringEnumConverterFactory checks for this marker
// because, in System.Text.Json, the globally-registered factory takes
// precedence over a type-level [JsonConverter] attribute.

using System;

namespace Retab
{
    /// <summary>Marks an integer-valued Retab enum for numeric JSON serialization.</summary>
    [AttributeUsage(AttributeTargets.Enum, AllowMultiple = false, Inherited = false)]
    public sealed class RetabNumberEnumAttribute : Attribute
    {
    }
}
