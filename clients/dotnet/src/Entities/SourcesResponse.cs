namespace Retab
{
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>An extraction's output annotated with the source that backs each value.</summary>
    /// <remarks>
    /// Returned when fetching the sources for an extraction. Carries the canonical
    /// `source_document` and its detected `document_type`, the original `extraction`
    /// output, and a parallel `sources` tree where each leaf is a `{value, source}`
    /// object locating the value in the document (a page region for PDFs, a cell for
    /// spreadsheets, a text span for plain text, and so on). The legacy `file` field is
    /// kept as an alias of `source_document` for compatibility.
    /// </remarks>
    public class SourcesResponse
    {
        public string? Object { get; set; } = "extraction.sources";

        /// <summary>ID of the extraction</summary>
        public string ExtractionId { get; set; } = default!;

        /// <summary>Detected document type of the source document</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public SourcesResponseDocumentType DocumentType { get; set; }

        /// <summary>Compatibility alias for source_document.</summary>
        public BlockExecFileRef File { get; set; } = default!;

        /// <summary>Canonical source document metadata (id, filename, mime_type).</summary>
        public BlockExecFileRef SourceDocument { get; set; } = default!;

        /// <summary>Original extraction output</summary>
        public Dictionary<string, object> Extraction { get; set; } = default!;

        /// <summary>Same shape as extraction but leaves are {value, source} objects. Non-null source entries include file_id.</summary>
        public Dictionary<string, object> Sources { get; set; } = default!;

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();

        /// <summary>
        /// Typed accessor for <see cref="Extraction"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetExtractionAttribute<T>(string key)
        {
            if (this.Extraction == null)
            {
                return default;
            }

            if (!this.Extraction.TryGetValue(key, out var value))
            {
                return default;
            }

            if (value is T typed)
            {
                return typed;
            }

            if (value is Newtonsoft.Json.Linq.JToken token)
            {
                return token.ToObject<T>();
            }

            if (value is System.Text.Json.JsonElement element)
            {
                return System.Text.Json.JsonSerializer.Deserialize<T>(element.GetRawText());
            }

            return default;
        }

        /// <summary>
        /// Typed accessor for <see cref="Sources"/>. Returns the value stored under
        /// <paramref name="key"/> coerced to <typeparamref name="T"/>, or the default
        /// value when the key is missing or the value is not convertible.
        /// </summary>
        /// <typeparam name="T">Expected value type.</typeparam>
        /// <param name="key">The key to look up.</param>
        public T? GetSourcesAttribute<T>(string key)
        {
            if (this.Sources == null)
            {
                return default;
            }

            if (!this.Sources.TryGetValue(key, out var value))
            {
                return default;
            }

            if (value is T typed)
            {
                return typed;
            }

            if (value is Newtonsoft.Json.Linq.JToken token)
            {
                return token.ToObject<T>();
            }

            if (value is System.Text.Json.JsonElement element)
            {
                return System.Text.Json.JsonSerializer.Deserialize<T>(element.GetRawText());
            }

            return default;
        }
    }
}
