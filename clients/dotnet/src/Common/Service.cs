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

        protected Task<TResult> PostAsync<TResult>(string path, BaseOptions? options, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, cancellationToken);

        protected Task<TResult> PutAsync<TResult>(string path, BaseOptions? options, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = HttpMethod.Put,
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, cancellationToken);

        protected Task<TResult> PatchAsync<TResult>(string path, BaseOptions? options, RequestOptions? requestOptions, CancellationToken cancellationToken)
            => this.Client.MakeAPIRequest<TResult>(new RetabRequest
            {
                Method = new HttpMethod("PATCH"),
                Path = path,
                Options = options,
                RequestOptions = requestOptions,
            }, cancellationToken);

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

        // ── Auto-paging ─────────────────────────────────────────────────

        protected async IAsyncEnumerable<T> ListAutoPagingAsync<T>(
            string path,
            ListOptions? options,
            RequestOptions? requestOptions,
            [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken)
        {
            var current = await this.GetAsync<RetabList<T>>(path, options, requestOptions, cancellationToken).ConfigureAwait(false);
            while (current != null)
            {
                if (current.Data != null)
                {
                    foreach (var item in current.Data) yield return item;
                }
                var after = current.ListMetadata?.After;
                if (string.IsNullOrEmpty(after)) yield break;
                var nextOptions = options ?? new RetabListOptions();
                nextOptions.After = after;
                current = await this.GetAsync<RetabList<T>>(path, nextOptions, requestOptions, cancellationToken).ConfigureAwait(false);
            }
        }

        /// <summary>Concrete <see cref="ListOptions"/> the auto-paging helper instantiates when callers pass null.</summary>
        private sealed class RetabListOptions : ListOptions { }
    }
}
