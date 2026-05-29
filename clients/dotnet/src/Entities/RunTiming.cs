namespace Retab
{
    using System;

    /// <summary>Timing information for a run.</summary>
    /// <remarks>
    /// `duration_ms` is the elapsed time between `started_at` and `completed_at`.
    /// </remarks>
    public class RunTiming
    {

        /// <summary>When the run record was created</summary>
        public DateTimeOffset? CreatedAt { get; set; }

        /// <summary>When the run started executing</summary>
        public DateTimeOffset? StartedAt { get; set; }

        /// <summary>When the run finished executing</summary>
        public DateTimeOffset? CompletedAt { get; set; }

        /// <summary>When the current awaiting_review period started</summary>
        public DateTimeOffset? ReviewWaitingStartedAt { get; set; }

        /// <summary>Accumulated time spent waiting for review across the run</summary>
        public long? AccumulatedReviewWaitingMs { get; set; }

        /// <summary>Total run duration in milliseconds. Backfilled from `completed_at - started_at` on read when not stored.</summary>
        public long? DurationMs { get; set; }

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
