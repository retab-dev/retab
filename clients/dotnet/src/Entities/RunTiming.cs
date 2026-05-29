namespace Retab
{
    using System;

    /// <summary>Timing information for a run.</summary>
    /// <remarks>
    /// Three event timestamps that consumers cannot reconstruct on their own.
    /// Wall-clock duration is a trivial `completed_at - started_at` subtraction
    /// done client-side; it is not stored or exposed.
    /// </remarks>
    public class RunTiming
    {

        /// <summary>When the run record was created</summary>
        public DateTimeOffset? CreatedAt { get; set; }

        /// <summary>When the run started executing</summary>
        public DateTimeOffset? StartedAt { get; set; }

        /// <summary>When the run finished executing</summary>
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
