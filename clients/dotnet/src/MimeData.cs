// @oagen-ignore-file
//
// Hand-maintained file. The generator emits this on first run and then
// leaves it alone; spec changes do not touch it. Mirrors the ergonomic
// MimeData input handling from the Python (`prepare_mime_document`),
// Node (`coerceMimeData`), Go (`InferMIMEData`), and Rust (`From<T>` impls)
// SDKs.

using System;
using System.IO;
using System.Text.Json.Serialization;

namespace Retab
{
    /// <summary>Wire-shape MimeData. Mirrors the spec's <c>MIMEData</c> component schema.</summary>
    /// <remarks>
    /// Customers rarely build this directly. Pass a <see cref="FileInfo"/>,
    /// <see cref="byte"/>[] array, <see cref="Uri"/>, or call one of the
    /// <c>FromX(...)</c> factory methods. Implicit conversion operators handle
    /// the common cases at the call site.
    /// </remarks>
    public sealed class MimeData
    {
        /// <summary>Filename associated with this document.</summary>
        [JsonPropertyName("filename")]
        public string Filename { get; set; } = string.Empty;

        /// <summary>The document URL — either an https:// link or a base64 data: URL.</summary>
        [JsonPropertyName("url")]
        public string Url { get; set; } = string.Empty;

        /// <summary>Construct an empty MimeData. Prefer the <c>FromX(...)</c> factories.</summary>
        public MimeData() { }

        /// <summary>Construct a MimeData with the given filename and URL.</summary>
        public MimeData(string filename, string url)
        {
            this.Filename = filename;
            this.Url = url;
        }

        /// <summary>Build a MimeData from a local file path.</summary>
        public static MimeData FromFile(FileInfo file)
        {
            if (file == null) throw new ArgumentNullException(nameof(file));
            var bytes = File.ReadAllBytes(file.FullName);
            var mime = DetectMimeType(bytes) ?? "application/octet-stream";
            var dataUrl = $"data:{mime};base64,{Convert.ToBase64String(bytes)}";
            return new MimeData(file.Name, dataUrl);
        }

        /// <summary>Build a MimeData from a local file path (string overload).</summary>
        public static MimeData FromFile(string path) => FromFile(new FileInfo(path));

        /// <summary>Build a MimeData from raw bytes. Content type detected from magic bytes when possible.</summary>
        public static MimeData FromBytes(byte[] bytes, string filename = "document")
        {
            if (bytes == null) throw new ArgumentNullException(nameof(bytes));
            var mime = DetectMimeType(bytes) ?? "application/octet-stream";
            var dataUrl = $"data:{mime};base64,{Convert.ToBase64String(bytes)}";
            return new MimeData(filename, dataUrl);
        }

        /// <summary>Build a MimeData from a remote URL (passthrough).</summary>
        public static MimeData FromUrl(Uri url)
        {
            if (url == null) throw new ArgumentNullException(nameof(url));
            var filename = Path.GetFileName(url.LocalPath);
            if (string.IsNullOrEmpty(filename)) filename = "document";
            return new MimeData(filename, url.ToString());
        }

        /// <summary>Build a MimeData from an already-formed data URL.</summary>
        public static MimeData FromDataUrl(string dataUrl, string filename = "document")
            => new MimeData(filename, dataUrl);

        // ── Implicit conversions for the unambiguous input types ────────────
        // No `implicit operator MimeData(string)` — strings are ambiguous
        // (could be a file path, a URL, or base64 data), so we force callers
        // to use the explicit `FromX(...)` factory. The runtime cost is one
        // extra method call; the readability win is large.

        public static implicit operator MimeData(byte[] bytes) => FromBytes(bytes);
        public static implicit operator MimeData(FileInfo file) => FromFile(file);
        public static implicit operator MimeData(Uri url) => FromUrl(url);

        // ── Content-type detection ─────────────────────────────────────────
        private static string? DetectMimeType(byte[] bytes)
        {
            if (bytes.Length < 4) return null;
            // Magic-bytes detection for the common Retab document types:
            // PDF, PNG, JPEG, GIF, WebP, ZIP (covers docx/xlsx/pptx). Extend as needed.
            if (bytes[0] == 0x25 && bytes[1] == 0x50 && bytes[2] == 0x44 && bytes[3] == 0x46) return "application/pdf";
            if (bytes[0] == 0x89 && bytes[1] == 0x50 && bytes[2] == 0x4E && bytes[3] == 0x47) return "image/png";
            if (bytes[0] == 0xFF && bytes[1] == 0xD8) return "image/jpeg";
            if (bytes[0] == 0x47 && bytes[1] == 0x49 && bytes[2] == 0x46) return "image/gif";
            if (bytes.Length >= 12
                && bytes[0] == 0x52 && bytes[1] == 0x49 && bytes[2] == 0x46 && bytes[3] == 0x46
                && bytes[8] == 0x57 && bytes[9] == 0x45 && bytes[10] == 0x42 && bytes[11] == 0x50) return "image/webp";
            if (bytes[0] == 0x50 && bytes[1] == 0x4B) return "application/zip";  // docx/xlsx/pptx, etc.
            return null;
        }
    }
}
