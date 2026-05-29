namespace Retab
{
    using Newtonsoft.Json;

    /// <summary>Represents create experiment request n consensus values.</summary>
    [RetabNumberEnum]
    [JsonConverter(typeof(RetabNewtonsoftNumberEnumConverter))]
    public enum CreateExperimentRequestNConsensus
    {
        Unknown = int.MinValue,

        Value3 = 3,
        Value5 = 5,
        Value7 = 7,
    }
}
