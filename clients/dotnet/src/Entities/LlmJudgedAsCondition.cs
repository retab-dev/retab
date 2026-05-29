namespace Retab
{

    /// <summary>Represents a llm judged as condition.</summary>
    public class LlmJudgedAsCondition
    {
        public string? Kind { get; set; }
        public string Rubric { get; set; } = default!;
        public string? ExpectedLabel { get; set; }

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
