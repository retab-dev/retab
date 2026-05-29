namespace Retab
{
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Result of evaluating ONE assertion against a block's output.</summary>
    /// <remarks>
    /// `outcome` is a verdict only — pass / fail / blocked. An execution
    /// error (the assertion couldn't be evaluated because of a type error,
    /// invalid regex, schema validation crash, block execution crash, etc.) is
    /// expressed by `outcome="blocked"` with a populated `failure` whose
    /// `code` identifies the specific failure mode (`execution_error`,
    /// `type_error`, `invalid_regex`, `schema_invalid`,
    /// `block_execution_failed`, ...).
    /// </remarks>
    public class AssertionResult
    {
        public string AssertionId { get; set; } = default!;
        public string ConditionKind { get; set; } = default!;
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public AssertionOutcome Outcome { get; set; }
        public object? ActualValue { get; set; }
        public object? ExpectedValue { get; set; }
        public double? Score { get; set; }
        public double? Threshold { get; set; }
        public string? MetricKind { get; set; }
        public string? AssertionLabel { get; set; }
        public AssertionFailure? Failure { get; set; }

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
