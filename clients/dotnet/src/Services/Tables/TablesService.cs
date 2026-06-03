namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the tables API operations on <see cref="Retab"/>.</summary>
    public class TablesService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="TablesService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public TablesService(Retab client) : base(client) { }

        /// <summary>List Tables</summary>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTableListResponse"/> result.</returns>
        public virtual async Task<WorkflowTableListResponse> ListAsync(RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowTableListResponse>("/v1/tables", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<WorkflowTableListResponse> List(RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(requestOptions, cancellationToken);
        }

        /// <summary>Create Table</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTableListResponse"/> result.</returns>
        public virtual async Task<WorkflowTableListResponse> CreateAsync(TablesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowTableListResponse>("/v1/tables", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<WorkflowTableListResponse> Create(TablesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Table</summary>
        /// <param name="tableId">The table id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTableResponse"/> result.</returns>
        public virtual async Task<WorkflowTableResponse> GetAsync(string tableId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowTableResponse>($"/v1/tables/{Uri.EscapeDataString(tableId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowTableResponse> Get(string tableId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(tableId, requestOptions, cancellationToken);
        }

        /// <summary>Replace Table</summary>
        /// <param name="tableId">The table id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTableListResponse"/> result.</returns>
        public virtual async Task<WorkflowTableListResponse> ReplaceAsync(string tableId, TablesReplaceOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PutAsync<WorkflowTableListResponse>($"/v1/tables/{Uri.EscapeDataString(tableId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ReplaceAsync"/>.</summary>
        public virtual Task<WorkflowTableListResponse> Replace(string tableId, TablesReplaceOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ReplaceAsync(tableId, options, requestOptions, cancellationToken);
        }

        /// <summary>Update Table</summary>
        /// <param name="tableId">The table id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTableListResponse"/> result.</returns>
        public virtual async Task<WorkflowTableListResponse> UpdateAsync(string tableId, TablesUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PatchAsync<WorkflowTableListResponse>($"/v1/tables/{Uri.EscapeDataString(tableId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="UpdateAsync"/>.</summary>
        public virtual Task<WorkflowTableListResponse> Update(string tableId, TablesUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.UpdateAsync(tableId, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Table</summary>
        /// <param name="tableId">The table id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string tableId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/tables/{Uri.EscapeDataString(tableId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string tableId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(tableId, requestOptions, cancellationToken);
        }

        /// <summary>Download Table Csv</summary>
        /// <param name="tableId">The table id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DownloadAsync(string tableId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.GetAsync<object>($"/v1/tables/{Uri.EscapeDataString(tableId)}/download", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DownloadAsync"/>.</summary>
        public virtual Task Download(string tableId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DownloadAsync(tableId, requestOptions, cancellationToken);
        }

        /// <summary>Profile Table</summary>
        /// <param name="tableId">The table id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTableProfileResponse"/> result.</returns>
        public virtual async Task<WorkflowTableProfileResponse> ProfileAsync(string tableId, TablesProfileOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowTableProfileResponse>($"/v1/tables/{Uri.EscapeDataString(tableId)}/profile", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ProfileAsync"/>.</summary>
        public virtual Task<WorkflowTableProfileResponse> Profile(string tableId, TablesProfileOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ProfileAsync(tableId, options, requestOptions, cancellationToken);
        }

        /// <summary>Query Table</summary>
        /// <param name="tableId">The table id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTableRowsResponse"/> result.</returns>
        public virtual async Task<WorkflowTableRowsResponse> QueryAsync(string tableId, TablesQueryOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowTableRowsResponse>($"/v1/tables/{Uri.EscapeDataString(tableId)}/query", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="QueryAsync"/>.</summary>
        public virtual Task<WorkflowTableRowsResponse> Query(string tableId, TablesQueryOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.QueryAsync(tableId, options, requestOptions, cancellationToken);
        }

        /// <summary>Get Table Schema</summary>
        /// <param name="tableId">The table id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTableSchemaResponse"/> result.</returns>
        public virtual async Task<WorkflowTableSchemaResponse> SchemaAsync(string tableId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowTableSchemaResponse>($"/v1/tables/{Uri.EscapeDataString(tableId)}/schema", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="SchemaAsync"/>.</summary>
        public virtual Task<WorkflowTableSchemaResponse> Schema(string tableId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.SchemaAsync(tableId, requestOptions, cancellationToken);
        }

        /// <summary>Validate Table</summary>
        /// <param name="tableId">The table id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTableValidationResponse"/> result.</returns>
        public virtual async Task<WorkflowTableValidationResponse> ValidateAsync(string tableId, TablesValidateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowTableValidationResponse>($"/v1/tables/{Uri.EscapeDataString(tableId)}/validate", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ValidateAsync"/>.</summary>
        public virtual Task<WorkflowTableValidationResponse> Validate(string tableId, TablesValidateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ValidateAsync(tableId, options, requestOptions, cancellationToken);
        }
    }
}
