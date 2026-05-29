namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>One attempt of an api_call (initial + retries).</summary>
    public class ApiCallAttempt
    {

        /// <summary>0-based attempt index</summary>
        public long AttemptNumber { get; set; }
        public string RequestMethod { get; set; } = default!;
        public string RequestUrl { get; set; } = default!;
        public Dictionary<string, string>? RequestHeaders { get; set; }
        public object? RequestBody { get; set; }
        public long? ResponseStatus { get; set; }
        public Dictionary<string, string>? ResponseHeaders { get; set; }
        public object? ResponseBody { get; set; }
        public long? DurationMs { get; set; }
        public ErrorDetails? Error { get; set; }
        public DateTimeOffset? StartedAt { get; set; }
        public DateTimeOffset? CompletedAt { get; set; }

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
