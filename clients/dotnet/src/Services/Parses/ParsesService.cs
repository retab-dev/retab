namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the parses API operations on <see cref="Retab"/>.</summary>
    public class ParsesService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="ParsesService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public ParsesService(Retab client) : base(client) { }

        /// <summary>List Parses</summary>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="Parse"/> results.</returns>
        public virtual async Task<PaginatedList<Parse>> ListAsync(string httpBearer, ParsesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<Parse>("/v1/parses", options, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<Parse>> List(string httpBearer, ParsesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="Parse"/> items.</returns>
        public virtual IAsyncEnumerable<Parse> ListAutoPagingAsync(ParsesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<Parse>("/v1/parses", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Parse</summary>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Parse"/> result.</returns>
        public virtual async Task<Parse> CreateAsync(string httpBearer, ParsesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = "/v1/parses",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<Parse>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<Parse> Create(string httpBearer, ParsesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Get Parse</summary>
        /// <param name="parseId">The parse id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Parse"/> result.</returns>
        public virtual async Task<Parse> GetAsync(string parseId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = $"/v1/parses/{Uri.EscapeDataString(parseId)}",
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<Parse>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Parse> Get(string parseId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(parseId, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Delete Parse</summary>
        /// <param name="parseId">The parse id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string parseId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Delete,
                Path = $"/v1/parses/{Uri.EscapeDataString(parseId)}",
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            await this.Client.MakeRawAPIRequest(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string parseId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(parseId, httpBearer, requestOptions, cancellationToken);
        }
    }
}
