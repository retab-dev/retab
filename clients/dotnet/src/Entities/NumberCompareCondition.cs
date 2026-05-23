namespace Retab
{
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents a number compare condition.</summary>
    public class NumberCompareCondition
    {
        public string? Kind { get; set; }
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public LengthCompareConditionOp Op { get; set; }
        public OneOf.OneOf<long, double> Expected { get; set; } = default!;
    }
}
