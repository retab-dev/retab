namespace Retab
{
    using System.Collections.Generic;

    /// <summary>The run is paused on at least one gated block.</summary>
    public class AwaitingReviewRun
    {
        public string? Status { get; set; }

        /// <summary>Block IDs that are waiting for review</summary>
        public List<string>? WaitingForBlockIds { get; set; }

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
