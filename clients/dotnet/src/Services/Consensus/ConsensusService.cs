namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the consensus API operations on <see cref="Retab"/>.</summary>
    public class ConsensusService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="ConsensusService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public ConsensusService(Retab client) : base(client) { }

        /// <summary>Create Consensus</summary>
        /// <remarks>
        /// Reconcile several objects into one.
        /// Takes a list of `inputs` that describe the same thing — typically the outputs of
        /// repeated extractions — and returns a single `consensus` object plus a `likelihoods`
        /// map giving the agreement score for every leaf path.
        /// Inputs are always aligned before the vote, so items in a list are matched across
        /// inputs by content rather than by position; a reordered or partially-omitted array
        /// still reconciles correctly. Supply `json_schema` when you have one: it makes numeric
        /// fields reconcile by declared type — `integer` votes on the exact value, `number`
        /// clusters and averages — instead of inferring the kind from the values present. Set
        /// `include_alignment` to also receive the canonical path mapping.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ReconciliationResult"/> result.</returns>
        public virtual async Task<ReconciliationResult> CreateAsync(ConsensusCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<ReconciliationResult>("/v1/consensus", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<ReconciliationResult> Create(ConsensusCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }
    }
}
