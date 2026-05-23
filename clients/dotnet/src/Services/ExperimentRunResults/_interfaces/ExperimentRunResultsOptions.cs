namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="ExperimentRunResultsService.ListAsync"/>: List Experiment Results</summary>
    public class ExperimentRunResultsListOptions : ListOptions
    {
        public string RunId { get; set; } = default!;

    }
}
