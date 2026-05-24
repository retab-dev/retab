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
}
