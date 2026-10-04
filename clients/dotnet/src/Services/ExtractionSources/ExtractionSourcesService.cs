namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the extraction sources API operations on <see cref="Retab"/>.</summary>
    public class ExtractionSourcesService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="ExtractionSourcesService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public ExtractionSourcesService(Retab client) : base(client) { }

        /// <summary>Get Extraction Sources</summary>
        /// <remarks>
        /// Return the extraction result enriched with per-leaf source provenance.
        /// Each extracted leaf value is wrapped as {value, source} where source
        /// contains citation content, surrounding context, and a format-specific
        /// anchor (bbox for PDFs, cell ref for spreadsheets, text span for plain text, etc.).
        /// </remarks>
        /// <param name="extractionId">The extraction id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="SourcesResponse"/> result.</returns>
        public virtual async Task<SourcesResponse> GetAsync(string extractionId, ExtractionSourcesGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<SourcesResponse>($"/v1/extractions/{Uri.EscapeDataString(extractionId)}/sources", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<SourcesResponse> Get(string extractionId, ExtractionSourcesGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(extractionId, options, requestOptions, cancellationToken);
        }

        /// <summary>Create Extraction Sources</summary>
        /// <remarks>
        /// Create or join a durable sources computation. Located finds printed answers; cited also seeks supporting inputs. Background requests return immediately; synchronous requests wait up to 20 seconds before returning 202 with a version-pinned Location. Repeated requests reuse work and never bill twice. Processing completion does not guarantee evidence for every field. A cited request during a located-only computation returns 409; retry after that computation finishes.
        /// </remarks>
        /// <param name="extractionId">The extraction id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="SourcesResponse"/> result.</returns>
        public virtual async Task<SourcesResponse> CreateAsync(string extractionId, ExtractionSourcesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<SourcesResponse>($"/v1/extractions/{Uri.EscapeDataString(extractionId)}/sources", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<SourcesResponse> Create(string extractionId, ExtractionSourcesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(extractionId, options, requestOptions, cancellationToken);
        }
    }
}
