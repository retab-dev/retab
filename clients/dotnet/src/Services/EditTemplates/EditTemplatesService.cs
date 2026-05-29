namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the edit templates API operations on <see cref="Retab"/>.</summary>
    public class EditTemplatesService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="EditTemplatesService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public EditTemplatesService(Retab client) : base(client) { }

        /// <summary>List Templates</summary>
        /// <remarks>
        /// List edit templates.
        /// Returns a paginated list of edit templates. Filter by `name`
        /// (case-insensitive substring match) and order results by `sort_by`
        /// (`created_at` or `name`). Page with `before`/`after` cursors, `limit`, and
        /// `order`. Form fields are omitted from list items; fetch a single template to
        /// retrieve them.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="EditTemplate"/> results.</returns>
        public virtual async Task<PaginatedList<EditTemplate>> ListAsync(EditTemplatesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<EditTemplate>("/v1/edits/templates", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<EditTemplate>> List(EditTemplatesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="EditTemplate"/> items.</returns>
        public virtual IAsyncEnumerable<EditTemplate> ListAutoPagingAsync(EditTemplatesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<EditTemplate>("/v1/edits/templates", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Template</summary>
        /// <remarks>
        /// Create an edit template.
        /// Stores a reusable form template from an empty `document` (PDF or Office
        /// document) plus its `form_fields` and a `name`. Later edits can reference the
        /// returned template id instead of re-uploading the document. An unsupported
        /// document format responds with `400`; on success responds with `201`.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="EditTemplate"/> result.</returns>
        public virtual async Task<EditTemplate> CreateAsync(EditTemplatesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<EditTemplate>("/v1/edits/templates", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<EditTemplate> Create(EditTemplatesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Template</summary>
        /// <remarks>
        /// Retrieve an edit template.
        /// Fetches a single edit template by its `template_id`, including its form
        /// fields. Responds with `404` if no template with that id exists.
        /// </remarks>
        /// <param name="templateId">The template id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="EditTemplate"/> result.</returns>
        public virtual async Task<EditTemplate> GetAsync(string templateId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<EditTemplate>($"/v1/edits/templates/{Uri.EscapeDataString(templateId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<EditTemplate> Get(string templateId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(templateId, requestOptions, cancellationToken);
        }

        /// <summary>Update Template</summary>
        /// <remarks>
        /// Update an edit template.
        /// Applies a partial update to the template identified by `template_id`. Set
        /// `name` to rename it and/or `form_fields` to replace its field list; omitted
        /// fields are left unchanged. Returns the updated template, or `404` if no
        /// template with that id exists.
        /// </remarks>
        /// <param name="templateId">The template id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="EditTemplate"/> result.</returns>
        public virtual async Task<EditTemplate> UpdateAsync(string templateId, EditTemplatesUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PatchAsync<EditTemplate>($"/v1/edits/templates/{Uri.EscapeDataString(templateId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="UpdateAsync"/>.</summary>
        public virtual Task<EditTemplate> Update(string templateId, EditTemplatesUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.UpdateAsync(templateId, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Template</summary>
        /// <remarks>
        /// Delete an edit template.
        /// Permanently deletes the edit template identified by `template_id`. Returns
        /// `204` on success, or `404` if no template with that id exists.
        /// </remarks>
        /// <param name="templateId">The template id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string templateId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/edits/templates/{Uri.EscapeDataString(templateId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string templateId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(templateId, requestOptions, cancellationToken);
        }
    }
}
