namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowTestsService.ListAsync"/>: List Tests</summary>
    public class WorkflowTestsListOptions : ListOptions
    {
        public string WorkflowId { get; set; } = default!;

        public string? TargetBlockId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowTestsService.CreateAsync"/>: Create Test</summary>
    public class WorkflowTestsCreateOptions : BaseOptions
    {
        public string WorkflowId { get; set; } = default!;

        public WorkflowTestBlockTarget Target { get; set; } = default!;

        public object Source { get; set; } = default!;

        public string? Name { get; set; }

        public AssertionSpec Assertion { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowTestsService.UpdateAsync"/>: Update Test</summary>
    public class WorkflowTestsUpdateOptions : BaseOptions
    {
        public string? Name { get; set; }

        public AssertionSpec? Assertion { get; set; }

        public object? Source { get; set; }

    }
}
