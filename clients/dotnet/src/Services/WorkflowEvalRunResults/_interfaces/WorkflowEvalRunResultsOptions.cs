namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowEvalRunResultsService.ListAsync"/>: List Workflow Eval Results</summary>
    public class WorkflowEvalRunResultsListOptions : ListOptions
    {
        public string RunId { get; set; } = default!;

    }
}
