// @oagen-ignore-file
// Hand-maintained — the central Retab API client. The companion file
// Retab.Generated.cs adds one accessor property per top-level resource
// service and IS regenerated from the spec on every run.

using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Reflection;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading;
using System.Threading.Tasks;
using Newtonsoft.Json;
using Newtonsoft.Json.Serialization;

namespace Retab
{
    /// <summary>Retab API client. Pass an API key (or a fully-configured
    /// <see cref="RetabOptions"/>) to construct.</summary>
    public partial class Retab
    {
        private static readonly string SdkVersion =
            typeof(Retab).Assembly.GetCustomAttribute<AssemblyInformationalVersionAttribute>()?.InformationalVersion
            ?? "0.0.0";

        /// <summary>The configured HTTP client.</summary>
        public HttpClient HttpClient { get; }

        /// <summary>The configured API key (sent with the Api-Key header).</summary>
        public string ApiKey { get; }

        /// <summary>The configured client-id, if any. Optional for Retab; required for some integrations.</summary>
        public string? ClientId { get; }

        /// <summary>The configured base URL. Defaults to <c>https://api.retab.com</c>.</summary>
        public Uri BaseUrl { get; }

        /// <summary>Shared <see cref="JsonSerializerOptions"/> used by every Retab request/response.</summary>
        public static readonly JsonSerializerOptions JsonOptions = new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
            DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,
            PropertyNameCaseInsensitive = true,
            Converters = { new RetabStringEnumConverterFactory() },
        };

        private static readonly JsonSerializerSettings NewtonsoftJsonSettings = new JsonSerializerSettings
        {
            ContractResolver = new DefaultContractResolver
            {
                NamingStrategy = new SnakeCaseNamingStrategy
                {
                    ProcessDictionaryKeys = false,
                    OverrideSpecifiedNames = false,
                },
            },
        };

        /// <summary>Construct a Retab client with the supplied API key.</summary>
        public Retab(string apiKey) : this(new RetabOptions { ApiKey = apiKey }) { }

        /// <summary>Construct a Retab client from a configured <see cref="RetabOptions"/>.</summary>
        public Retab(RetabOptions options)
        {
            if (options == null) throw new ArgumentNullException(nameof(options));
            if (string.IsNullOrWhiteSpace(options.ApiKey))
                throw new ArgumentException("Retab requires a non-empty API key.", nameof(options));

            this.ApiKey = options.ApiKey!;
            this.ClientId = options.ClientId;
            this.BaseUrl = NormalizeBaseUrl(options.BaseUrl ?? new Uri("https://api.retab.com"));
            this.HttpClient = options.HttpClient ?? new HttpClient();

            this.HttpClient.DefaultRequestHeaders.UserAgent.TryParseAdd($"retab-dotnet/{SdkVersion}");
        }

        /// <summary>Returns the configured <see cref="ClientId"/> or throws if unset.</summary>
        public string RequireClientId()
        {
            if (string.IsNullOrEmpty(this.ClientId))
                throw new InvalidOperationException("Retab client_id is required for this operation but was not configured.");
            return this.ClientId!;
        }

        /// <summary>Issue a request whose body / path / overrides are already on a <see cref="RetabRequest"/>.</summary>
        public virtual async Task<TResult> MakeAPIRequest<TResult>(RetabRequest request, CancellationToken cancellationToken)
        {
            using var httpRequest = BuildHttpRequest(request);
            using var response = await this.HttpClient.SendAsync(httpRequest, cancellationToken).ConfigureAwait(false);
            await EnsureSuccessAsync(response).ConfigureAwait(false);
            return await DeserializeAsync<TResult>(response, cancellationToken).ConfigureAwait(false);
        }

