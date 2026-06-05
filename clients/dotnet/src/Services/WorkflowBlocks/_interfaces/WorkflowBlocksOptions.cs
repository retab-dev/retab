namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowBlocksService.ListAsync"/>: List Blocks</summary>
    public class WorkflowBlocksListOptions : ListOptions
    {
        public string WorkflowId { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowBlocksService.CreateAsync"/>: Create Block</summary>
    public class WorkflowBlocksCreateOptions : BaseOptions
    {
        /// <summary>Workflow to create the block in.</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>Block ID. Omit to let the server generate one (recommended). Block IDs must be unique across your organization, not just within a workflow — reusing a custom id like 'block_extract' in more than one workflow fails with 409.</summary>
        public string? Id { get; set; }

        /// <summary>Block type</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public WorkflowBlockCreateRequestType Type { get; set; }

        /// <summary>Display label</summary>
        public string? Label { get; set; }

        /// <summary>X position</summary>
        public double? PositionX { get; set; }

        /// <summary>Y position</summary>
        public double? PositionY { get; set; }

        /// <summary>Block width</summary>
        public double? Width { get; set; }

        /// <summary>Block height</summary>
        public double? Height { get; set; }

        /// <summary>Block configuration</summary>
        public Dictionary<string, object>? Config { get; set; }

        /// <summary>ID of parent container block (while_loop, for_each)</summary>
        public string? ParentId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowBlocksService.ListVersionsAsync"/>: List Block Versions</summary>
    public class WorkflowBlocksListVersionsOptions : ListOptions
    {
        public string WorkflowId { get; set; } = default!;

        /// <summary>Filter by stable block ID</summary>
        public string? BlockId { get; set; }

        /// <summary>Filter by workflow version ID</summary>
        public string? WorkflowVersionId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowBlocksService.ListDiffAsync"/>: Diff Block Versions</summary>
    public class WorkflowBlocksListDiffOptions : BaseOptions
    {
        public string FromBlockVersionId { get; set; } = default!;

        public string ToBlockVersionId { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowBlocksService.GetAsync"/>: Get Block</summary>
    public class WorkflowBlocksGetOptions : BaseOptions
    {
        /// <summary>Disambiguates a block id that is shared by more than one workflow. Required only when the block id is not unique within your organization — otherwise the call returns 409 listing the colliding workflow_ids. Server-generated block IDs are always unique and never need this.</summary>
        public string? WorkflowId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowBlocksService.UpdateAsync"/>: Update Block</summary>
    public class WorkflowBlocksUpdateOptions : BaseOptions
    {
        public string? Label { get; set; }

        public double? PositionX { get; set; }

        public double? PositionY { get; set; }

        public double? Width { get; set; }

        public double? Height { get; set; }

        public Dictionary<string, object>? Config { get; set; }

        public string? ParentId { get; set; }

        /// <summary>How to apply the `config` field. 'merge' (default) deep-merges the patch into the existing config with null-as-delete; 'replace' uses the patch as the full new config.</summary>
        public UpdateWorkflowBlockRequestConfigMode? ConfigMode { get; set; }

        /// <summary>Disambiguates a block id that is shared by more than one workflow. Required only when the block id is not unique within your organization.</summary>
        public string? WorkflowId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowBlocksService.DeleteAsync"/>: Delete Block</summary>
    public class WorkflowBlocksDeleteOptions : BaseOptions
    {
        /// <summary>Disambiguates a block id that is shared by more than one workflow. Required only when the block id is not unique within your organization.</summary>
        public string? WorkflowId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowBlocksService.CreateBlockValidateConfigAsync"/>: Validate Block Config</summary>
    public class WorkflowBlocksCreateBlockValidateConfigOptions : BaseOptions
    {
        /// <summary>Assembled block config to validate.</summary>
        public Dictionary<string, object> Config { get; set; } = default!;

        /// <summary>How to apply the config before validation. 'replace' validates the config as the full block config; 'merge' validates the result of merging it into the existing block config.</summary>
        public UpdateWorkflowBlockRequestConfigMode? ConfigMode { get; set; }

        /// <summary>Workflow ID to disambiguate legacy duplicate block IDs. Omit for normal server-generated block IDs.</summary>
        public string? WorkflowId { get; set; }

    }
}
