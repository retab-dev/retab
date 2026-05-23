// @oagen-ignore-file
using System.Text.Json.Serialization;

namespace Retab
{
    /// <summary>Pagination cursor metadata used by every list endpoint.</summary>
    public class ListMetadata
    {
        [JsonPropertyName("before")]
        public string? Before { get; set; }

        [JsonPropertyName("after")]
        public string? After { get; set; }
    }
}
