namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowEvalsService.ListAsync"/>: List Workflow Evals</summary>
    public class WorkflowEvalsListOptions : ListOptions
    {
        public string WorkflowId { get; set; } = default!;

        public string? TargetBlockId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowEvalsService.CreateAsync"/>: Create Workflow Eval</summary>
    public class WorkflowEvalsCreateOptions : BaseOptions
    {
        public string WorkflowId { get; set; } = default!;

        public EvalRunBlockTarget Target { get; set; } = default!;

        public object Source { get; set; } = default!;

        public string? Name { get; set; }

        public AssertionSpec Assertion { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowEvalsService.UpdateAsync"/>: Update Workflow Eval</summary>
    public class WorkflowEvalsUpdateOptions : BaseOptions
    {
        public string? Name { get; set; }

        public AssertionSpec? Assertion { get; set; }

        public object? Source { get; set; }

    }
}
