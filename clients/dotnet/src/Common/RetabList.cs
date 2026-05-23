// @oagen-ignore-file
// Hand-maintained — paginated list wrapper for every list endpoint.

using System.Collections.Generic;
using System.Text.Json.Serialization;

namespace Retab
{
    /// <summary>Paginated list of <typeparamref name="T"/> values returned by list endpoints.</summary>
    public class RetabList<T>
    {
        [JsonPropertyName("data")]
        public List<T>? Data { get; set; }

        [JsonPropertyName("list_metadata")]
        public ListMetadata? ListMetadata { get; set; }

        [JsonPropertyName("object")]
        public string? Object { get; set; }
    }
}
