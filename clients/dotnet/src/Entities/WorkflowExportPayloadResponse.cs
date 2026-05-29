namespace Retab
{

    /// <summary>The exported data as CSV, with its row and column counts.</summary>
    public class WorkflowExportPayloadResponse
    {

        /// <summary>CSV content</summary>
        public string CsvData { get; set; } = default!;

        /// <summary>Data row count</summary>
        public long Rows { get; set; }

        /// <summary>Column count including fixed columns</summary>
        public long Columns { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();
    }
}
