namespace Retab
{
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>One actor shape for humans, agents, and models.</summary>
    public class Actor
    {
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public ActorKind Kind { get; set; }
        public string Id { get; set; } = default!;
        public string DisplayName { get; set; } = default!;
    }
}
