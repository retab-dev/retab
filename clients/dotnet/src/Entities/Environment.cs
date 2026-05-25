namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents an environment.</summary>
    public class Environment
    {
        public string Id { get; set; } = default!;
        public string Name { get; set; } = default!;
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public AuthStatusEnvironmentType Type { get; set; }
        public bool? IsDefault { get; set; }
        public DateTimeOffset? CreatedAt { get; set; }
        public DateTimeOffset? UpdatedAt { get; set; }
    }
}
