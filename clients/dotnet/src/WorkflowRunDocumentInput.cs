// @oagen-ignore-file
//
// Hand-maintained. Ergonomic document input for a workflow start-document
// block. The document routes accept URL-backed MIMEData only ({filename, url});
// a stored file id is resolved CLIENT-SIDE into MIMEData via the Files
// download-link endpoint before the wire body is built — mirroring the Node SDK
// (`coerceMimeData`) and the Go CLI (`resolveFileIDToMIMEData`). The old
// FileRef wire body ({id, filename, mime_type}) now 422s and is never sent.

using System;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading;
using System.Threading.Tasks;

namespace Retab
{
    /// <summary>
    /// Ergonomic document input for a workflow start-document block. Accepts a
    /// <see cref="MimeData"/> (URL/data-url passthrough) or a <see cref="FileRef"/>
    /// (a stored file id). A file id is resolved into URL-backed MIMEData via the
    /// Files download-link endpoint when the run is created; the wire always
    /// carries the <c>{filename, url}</c> MIMEData shape.
    /// </summary>
    [JsonConverter(typeof(WorkflowRunDocumentInputJsonConverter))]
    public sealed class WorkflowRunDocumentInput
    {
        internal object Value { get; }

        private WorkflowRunDocumentInput(object value)
        {
            this.Value = value ?? throw new ArgumentNullException(nameof(value));
        }

        public static WorkflowRunDocumentInput FromMimeData(MimeData document)
            => new WorkflowRunDocumentInput(document);

        public static WorkflowRunDocumentInput FromFileRef(FileRef document)
            => new WorkflowRunDocumentInput(document);

        public static implicit operator WorkflowRunDocumentInput(MimeData document)
            => FromMimeData(document);

        public static implicit operator WorkflowRunDocumentInput(FileRef document)
            => FromFileRef(document);

        /// <summary>
        /// Resolve this input into URL-backed <see cref="MimeData"/>. A
        /// <see cref="MimeData"/> passes through; a <see cref="FileRef"/> file id
        /// is resolved against the Files download-link endpoint (mirrors the CLI):
        /// the link's durable <c>mime_data.url</c> wins, falling back to the
        /// signed <c>download_url</c>, and the filename prefers the caller's
        /// FileRef filename, then the link's MIMEData filename, then the link
        /// filename, then <c>"document"</c>.
        /// </summary>
        internal async Task<MimeData> ResolveAsync(Retab client, CancellationToken cancellationToken = default)
        {
            if (this.Value is MimeData mime)
            {
                return mime;
            }

            if (this.Value is FileRef fileRef)
            {
                if (client == null)
                {
                    throw new InvalidOperationException(
                        "WorkflowRunDocumentInput.ResolveAsync: a file-id document requires a Retab client to resolve a download link.");
                }

                var link = await client.Files.GetDownloadLinkAsync(fileRef.Id, null, cancellationToken).ConfigureAwait(false);
                var linkMime = link.MimeData;
                var url = linkMime != null && !string.IsNullOrEmpty(linkMime.Url) ? linkMime.Url : link.DownloadUrl;
                var filename = !string.IsNullOrEmpty(fileRef.Filename)
                    ? fileRef.Filename
                    : (linkMime != null && !string.IsNullOrEmpty(linkMime.Filename)) ? linkMime.Filename
                    : !string.IsNullOrEmpty(link.Filename) ? link.Filename
                    : "document";
                return new MimeData(filename, url);
            }

            throw new InvalidOperationException(
                $"WorkflowRunDocumentInput.ResolveAsync: unsupported value type {this.Value.GetType().Name}.");
        }
    }

    public sealed class WorkflowRunDocumentInputJsonConverter : JsonConverter<WorkflowRunDocumentInput>
    {
        public override WorkflowRunDocumentInput? Read(ref Utf8JsonReader reader, Type typeToConvert, JsonSerializerOptions options)
        {
            using var document = JsonDocument.ParseValue(ref reader);
            var json = document.RootElement.GetRawText();
            // Wire shape is MIMEData ({filename, url}); a legacy {id, ...} body is
            // still parsed back into a FileRef for round-trip ergonomics.
            if (document.RootElement.TryGetProperty("id", out _))
            {
                var fileRef = JsonSerializer.Deserialize<FileRef>(json, options);
                return fileRef == null ? null : WorkflowRunDocumentInput.FromFileRef(fileRef);
            }

            var mimeData = JsonSerializer.Deserialize<MimeData>(json, options);
            return mimeData == null ? null : WorkflowRunDocumentInput.FromMimeData(mimeData);
        }

        public override void Write(Utf8JsonWriter writer, WorkflowRunDocumentInput value, JsonSerializerOptions options)
        {
            // The document routes accept URL-backed MIMEData only. A file-id input
            // must be resolved via WorkflowRunDocumentInput.ResolveAsync before
            // serialization (the resource method does this), so an unresolved
            // FileRef here would produce a 422 body — fail loudly instead.
            if (value.Value is FileRef)
            {
                throw new InvalidOperationException(
                    "WorkflowRunDocumentInput: a file-id document must be resolved into MIMEData before serialization. " +
                    "Use the resource method (e.g. WorkflowRuns.CreateAsync), which resolves file ids via the Files download-link endpoint.");
            }

            JsonSerializer.Serialize(writer, value.Value, value.Value.GetType(), options);
        }
    }
}
