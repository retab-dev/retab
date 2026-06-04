namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowsService.ListAsync"/>: List Workflows</summary>
    public class WorkflowsListOptions : ListOptions
    {
        public string? SortBy { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowsService.CreateAsync"/>: Create Workflow</summary>
    public class WorkflowsCreateOptions : BaseOptions
    {
        /// <summary>The name of the workflow</summary>
        public string? Name { get; set; }

        /// <summary>Description of the workflow</summary>
        public string? Description { get; set; }

        /// <summary>Project that should own this workflow. Omit to use the organization's shared workflows project.</summary>
        public string? ProjectId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowsService.ListVersionsAsync"/>: List Workflow Versions Route</summary>
    public class WorkflowsListVersionsOptions : BaseOptions
    {
        /// <summary>Workflow whose versions to list</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>Maximum number of versions to return</summary>
        public long? Limit { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowsService.ListDiffAsync"/>: Diff Workflow Versions Route</summary>
    public class WorkflowsListDiffOptions : BaseOptions
    {
        /// <summary>Workflow whose versions to diff</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>Base workflow version ID</summary>
        public string FromWorkflowVersionId { get; set; } = default!;

        /// <summary>Target workflow version ID</summary>
        public string ToWorkflowVersionId { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowsService.GetVersionAsync"/>: Get Workflow Version Route</summary>
    public class WorkflowsGetVersionOptions : BaseOptions
    {
        /// <summary>Workflow that owns the version. Workflow version ids are content-addressed by executable spec, so workflow_id disambiguates identical specs reused across workflows.</summary>
        public string WorkflowId { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowsService.CreateVersionRestoreAsync"/>: Restore Workflow Version Route</summary>
    public class WorkflowsCreateVersionRestoreOptions : BaseOptions
    {
        /// <summary>Workflow to restore into a new draft</summary>
        public string WorkflowId { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowsService.UpdateAsync"/>: Update Workflow</summary>
    public class WorkflowsUpdateOptions : BaseOptions
    {
        /// <summary>The name of the workflow</summary>
        public string? Name { get; set; }

        /// <summary>Description of the workflow</summary>
        public string? Description { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowsService.PublishAsync"/>: Publish Workflow</summary>
    public class WorkflowsPublishOptions : BaseOptions
    {
    }

    /// <summary>Request options for <see cref="WorkflowsService.CreatePlanAsync"/>: Plan Workflow Spec For Existing Workflow</summary>
    public class WorkflowsCreatePlanOptions : BaseOptions
    {
        /// <summary>Workflow YAML definition</summary>
        public string YamlDefinition { get; set; } = default!;

    }
}
