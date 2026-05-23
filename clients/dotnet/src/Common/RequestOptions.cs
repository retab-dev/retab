// @oagen-ignore-file
// Hand-maintained — per-request configuration overrides.

using System.Collections.Generic;

namespace Retab
{
    /// <summary>Per-request configuration overrides applied to a single API call.</summary>
    public class RequestOptions
    {
        /// <summary>Override the API key for this request only.</summary>
        public string? ApiKey { get; set; }

        /// <summary>Additional HTTP headers merged with the client defaults.</summary>
        public Dictionary<string, string>? Headers { get; set; }
    }
}
