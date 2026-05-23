// @oagen-ignore-file
using System.Text.Json.Serialization;

namespace Retab
{
    /// <summary>Base class for paginated list options (before / after / limit / order).</summary>
    public abstract class ListOptions : BaseOptions
    {
        [JsonPropertyName("before")]
        public string? Before { get; set; }

        [JsonPropertyName("after")]
        public string? After { get; set; }

        [JsonPropertyName("limit")]
        public int? Limit { get; set; }

        [JsonPropertyName("order")]
        public string? Order { get; set; }
    }
}
