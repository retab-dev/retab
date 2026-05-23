namespace Retab
{

    /// <summary>Block-test assertion against one declared output handle.</summary>
    /// <remarks>
    /// ``target`` is the only supported shape: an output handle id and an
    /// optional relative path inside that handle's payload.
    /// </remarks>
    public class AssertionSpec
    {
        public string? Id { get; set; }
        public OutputTarget Target { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(ExistConditionDiscriminatorConverter))]
        public object Condition { get; set; } = default!;
        public string? Label { get; set; }
    }
}
