namespace Retab
{
    using System.Collections.Generic;

    /// <summary>The run is paused on at least one gated block.</summary>
    public class AwaitingReviewRun
    {
        public string? Status { get; set; }

        /// <summary>Block IDs that are waiting for review</summary>
        public List<string>? WaitingForBlockIds { get; set; }
    }
}