        /// <summary>
        /// Issue a request and deserialize the response through an explicit
        /// <paramref name="converter"/>. Used for discriminated-union response
        /// roots, where <typeparamref name="TResult"/> is <see cref="object"/>
        /// and the converter dispatches to the concrete variant by its
        /// discriminator property so no variant-specific field is lost.
        /// </summary>
        public virtual async Task<TResult> MakeAPIRequest<TResult>(RetabRequest request, Newtonsoft.Json.JsonConverter converter, CancellationToken cancellationToken)
        {
            using var httpRequest = BuildHttpRequest(request);
            using var response = await this.HttpClient.SendAsync(httpRequest, cancellationToken).ConfigureAwait(false);
            await EnsureSuccessAsync(response).ConfigureAwait(false);
            return await DeserializeAsync<TResult>(response, converter, cancellationToken).ConfigureAwait(false);
        }

        /// <summary>Issue a request and discard the response body.</summary>
        public virtual async Task MakeRawAPIRequest(RetabRequest request, CancellationToken cancellationToken)
        {
            using var httpRequest = BuildHttpRequest(request);
            using var response = await this.HttpClient.SendAsync(httpRequest, cancellationToken).ConfigureAwait(false);
            await EnsureSuccessAsync(response).ConfigureAwait(false);
        }

        /// <summary>Compute the fully-qualified URI for a <see cref="RetabRequest"/> without issuing it.</summary>
        public virtual Uri BuildRequestUri(RetabRequest request)
        {
            var basePart = this.BaseUrl.ToString().TrimEnd('/');
            var pathPart = request.Path.StartsWith("/") ? request.Path : "/" + request.Path;
            var builder = new UriBuilder(basePart + pathPart);
            var method = request.Method ?? HttpMethod.Get;
            var query = IsBodyMethod(method)
                ? new Dictionary<string, string>(StringComparer.Ordinal)
                : ExtractQueryParams(request.Options);
            if (request.ExtraQuery != null)
            {
                foreach (var kv in request.ExtraQuery) query[kv.Key] = kv.Value;
            }
            builder.Query = SerializeQuery(query);
            return builder.Uri;
        }

        private HttpRequestMessage BuildHttpRequest(RetabRequest request)
        {
            var uri = BuildRequestUri(request);
            var httpRequest = new HttpRequestMessage(request.Method ?? HttpMethod.Get, uri);
            var requestApiKey = request.RequestOptions?.ApiKey ?? this.ApiKey;

            if (!string.IsNullOrWhiteSpace(requestApiKey) && !HasHeader(request.RequestOptions?.Headers, "Api-Key"))
            {
                httpRequest.Headers.TryAddWithoutValidation("Api-Key", requestApiKey);
            }

            if (!string.IsNullOrEmpty(request.AccessToken))
            {
                httpRequest.Headers.Authorization = new AuthenticationHeaderValue("Bearer", request.AccessToken);
            }

            if (request.MultipartBody != null)
            {
                // multipart/form-data body (e.g. table create/replace). The
                // MultipartFormDataContent sets its own Content-Type header,
                // including the boundary, so we must NOT overwrite it with a
                // JSON Content-Type. The gateway rejects a JSON `file` string
                // with 400 "Expected a multipart/form-data file upload".
                httpRequest.Content = request.MultipartBody;
            }
            else if (request.Body != null)
            {
                var json = System.Text.Json.JsonSerializer.Serialize(request.Body, request.Body.GetType(), JsonOptions);
                httpRequest.Content = new StringContent(json, Encoding.UTF8, "application/json");
            }
            else if (request.Options != null && IsBodyMethod(httpRequest.Method))
            {
                var json = System.Text.Json.JsonSerializer.Serialize(request.Options, request.Options.GetType(), JsonOptions);
                httpRequest.Content = new StringContent(json, Encoding.UTF8, "application/json");
            }

            if (request.RequestOptions?.Headers != null)
            {
                foreach (var kv in request.RequestOptions.Headers)
                {
                    httpRequest.Headers.TryAddWithoutValidation(kv.Key, kv.Value);
                }
            }

            return httpRequest;
        }

        private static bool HasHeader(Dictionary<string, string>? headers, string name)
        {
            if (headers == null) return false;
            return headers.Keys.Any(key => string.Equals(key, name, StringComparison.OrdinalIgnoreCase));
        }

