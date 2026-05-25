namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the auth API operations on <see cref="Retab"/>.</summary>
    public class AuthService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="AuthService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public AuthService(Retab client) : base(client) { }

        /// <summary>Get Auth Status</summary>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="AuthStatusResponse"/> result.</returns>
        public virtual async Task<AuthStatusResponse> ListStatusAsync(RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<AuthStatusResponse>("/v1/auth/status", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListStatusAsync"/>.</summary>
        public virtual Task<AuthStatusResponse> ListStatus(RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListStatusAsync(requestOptions, cancellationToken);
        }
    }
}
