namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowEdgesService.ListAsync"/>: List Edges</summary>
    public class WorkflowEdgesListOptions : ListOptions
    {
        public string WorkflowId { get; set; } = default!;

        /// <summary>Filter by source block ID</summary>
        public string? SourceBlock { get; set; }

        /// <summary>Filter by target block ID</summary>
        public string? TargetBlock { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowEdgesService.CreateAsync"/>: Create Edge</summary>
    public class WorkflowEdgesCreateOptions : BaseOptions
    {
        /// <summary>Workflow to create the edge in.</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>If omitted, the server generates an opaque ``edg_&lt;nanoid&gt;``. Opaque edge ID. Omit to let the server generate one.</summary>
        public string? Id { get; set; }

        /// <summary>Source block ID</summary>
        public string SourceBlock { get; set; } = default!;

        /// <summary>Target block ID</summary>
        public string TargetBlock { get; set; } = default!;

        /// <summary>Output handle</summary>
        public string? SourceHandle { get; set; }

        /// <summary>Input handle</summary>
        public string? TargetHandle { get; set; }

    }
}