        private static Uri NormalizeBaseUrl(Uri baseUrl)
        {
            var builder = new UriBuilder(baseUrl);
            var path = builder.Path.TrimEnd('/');
            if (path.EndsWith("/v1", StringComparison.OrdinalIgnoreCase))
            {
                var withoutVersion = path.Substring(0, path.Length - "/v1".Length);
                builder.Path = string.IsNullOrEmpty(withoutVersion) ? "/" : withoutVersion;
            }
            return builder.Uri;
        }

        private static bool IsBodyMethod(HttpMethod method)
        {
            return method == HttpMethod.Post
                || method == HttpMethod.Put
                || method.Method == "PATCH";
        }

        private static async Task EnsureSuccessAsync(HttpResponseMessage response)
        {
            if (response.IsSuccessStatusCode) return;
            var body = response.Content != null ? await response.Content.ReadAsStringAsync().ConfigureAwait(false) : string.Empty;
            throw RetabException.From((int)response.StatusCode, body);
        }

        private static async Task<TResult> DeserializeAsync<TResult>(HttpResponseMessage response, CancellationToken cancellationToken)
        {
            if (response.StatusCode == System.Net.HttpStatusCode.NoContent || response.Content == null)
            {
                return default!;
            }
            var body = await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
            var result = JsonConvert.DeserializeObject<TResult>(body, NewtonsoftJsonSettings);
            return result!;
        }

        private static async Task<TResult> DeserializeAsync<TResult>(HttpResponseMessage response, Newtonsoft.Json.JsonConverter converter, CancellationToken cancellationToken)
        {
            if (response.StatusCode == System.Net.HttpStatusCode.NoContent || response.Content == null)
            {
                return default!;
            }
            var body = await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
            // Deserialize through the discriminated-union converter without
            // mutating the shared NewtonsoftJsonSettings: clone the snake-case
            // contract resolver settings and add the converter for this call.
            var settings = new JsonSerializerSettings
            {
                ContractResolver = NewtonsoftJsonSettings.ContractResolver,
            };
            settings.Converters.Add(converter);
            var result = JsonConvert.DeserializeObject<TResult>(body, settings);
            return result!;
        }

        private static Dictionary<string, string> ExtractQueryParams(BaseOptions? options)
        {
            var result = new Dictionary<string, string>(StringComparer.Ordinal);
            if (options == null) return result;
            var type = options.GetType();
            foreach (var prop in type.GetProperties(BindingFlags.Public | BindingFlags.Instance))
            {
                if (!prop.CanRead) continue;
                if (prop.GetSetMethod()?.IsPublic != true) continue;
                var value = prop.GetValue(options);
                if (value == null) continue;
                var wireName = GetWireName(prop);
                if (value is System.Collections.IEnumerable enumerable && value is not string)
                {
                    var parts = new List<string>();
                    foreach (var item in enumerable)
                    {
                        if (item == null) continue;
                        parts.Add(Convert.ToString(item, System.Globalization.CultureInfo.InvariantCulture) ?? string.Empty);
                    }
                    if (parts.Count > 0) result[wireName] = string.Join(",", parts);
                }
                else
                {
                    var str = Convert.ToString(value, System.Globalization.CultureInfo.InvariantCulture);
                    if (!string.IsNullOrEmpty(str)) result[wireName] = str;
                }
            }
            return result;
        }

        private static string GetWireName(PropertyInfo prop)
        {
            var json = prop.GetCustomAttribute<JsonPropertyNameAttribute>();
            if (json != null) return json.Name;
            return ToSnakeCase(prop.Name);
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

        private static string SerializeQuery(Dictionary<string, string> q)
        {
            if (q.Count == 0) return string.Empty;
            return string.Join("&", q.Select(kv => $"{Uri.EscapeDataString(kv.Key)}={Uri.EscapeDataString(kv.Value)}"));
        }
    }
}
