namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the schemas API operations on <see cref="Retab"/>.</summary>
    public class SchemasService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="SchemasService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public SchemasService(Retab client) : base(client) { }

        /// <summary>Generate Schema From Examples</summary>
        /// <remarks>
        /// Generates a JSON Schema from scratch by inferring structure from the content of the provided example documents.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="MainServerServicesV1SchemasModelsSchemaGeneration"/> result.</returns>
        public virtual async Task<MainServerServicesV1SchemasModelsSchemaGeneration> GenerateAsync(SchemasGenerateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<MainServerServicesV1SchemasModelsSchemaGeneration>("/v1/schemas/generate", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GenerateAsync"/>.</summary>
        public virtual Task<MainServerServicesV1SchemasModelsSchemaGeneration> Generate(SchemasGenerateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GenerateAsync(options, requestOptions, cancellationToken);
        }
    }
}
