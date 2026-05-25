namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the environments API operations on <see cref="Retab"/>.</summary>
    public class EnvironmentsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="EnvironmentsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public EnvironmentsService(Retab client) : base(client) { }

        /// <summary>List Organization Environments</summary>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="Environment"/> results.</returns>
        public virtual async Task<PaginatedList<Environment>> ListAsync(RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<Environment>("/v1/environments", null, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<Environment>> List(RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(requestOptions, cancellationToken);
        }

        /// <summary>Create Organization Environment</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Environment"/> result.</returns>
        public virtual async Task<Environment> CreateAsync(EnvironmentsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Environment>("/v1/environments", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<Environment> Create(EnvironmentsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Organization Environment</summary>
        /// <param name="environmentId">The environment id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Environment"/> result.</returns>
        public virtual async Task<Environment> GetAsync(string environmentId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<Environment>($"/v1/environments/{Uri.EscapeDataString(environmentId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Environment> Get(string environmentId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(environmentId, requestOptions, cancellationToken);
        }

        /// <summary>Update Organization Environment</summary>
        /// <param name="environmentId">The environment id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Environment"/> result.</returns>
        public virtual async Task<Environment> UpdateAsync(string environmentId, EnvironmentsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PatchAsync<Environment>($"/v1/environments/{Uri.EscapeDataString(environmentId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="UpdateAsync"/>.</summary>
        public virtual Task<Environment> Update(string environmentId, EnvironmentsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.UpdateAsync(environmentId, options, requestOptions, cancellationToken);
        }

        /// <summary>Archive Organization Environment</summary>
        /// <param name="environmentId">The environment id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string environmentId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/environments/{Uri.EscapeDataString(environmentId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string environmentId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(environmentId, requestOptions, cancellationToken);
        }
    }
}
