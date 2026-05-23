namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowReviewVersionsService.ListAsync"/>: List Review Versions Route</summary>
    public class WorkflowReviewVersionsListOptions : ListOptions
    {
        /// <summary>Required: the review whose versions to list.</summary>
        public string ReviewId { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowReviewVersionsService.CreateAsync"/>: Create Review Version Route</summary>
    public class WorkflowReviewVersionsCreateOptions : BaseOptions
    {
        public string ReviewId { get; set; } = default!;

        public string ParentId { get; set; } = default!;

        /// <summary>The full reviewed snapshot to store as an immutable version. The object must match the gated block type: extract uses the raw output object; classifier uses {'category': string}; split uses {'documents': [{'name': string, 'pages': positive sorted int[]}]}; for_each uses {'partitions': [{'key': string, 'pages': positive sorted int[]}]}. The server validates the shape and stores the exact submitted object when valid.</summary>
        public Dictionary<string, object> Snapshot { get; set; } = default!;

        public string? Note { get; set; }

    }
}
