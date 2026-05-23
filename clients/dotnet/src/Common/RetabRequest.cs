// @oagen-ignore-file
// Hand-maintained — low-level request object used by URL-builders, bearer-
// override endpoints, and parameter-group serialization paths.

using System.Collections.Generic;
using System.Net.Http;

namespace Retab
{
    /// <summary>Low-level Retab API request descriptor.</summary>
    public class RetabRequest
    {
        /// <summary>The HTTP method.</summary>
        public HttpMethod? Method { get; set; }

        /// <summary>The path component appended to the base URL.</summary>
        public string Path { get; set; } = string.Empty;

        /// <summary>Optional options object whose public properties become query params (GET/DELETE) or body (POST/PUT/PATCH).</summary>
        public BaseOptions? Options { get; set; }

        /// <summary>Pre-serialized request body (overrides <see cref="Options"/>).</summary>
        public object? Body { get; set; }

        /// <summary>Per-request configuration overrides.</summary>
        public RequestOptions? RequestOptions { get; set; }

        /// <summary>Per-operation bearer-token override (e.g. SSO GetProfile).</summary>
        public string? AccessToken { get; set; }

        /// <summary>Additional query parameters merged with the ones derived from <see cref="Options"/>.</summary>
        public Dictionary<string, string>? ExtraQuery { get; set; }

        // ── Convenience setters used by the parameter-group serialization paths ──

        /// <summary>Add a query parameter (used by the parameter-group serialization paths).</summary>
        public void AddQueryParam(string wireName, object? value)
        {
            if (value == null) return;
            this.ExtraQuery ??= new Dictionary<string, string>(System.StringComparer.Ordinal);
            this.ExtraQuery[wireName] = System.Convert.ToString(value, System.Globalization.CultureInfo.InvariantCulture) ?? string.Empty;
        }

        /// <summary>Add a body parameter (used by the parameter-group serialization paths).</summary>
        public void AddBodyParam(string wireName, object? value)
        {
            if (value == null) return;
            this.Body ??= new Dictionary<string, object>(System.StringComparer.Ordinal);
            ((Dictionary<string, object>)this.Body)[wireName] = value;
        }
    }
}
