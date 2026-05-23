namespace Retab
{

    /// <summary>Public workflow-test target.</summary>
    /// <remarks>
    /// The storage layer remains block-scoped today, but the API shape names the
    /// tested entity explicitly so workflow-level targets can be added later.
    /// </remarks>
    public class WorkflowTestBlockTarget
    {
        public string? Type { get; set; }
        public string BlockId { get; set; } = default!;
    }
}
