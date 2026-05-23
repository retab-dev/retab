namespace Retab
{

    /// <summary>The step failed. Error details are bundled into lifecycle.</summary>
    public class ErrorStepLifecycle
    {
        public string? Status { get; set; }

        /// <summary>Human-readable error message</summary>
        public string Message { get; set; } = default!;
        public ErrorStepLifecycleStage? Stage { get; set; }
        public ErrorStepLifecycleCategory? Category { get; set; }
        public ErrorDetails? Details { get; set; }
    }
}
