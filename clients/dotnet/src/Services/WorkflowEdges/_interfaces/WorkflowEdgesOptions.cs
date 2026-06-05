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

        /// <summary>Edge ID. Omit to let the server generate one (recommended).</summary>
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

    /// <summary>Request options for <see cref="WorkflowEdgesService.ListVersionsAsync"/>: List Edge Versions</summary>
    public class WorkflowEdgesListVersionsOptions : ListOptions
    {
        public string WorkflowId { get; set; } = default!;

        /// <summary>Filter by stable edge ID</summary>
        public string? EdgeId { get; set; }

        /// <summary>Filter by workflow version ID</summary>
        public string? WorkflowVersionId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowEdgesService.ListDiffAsync"/>: Diff Edge Versions</summary>
    public class WorkflowEdgesListDiffOptions : BaseOptions
    {
        public string FromEdgeVersionId { get; set; } = default!;

        public string ToEdgeVersionId { get; set; } = default!;

    }
}
