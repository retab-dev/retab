namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the extractions API operations on <see cref="Retab"/>.</summary>
    public class ExtractionsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="ExtractionsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public ExtractionsService(Retab client) : base(client) { }

        /// <summary>List Extractions</summary>
        /// <remarks>
        /// List and paginate extractions with optional filtering.
        /// Returns a paginated list of extraction documents matching the filter criteria.
        /// The `metadata` parameter accepts a JSON string of key-value pairs to filter by.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="Extraction"/> results.</returns>
        public virtual async Task<PaginatedList<Extraction>> ListAsync(ExtractionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<Extraction>("/v1/extractions", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<Extraction>> List(ExtractionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="Extraction"/> items.</returns>
        public virtual IAsyncEnumerable<Extraction> ListAutoPagingAsync(ExtractionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<Extraction>("/v1/extractions", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Extraction</summary>
        /// <remarks>
        /// Run a structured extraction on a document.
        /// Extracts structured data from the `document` according to the supplied
        /// `json_schema`, using the requested `model`. Returns the extraction
        /// with its `output`, consensus details, and usage on `201`. When
        /// `stream` is `true`, partial results are streamed back as they are produced.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Extraction"/> result.</returns>
        public virtual async Task<Extraction> CreateAsync(ExtractionsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Extraction>("/v1/extractions", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<Extraction> Create(ExtractionsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Create Extraction Stream</summary>
        /// <remarks>
        /// Run a structured extraction on a document and stream partial results as they are produced.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task CreateStreamAsync(ExtractionsCreateStreamOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.PostAsync<object>("/v1/extractions/stream", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateStreamAsync"/>.</summary>
        public virtual Task CreateStream(ExtractionsCreateStreamOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateStreamAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Extraction</summary>
        /// <remarks>
        /// Retrieve an extraction.
        /// Returns the extraction identified by `extraction_id`, including its source
        /// file, schema, `output`, and consensus details. Responds with `404` if no
        /// matching extraction exists.
        /// </remarks>
        /// <param name="extractionId">The extraction id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Extraction"/> result.</returns>
        public virtual async Task<Extraction> GetAsync(string extractionId, ExtractionsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<Extraction>($"/v1/extractions/{Uri.EscapeDataString(extractionId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Extraction> Get(string extractionId, ExtractionsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(extractionId, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Extraction</summary>
        /// <remarks>
        /// Delete an extraction
        /// </remarks>
        /// <param name="extractionId">The extraction id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string extractionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/extractions/{Uri.EscapeDataString(extractionId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string extractionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(extractionId, requestOptions, cancellationToken);
        }

        /// <summary>Cancel Extraction</summary>
        /// <param name="extractionId">The extraction id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Extraction"/> result.</returns>
        public virtual async Task<Extraction> CreateCancelAsync(string extractionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Extraction>($"/v1/extractions/{Uri.EscapeDataString(extractionId)}/cancel", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateCancelAsync"/>.</summary>
        public virtual Task<Extraction> CreateCancel(string extractionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateCancelAsync(extractionId, requestOptions, cancellationToken);
        }

        /// <summary>Get Extraction Sources</summary>
        /// <remarks>
        /// Return the extraction result enriched with per-leaf source provenance.
        /// Each extracted leaf value is wrapped as {value, source} where source
        /// contains citation content, surrounding context, and a format-specific
        /// anchor (bbox for PDFs, cell ref for spreadsheets, text span for plain text, etc.).
        /// </remarks>
        /// <param name="extractionId">The extraction id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="SourcesResponse"/> result.</returns>
        public virtual async Task<SourcesResponse> SourcesAsync(string extractionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<SourcesResponse>($"/v1/extractions/{Uri.EscapeDataString(extractionId)}/sources", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="SourcesAsync"/>.</summary>
        public virtual Task<SourcesResponse> Sources(string extractionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.SourcesAsync(extractionId, requestOptions, cancellationToken);
        }
    }
}
