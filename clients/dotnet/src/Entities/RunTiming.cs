namespace Retab
{
    using System;

    /// <summary>All timing information for a run.</summary>
    /// <remarks>
    /// ``duration_ms`` is backfilled at read time from
    /// ``completed_at - started_at`` when both timestamps are present and the
    /// stored value is ``None`` (the Mongo projection helpers in
    /// ``run_duration.py`` compute and persist the canonical value). Records that
    /// already store ``duration_ms`` are left untouched (idempotent), so backfill
    /// cannot drift from the canonical value written by the projection.
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

        /// <summary>Total run duration in milliseconds. Backfilled from ``completed_at - started_at`` on read when not stored.</summary>
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
