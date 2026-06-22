namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the edits API operations on <see cref="Retab"/>.</summary>
    public class EditsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="EditsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public EditsService(Retab client) : base(client) { }

        /// <summary>Gets the nested <see cref="EditTemplatesService"/> service.</summary>
        public virtual EditTemplatesService Templates => new EditTemplatesService(this.Client);

        /// <summary>List Edits</summary>
        /// <remarks>
        /// List edits.
        /// Returns a paginated list of edits. Filter by `filename` (case-insensitive
        /// prefix match), `template_id`, and a `from_date`/`to_date` creation range
        /// (each `YYYY-MM-DD`). Page with `before`/`after` cursors, `limit`, and
        /// `order`; an invalid date format responds with `400`.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="Edit"/> results.</returns>
        public virtual async Task<PaginatedList<Edit>> ListAsync(EditsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<Edit>("/v1/edits", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<Edit>> List(EditsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="Edit"/> items.</returns>
        public virtual IAsyncEnumerable<Edit> ListAutoPagingAsync(EditsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<Edit>("/v1/edits", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Edit</summary>
        /// <remarks>
        /// Create an edit.
        /// Fills the form fields of a document according to `instructions` and renders
        /// the values into a PDF. Provide exactly one of `document` (a PDF, DOCX, XLSX,
        /// or PPTX) or `template_id` (an existing edit template) — supplying both or
        /// neither responds with `400`. Returns the created edit with the filled form
        /// data and rendered document; responds with `201`.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Edit"/> result.</returns>
        public virtual async Task<Edit> CreateAsync(EditsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Edit>("/v1/edits", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<Edit> Create(EditsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Edit</summary>
        /// <remarks>
        /// Retrieve an edit.
        /// Fetches a single edit by its `edit_id`. Returns the edit with its filled
        /// form data and rendered document; responds with `404` if no edit with that
        /// id exists.
        /// </remarks>
        /// <param name="editId">The edit id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Edit"/> result.</returns>
        public virtual async Task<Edit> GetAsync(string editId, EditsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<Edit>($"/v1/edits/{Uri.EscapeDataString(editId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Edit> Get(string editId, EditsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(editId, options, requestOptions, cancellationToken);
        }

        /// <summary>Cancel Edit</summary>
        /// <param name="editId">The edit id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Edit"/> result.</returns>
        public virtual async Task<Edit> CreateCancelAsync(string editId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Edit>($"/v1/edits/{Uri.EscapeDataString(editId)}/cancel", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateCancelAsync"/>.</summary>
        public virtual Task<Edit> CreateCancel(string editId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateCancelAsync(editId, requestOptions, cancellationToken);
        }
    }
}
