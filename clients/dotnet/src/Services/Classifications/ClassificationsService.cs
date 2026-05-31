namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the classifications API operations on <see cref="Retab"/>.</summary>
    public class ClassificationsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="ClassificationsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public ClassificationsService(Retab client) : base(client) { }

        /// <summary>List Classifications</summary>
        /// <remarks>
        /// List classifications.
        /// Returns a paginated list of classifications, most recent first. Filter by
        /// `filename` or a `from_date`/`to_date` range, and page with `before`/`after`.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="Classification"/> results.</returns>
        public virtual async Task<PaginatedList<Classification>> ListAsync(ClassificationsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<Classification>("/v1/classifications", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<Classification>> List(ClassificationsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="Classification"/> items.</returns>
        public virtual IAsyncEnumerable<Classification> ListAutoPagingAsync(ClassificationsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<Classification>("/v1/classifications", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Classification</summary>
        /// <remarks>
        /// Classify a document.
        /// Runs a classification on the supplied `document` against the provided
        /// `categories`. Tune the run with `model`, `instructions`, `first_n_pages`
        /// (limit to the first pages), and `n_consensus` (number of votes to combine).
        /// Returns the created classification with the chosen category and reasoning;
        /// responds with `201`.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Classification"/> result.</returns>
        public virtual async Task<Classification> CreateAsync(ClassificationsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Classification>("/v1/classifications", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<Classification> Create(ClassificationsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Classification</summary>
        /// <remarks>
        /// Retrieve a classification.
        /// Fetches a single classification by its `classification_id`. Returns the
        /// classification with its file reference, categories, and result; responds
        /// with `404` if no classification with that id exists.
        /// </remarks>
        /// <param name="classificationId">The classification id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Classification"/> result.</returns>
        public virtual async Task<Classification> GetAsync(string classificationId, ClassificationsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<Classification>($"/v1/classifications/{Uri.EscapeDataString(classificationId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Classification> Get(string classificationId, ClassificationsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(classificationId, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Classification</summary>
        /// <remarks>
        /// Delete a classification.
        /// Permanently deletes the classification identified by `classification_id`.
        /// Responds with 204 on success, or 404 if no such classification exists.
        /// </remarks>
        /// <param name="classificationId">The classification id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string classificationId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/classifications/{Uri.EscapeDataString(classificationId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string classificationId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(classificationId, requestOptions, cancellationToken);
        }

        /// <summary>Cancel Classification</summary>
        /// <param name="classificationId">The classification id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Classification"/> result.</returns>
        public virtual async Task<Classification> CreateCancelAsync(string classificationId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Classification>($"/v1/classifications/{Uri.EscapeDataString(classificationId)}/cancel", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateCancelAsync"/>.</summary>
        public virtual Task<Classification> CreateCancel(string classificationId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateCancelAsync(classificationId, requestOptions, cancellationToken);
        }
    }
}
