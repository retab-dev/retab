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

        /// <summary>Only return workflows belonging to this project. Use the shared project's id to list the organization's shared workflows.</summary>
        public string? ProjectId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowsService.CreateAsync"/>: Create Workflow</summary>
    public class WorkflowsCreateOptions : BaseOptions
    {
        /// <summary>The name of the workflow</summary>
        public string? Name { get; set; }

        /// <summary>Description of the workflow</summary>
        public string? Description { get; set; }

        /// <summary>Project that should own this workflow.</summary>
        public string ProjectId { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowsService.ListVersionsAsync"/>: List Workflow Versions</summary>
    public class WorkflowsListVersionsOptions : ListOptions
    {
        /// <summary>Workflow whose versions to list</summary>
        public string WorkflowId { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowsService.ListDiffAsync"/>: Diff Workflow Versions</summary>
    public class WorkflowsListDiffOptions : BaseOptions
    {
        /// <summary>Workflow whose versions to diff</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>Base workflow version ID</summary>
        public string FromWorkflowVersionId { get; set; } = default!;

        /// <summary>Target workflow version ID</summary>
        public string ToWorkflowVersionId { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowsService.GetVersionAsync"/>: Get Workflow Version</summary>
    public class WorkflowsGetVersionOptions : BaseOptions
    {
        /// <summary>Workflow that owns the version. Workflow version ids are content-addressed by executable spec, so workflow_id disambiguates identical specs reused across workflows.</summary>
        public string WorkflowId { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowsService.CreateVersionRestoreAsync"/>: Restore Workflow Version</summary>
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

    /// <summary>Request options for <see cref="WorkflowsService.CreatePlanAsync"/>: Plan Existing Workflow Spec</summary>
    public class WorkflowsCreatePlanOptions : BaseOptions
    {
        /// <summary>Workflow YAML definition</summary>
        public string YamlDefinition { get; set; } = default!;

        /// <summary>Project that should own a workflow created from this spec. Required when applying a spec that creates a new workflow.</summary>
        public string? ProjectId { get; set; }

    }
}
