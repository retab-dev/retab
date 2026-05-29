namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

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
        /// <remarks>
        /// List parses.
        /// Returns a paginated list of parses for the authenticated environment, newest first by
        /// default. Filter by `filename` prefix (case-insensitive) and by a `created_at` window
        /// using `from_date`/`to_date` (`YYYY-MM-DD`). Page through results with `before`/`after`,
        /// `limit`, and `order`.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="Parse"/> results.</returns>
        public virtual async Task<PaginatedList<Parse>> ListAsync(ParsesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<Parse>("/v1/parses", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<Parse>> List(ParsesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
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
        /// <remarks>
        /// Create a parse.
        /// Extracts the full text of a `document` into per-page and concatenated text using
        /// the chosen `model`. Tables are rendered in the requested `table_parsing_format`, and
        /// optional `instructions` steer the parse. Returns the stored `Parse` with its `output`
        /// and `usage`, and responds with `201`.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Parse"/> result.</returns>
        public virtual async Task<Parse> CreateAsync(ParsesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Parse>("/v1/parses", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<Parse> Create(ParsesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Parse</summary>
        /// <remarks>
        /// Retrieve a parse.
        /// Fetches a single parse by its `parse_id` within the authenticated environment and
        /// returns the full `Parse` including its `output`. Responds with `404` if no parse with
        /// that id exists.
        /// </remarks>
        /// <param name="parseId">The parse id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Parse"/> result.</returns>
        public virtual async Task<Parse> GetAsync(string parseId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<Parse>($"/v1/parses/{Uri.EscapeDataString(parseId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Parse> Get(string parseId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(parseId, requestOptions, cancellationToken);
        }

        /// <summary>Delete Parse</summary>
        /// <remarks>
        /// Delete a parse.
        /// Permanently deletes the parse identified by `parse_id` in the authenticated
        /// environment. Returns `204` with no body on success, or `404` if the parse does not
        /// exist.
        /// </remarks>
        /// <param name="parseId">The parse id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string parseId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/parses/{Uri.EscapeDataString(parseId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string parseId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(parseId, requestOptions, cancellationToken);
        }
    }
}
