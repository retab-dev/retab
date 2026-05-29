namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Record of an API-call block's outbound HTTP request during a run.</summary>
    /// <remarks>
    /// Lists each request `attempts` made (including retries) and any `error`
    /// if the call ultimately failed.
    /// </remarks>
    public class ApiCallInvocation
    {

        /// <summary>The operation that produced this artifact</summary>
        public string? Operation { get; set; } = "api_call_invocation";
        public string Id { get; set; } = default!;
        public string RunId { get; set; } = default!;
        public string StepId { get; set; } = default!;
        public List<ApiCallAttempt>? Attempts { get; set; }
        public ErrorDetails? Error { get; set; }

        /// <summary>Timestamp when this artifact was created.</summary>
        public DateTimeOffset? CreatedAt { get; set; }

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
