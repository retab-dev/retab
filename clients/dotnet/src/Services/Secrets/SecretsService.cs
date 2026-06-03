namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the secrets API operations on <see cref="Retab"/>.</summary>
    public class SecretsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="SecretsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public SecretsService(Retab client) : base(client) { }

        /// <summary>List Secrets</summary>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="SecretListResponse"/> result.</returns>
        public virtual async Task<SecretListResponse> ListAsync(RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<SecretListResponse>("/v1/secrets", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<SecretListResponse> List(RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(requestOptions, cancellationToken);
        }

        /// <summary>Create Secret</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="SecretResponse"/> result.</returns>
        public virtual async Task<SecretResponse> CreateAsync(SecretsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<SecretResponse>("/v1/secrets", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<SecretResponse> Create(SecretsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Secret</summary>
        /// <param name="name">The name.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="SecretResponse"/> result.</returns>
        public virtual async Task<SecretResponse> GetAsync(string name, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<SecretResponse>($"/v1/secrets/{Uri.EscapeDataString(name)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<SecretResponse> Get(string name, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(name, requestOptions, cancellationToken);
        }

        /// <summary>Set Secret</summary>
        /// <param name="name">The name.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="SecretResponse"/> result.</returns>
        public virtual async Task<SecretResponse> UpdateAsync(string name, SecretsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PutAsync<SecretResponse>($"/v1/secrets/{Uri.EscapeDataString(name)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="UpdateAsync"/>.</summary>
        public virtual Task<SecretResponse> Update(string name, SecretsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.UpdateAsync(name, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Secret</summary>
        /// <param name="name">The name.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string name, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/secrets/{Uri.EscapeDataString(name)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string name, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(name, requestOptions, cancellationToken);
        }
    }
}
