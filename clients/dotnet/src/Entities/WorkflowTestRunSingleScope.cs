namespace Retab
{

    /// <summary>Run one saved workflow test in the workflow.</summary>
    public class WorkflowTestRunSingleScope
    {
        public string Type { get; internal set; } = "single";
        public string TestId { get; set; } = default!;
    }
}
