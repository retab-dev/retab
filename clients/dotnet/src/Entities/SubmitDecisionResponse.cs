namespace Retab
{

    /// <summary>Response to a review approve or reject request.</summary>
    /// <remarks>
    /// Carries `resume_status` so callers can see whether the run resumed
    /// successfully.
    /// </remarks>
    public class SubmitDecisionResponse
    {
        public SubmissionStatus? SubmissionStatus { get; set; }
        public Review Review { get; set; } = default!;
        public ResumeStatus? ResumeStatus { get; set; }
        public string? ResumeError { get; set; }

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
