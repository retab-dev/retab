namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Represents an api call invocation.</summary>
    public class ApiCallInvocation
    {

        /// <summary>Artifact operation that determines the backing record type</summary>
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
