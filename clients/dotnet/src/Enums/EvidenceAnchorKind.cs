namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents evidence anchor kind values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum EvidenceAnchorKind
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "pdf_bbox")]
        PdfBbox,
        [EnumMember(Value = "image_bbox")]
        ImageBbox,
        [EnumMember(Value = "text_span")]
        TextSpan,
        [EnumMember(Value = "spreadsheet_cell")]
        SpreadsheetCell,
        [EnumMember(Value = "csv_cell")]
        CsvCell,
        [EnumMember(Value = "docx_text_span")]
        DocxTextSpan,
        [EnumMember(Value = "docx_table_cell")]
        DocxTableCell,
    }
}
