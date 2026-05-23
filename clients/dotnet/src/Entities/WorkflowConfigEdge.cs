namespace Retab
{

    /// <summary>Represents a workflow config edge.</summary>
    public class WorkflowConfigEdge
    {
        public string? Id { get; set; }

        /// <summary>ID of the source block</summary>
        public string Source { get; set; } = default!;

        /// <summary>ID of the target block</summary>
        public string Target { get; set; } = default!;
        public string? SourceHandle { get; set; }
        public string? TargetHandle { get; set; }
        public bool? Animated { get; set; }
    }
}
