namespace Retab
{

    /// <summary>Run created by restarting a parent run.</summary>
    public class RestartTrigger
    {
        public string? Type { get; set; }

        /// <summary>ID of the parent run that was restarted</summary>
        public string ParentRunId { get; set; } = default!;
    }
}
