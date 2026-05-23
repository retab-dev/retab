namespace Retab
{
    using System;

    /// <summary>Public live workflow edge object.</summary>
    public class WorkflowEdgeDoc
    {
        public string Id { get; set; } = default!;

        /// <summary>Foreign key to workflow</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>ID of the source block</summary>
        public string SourceBlock { get; set; } = default!;

        /// <summary>ID of the target block</summary>
        public string TargetBlock { get; set; } = default!;

        /// <summary>Output handle on source block</summary>
        public string? SourceHandle { get; set; }

        /// <summary>Input handle on target block</summary>
        public string? TargetHandle { get; set; }
        public DateTimeOffset UpdatedAt { get; set; }
    }
}
