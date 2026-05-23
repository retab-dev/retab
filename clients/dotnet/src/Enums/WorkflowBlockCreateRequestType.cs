namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow block create request type values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowBlockCreateRequestType
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "start_document")]
        StartDocument,
        [EnumMember(Value = "start_json")]
        StartJson,
        [EnumMember(Value = "note")]
        Note,
        [EnumMember(Value = "parse")]
        Parse,
        [EnumMember(Value = "edit")]
        Edit,
        [EnumMember(Value = "extract")]
        Extract,
        [EnumMember(Value = "split")]
        Split,
        [EnumMember(Value = "classifier")]
        Classifier,
        [EnumMember(Value = "conditional")]
        Conditional,
        [EnumMember(Value = "api_call")]
        ApiCall,
        [EnumMember(Value = "function")]
        Function,
        [EnumMember(Value = "while_loop")]
        WhileLoop,
        [EnumMember(Value = "for_each")]
        ForEach,
        [EnumMember(Value = "merge_dicts")]
        MergeDicts,
    }
}
