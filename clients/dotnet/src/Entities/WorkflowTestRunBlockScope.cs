namespace Retab
{

    /// <summary>Run every workflow test for one block in the workflow.</summary>
    public class WorkflowTestRunBlockScope
    {
        public string Type { get; internal set; } = "block";
        public string BlockId { get; set; } = default!;
    }
}
