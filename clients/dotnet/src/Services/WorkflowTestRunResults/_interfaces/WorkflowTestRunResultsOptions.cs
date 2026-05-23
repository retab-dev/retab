namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowTestRunResultsService.ListAsync"/>: List Test Execution Results</summary>
    public class WorkflowTestRunResultsListOptions : ListOptions
    {
        public string RunId { get; set; } = default!;

    }
}
