// @oagen-ignore-file
// Hand-maintained — base class for every generated XxxService. Centralises
// the HTTP method helpers (GetAsync / PostAsync / etc.) the resource emitter
// inlines as one-liners. The inline RetabRequest path (URL-builders, bearer
// overrides, parameter groups) goes through Retab.MakeAPIRequest /
// MakeRawAPIRequest / BuildRequestUri instead.

using System.Collections.Generic;
using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;

namespace Retab
{
    /// <summary>Base class for every generated resource service.</summary>
    public abstract class Service
    {
        /// <summary>The Retab client this service is bound to.</summary>
        protected internal Retab Client { get; }

        /// <summary>Bind a service to an explicit <see cref="Retab"/> client.</summary>
        protected Service(Retab client) { this.Client = client; }

        // ── Verb helpers — match the names the resource emitter calls ──

        protected Task<TResult> GetAsync<TResult>(string path, BaseOptions? options, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, cancellationToken);

        // Converter-threading overloads: a discriminated-union response root is
        // deserialized as 'object' through the variant-dispatching converter so
        // the concrete variant survives instead of collapsing to one class.
        protected Task<TResult> GetAsync<TResult>(string path, BaseOptions? options, Newtonsoft.Json.JsonConverter converter, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, converter, cancellationToken);

        protected Task<TResult> PostAsync<TResult>(string path, BaseOptions? options, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, cancellationToken);

        protected Task<TResult> PostAsync<TResult>(string path, BaseOptions? options, Newtonsoft.Json.JsonConverter converter, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, converter, cancellationToken);

        protected Task<TResult> PutAsync<TResult>(string path, BaseOptions? options, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = HttpMethod.Put,
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, cancellationToken);

        protected Task<TResult> PutAsync<TResult>(string path, BaseOptions? options, Newtonsoft.Json.JsonConverter converter, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = HttpMethod.Put,
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, converter, cancellationToken);

        protected Task<TResult> PatchAsync<TResult>(string path, BaseOptions? options, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = new HttpMethod("PATCH"),
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, cancellationToken);

        protected Task<TResult> PatchAsync<TResult>(string path, BaseOptions? options, Newtonsoft.Json.JsonConverter converter, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = new HttpMethod("PATCH"),
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, converter, cancellationToken);

        protected Task DeleteAsync(string path, BaseOptions? options, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeRawAPIRequest(new RetabRequest
            {
                Method = HttpMethod.Delete,
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, cancellationToken);

        protected Task<TResult> DeleteAsync<TResult>(string path, BaseOptions? options, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = HttpMethod.Delete,
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, cancellationToken);

        // ── Pagination ─────────────────────────────────────────────────

        protected async Task<PaginatedList<T>> FetchPageAsync<T>(
            string path,
            ListOptions? options,
            string? httpBearer,
            RequestOptions? requestOptions,
            CancellationToken cancellationToken)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = path,
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };

            var page = await this.Client.MakeAPIRequest<PaginatedList<T>>(request, cancellationToken).ConfigureAwait(false)
                ?? new PaginatedList<T>();

            page.FetchNextPage = (after, ct) =>
            {
                var nextOptions = CloneOptionsForNextPage(options, after);
                return this.FetchPageAsync<T>(path, nextOptions, httpBearer, requestOptions, ct);
            };

            return page;
        }

        protected async IAsyncEnumerable<T> ListAutoPagingAsync<T>(
            string path,
            ListOptions? options,
            RequestOptions? requestOptions,
            [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken)
        {
            var page = await this.FetchPageAsync<T>(path, options, null, requestOptions, cancellationToken).ConfigureAwait(false);
            await foreach (var item in page.AutoPagingIterAsync(cancellationToken).ConfigureAwait(false))
            {
                yield return item;
            }
        }

        private static ListOptions CloneOptionsForNextPage(ListOptions? source, string after)
        {
            if (source == null)
            {
                return new RetabListOptions { After = after };
            }

            var json = System.Text.Json.JsonSerializer.Serialize(source, source.GetType(), Retab.JsonOptions);
            var clone = (ListOptions?)System.Text.Json.JsonSerializer.Deserialize(json, source.GetType(), Retab.JsonOptions)
                ?? new RetabListOptions();
            clone.After = after;
            clone.Before = null;
            return clone;
        }

        /// <summary>Concrete <see cref="ListOptions"/> the pagination helper instantiates when callers pass null.</summary>
        private sealed class RetabListOptions : ListOptions { }
    }
}
