namespace Retab
{

    /// <summary>Represents an all items match condition.</summary>
    public class AllItemsMatchCondition
    {
        public string? Kind { get; set; }
        [Newtonsoft.Json.JsonConverter(typeof(ExistConditionDiscriminatorConverter))]
        public object Condition { get; set; } = default!;
    }
}
