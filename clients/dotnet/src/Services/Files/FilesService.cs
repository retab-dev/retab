namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the files API operations on <see cref="Retab"/>.</summary>
    public class FilesService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="FilesService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public FilesService(Retab client) : base(client) { }

        /// <summary>List Files</summary>
        /// <remarks>
        /// List files with pagination and optional filtering.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="File"/> results.</returns>
        public virtual async Task<PaginatedList<File>> ListAsync(FilesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<File>("/v1/files", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<File>> List(FilesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="File"/> items.</returns>
        public virtual IAsyncEnumerable<File> ListAutoPagingAsync(FilesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<File>("/v1/files", options, requestOptions, cancellationToken);
        }

        /// <summary>Upload File</summary>
        /// <remarks>
        /// Start a file upload.
        /// Reserves a file record for the given `filename`, `content_type`, and
        /// `size_bytes`, and returns a short-lived signed `upload_url` the client uses
        /// to `PUT` the file content directly. Call the complete-upload endpoint with
        /// the returned `file_id` once the bytes have been uploaded.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="CreateUploadResponse"/> result.</returns>
        public virtual async Task<CreateUploadResponse> CreateUploadAsync(FilesCreateUploadOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<CreateUploadResponse>("/v1/files/upload", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateUploadAsync"/>.</summary>
        public virtual Task<CreateUploadResponse> CreateUpload(FilesCreateUploadOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateUploadAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Complete Upload File</summary>
        /// <remarks>
        /// Finalize a file upload.
        /// Confirms that the content for `file_id` has been uploaded, verifying the
        /// object's size and optional `sha256` checksum against the upload session,
        /// and marks the file ready. Returns a durable reference to the stored file.
        /// Responds with `404` if the upload session is unknown, `410` if it has
        /// expired, and `422` if the size or checksum does not match.
        /// </remarks>
        /// <param name="fileId">The file id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="MimeData"/> result.</returns>
        public virtual async Task<MimeData> CompleteUploadAsync(string fileId, FilesCompleteUploadOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<MimeData>($"/v1/files/upload/{Uri.EscapeDataString(fileId)}/complete", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CompleteUploadAsync"/>.</summary>
        public virtual Task<MimeData> CompleteUpload(string fileId, FilesCompleteUploadOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CompleteUploadAsync(fileId, options, requestOptions, cancellationToken);
        }

        /// <summary>Get File</summary>
        /// <remarks>
        /// Retrieve a file.
        /// Returns metadata for the file identified by `file_id`, including its
        /// `filename`, `page_count`, and timestamps. Responds with `404` if no
        /// matching file exists.
        /// </remarks>
        /// <param name="fileId">The file id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="File"/> result.</returns>
        public virtual async Task<File> GetAsync(string fileId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<File>($"/v1/files/{Uri.EscapeDataString(fileId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<File> Get(string fileId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(fileId, requestOptions, cancellationToken);
        }

        /// <summary>Download Link</summary>
        /// <remarks>
        /// Get a temporary download link for a file.
        /// Returns a short-lived signed `download_url` for the file identified by
        /// `file_id`, along with its `filename` and expiration. Responds with `404`
        /// if no matching file exists.
        /// </remarks>
        /// <param name="fileId">The file id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="FileLink"/> result.</returns>
        public virtual async Task<FileLink> GetDownloadLinkAsync(string fileId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<FileLink>($"/v1/files/{Uri.EscapeDataString(fileId)}/download-link", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetDownloadLinkAsync"/>.</summary>
        public virtual Task<FileLink> GetDownloadLink(string fileId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetDownloadLinkAsync(fileId, requestOptions, cancellationToken);
        }
    }
}
