namespace Retab
{

    /// <summary>The run was cancelled before reaching a natural terminal state.</summary>
    public class CancelledTerminal
    {
        public string? Status { get; set; }

        /// <summary>Human-readable reason, when known</summary>
        public string? Reason { get; set; }
    }
}
