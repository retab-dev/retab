namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the workflow block executions API operations on <see cref="Retab"/>.</summary>
    public class WorkflowBlockExecutionsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowBlockExecutionsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowBlockExecutionsService(Retab client) : base(client) { }

        /// <summary>List Block Executions</summary>
        /// <remarks>
        /// List recent block executions for one workflow run block.
        /// Cursor pagination matches the conventions used by
        /// ``GET /v1/extractions`` — pass ``after`` from the previous page's
        /// ``list_metadata.after`` to advance, ``before`` to step backwards, and
        /// ``order`` to flip the sort direction. ``run_id`` + ``block_id`` are
        /// required scope filters; without them this endpoint would expose
        /// cross-run cursors that walk arbitrary block executions.
        /// </remarks>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="StoredBlockExecution"/> results.</returns>
        public virtual async Task<PaginatedList<StoredBlockExecution>> ListAsync(string httpBearer, WorkflowBlockExecutionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<StoredBlockExecution>("/v1/workflows/blocks/executions", options, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<StoredBlockExecution>> List(string httpBearer, WorkflowBlockExecutionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="StoredBlockExecution"/> items.</returns>
        public virtual IAsyncEnumerable<StoredBlockExecution> ListAutoPagingAsync(WorkflowBlockExecutionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<StoredBlockExecution>("/v1/workflows/blocks/executions", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Block Execution</summary>
        /// <remarks>
        /// Create a block execution for ``block_id`` against the current draft.
        /// </remarks>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="StoredBlockExecution"/> result.</returns>
        public virtual async Task<StoredBlockExecution> CreateAsync(string httpBearer, WorkflowBlockExecutionsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = "/v1/workflows/blocks/executions",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<StoredBlockExecution>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<StoredBlockExecution> Create(string httpBearer, WorkflowBlockExecutionsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(httpBearer, options, requestOptions, cancellationToken);
        }
    }
}
