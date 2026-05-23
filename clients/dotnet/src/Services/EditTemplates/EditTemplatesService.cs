namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

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
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="EditTemplate"/> results.</returns>
        public virtual async Task<PaginatedList<EditTemplate>> ListAsync(string httpBearer, EditTemplatesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<EditTemplate>("/v1/edits/templates", options, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<EditTemplate>> List(string httpBearer, EditTemplatesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(httpBearer, options, requestOptions, cancellationToken);
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
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="EditTemplate"/> result.</returns>
        public virtual async Task<EditTemplate> CreateAsync(string httpBearer, EditTemplatesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = "/v1/edits/templates",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<EditTemplate>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<EditTemplate> Create(string httpBearer, EditTemplatesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Get Template</summary>
        /// <param name="templateId">The template id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="EditTemplate"/> result.</returns>
        public virtual async Task<EditTemplate> GetAsync(string templateId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = $"/v1/edits/templates/{Uri.EscapeDataString(templateId)}",
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<EditTemplate>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<EditTemplate> Get(string templateId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(templateId, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Update Template</summary>
        /// <param name="templateId">The template id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="EditTemplate"/> result.</returns>
        public virtual async Task<EditTemplate> UpdateAsync(string templateId, string httpBearer, EditTemplatesUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Patch,
                Path = $"/v1/edits/templates/{Uri.EscapeDataString(templateId)}",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<EditTemplate>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="UpdateAsync"/>.</summary>
        public virtual Task<EditTemplate> Update(string templateId, string httpBearer, EditTemplatesUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.UpdateAsync(templateId, httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Template</summary>
        /// <param name="templateId">The template id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string templateId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Delete,
                Path = $"/v1/edits/templates/{Uri.EscapeDataString(templateId)}",
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            await this.Client.MakeRawAPIRequest(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string templateId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(templateId, httpBearer, requestOptions, cancellationToken);
        }
    }
}
