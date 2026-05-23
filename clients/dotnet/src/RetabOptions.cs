// @oagen-ignore-file
// Hand-maintained — constructor options for the Retab client.

using System;
using System.Net.Http;

namespace Retab
{
    /// <summary>Construction options for <see cref="Retab"/>.</summary>
    public class RetabOptions
    {
        /// <summary>The Retab API key (required).</summary>
        public string? ApiKey { get; set; }

        /// <summary>The Retab client-id (optional, only required by some integrations).</summary>
        public string? ClientId { get; set; }

        /// <summary>The base URL for the Retab API. Defaults to <c>https://api.retab.com</c>.</summary>
        public Uri? BaseUrl { get; set; }

        /// <summary>An optional pre-configured HTTP client. When unset, the SDK constructs its own.</summary>
        public HttpClient? HttpClient { get; set; }
    }
}
