namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="UsageService.ListBlocksAsync"/>: List Usage Blocks</summary>
    public class UsageListBlocksOptions : ListOptions
    {
        /// <summary>Filter to a single workflow id.</summary>
        public string? WorkflowId { get; set; }

        /// <summary>Filter by block type: extract, classify, split, parse, edit, or partition. These are the primitive-kind names recorded on the execution, which differ from the workflow block type enum (a classifier block records classify). An unknown value is rejected with 422.</summary>
        public string? BlockType { get; set; }

        /// <summary>Inclusive activity lower bound (YYYY-MM-DD, UTC).</summary>
        public string? FromDate { get; set; }

        /// <summary>Inclusive activity upper bound (YYYY-MM-DD, UTC).</summary>
        public string? ToDate { get; set; }

    }

    /// <summary>Request options for <see cref="UsageService.ListPrimitivesAsync"/>: List Usage Primitives</summary>
    public class UsageListPrimitivesOptions : ListOptions
    {
        /// <summary>Scope the export to this environment id within the caller's organization. Defaults to the authenticated identity's environment.</summary>
        public string? EnvironmentId { get; set; }

        /// <summary>Filter to a single workflow id (origin workflow).</summary>
        public string? WorkflowId { get; set; }

        /// <summary>Filter to executions owned by a single project id.</summary>
        public string? ProjectId { get; set; }

        /// <summary>Filter to executions triggered by a single API key id (the api_key_id returned under triggered_by).</summary>
        public string? ApiKeyId { get; set; }

        /// <summary>Filter to executions triggered by a single access token id (the access_token_id returned under triggered_by).</summary>
        public string? AccessTokenId { get; set; }

        /// <summary>Filter to executions triggered by a single user id (the user_id returned under triggered_by).</summary>
        public string? UserId { get; set; }

        /// <summary>Filter to a single workflow run id (origin run).</summary>
        public string? RunId { get; set; }

        /// <summary>Filter to a single workflow block id (origin block).</summary>
        public string? BlockId { get; set; }

        /// <summary>Filter by operation: extraction, classify, split, parse, edit, partition, or schema_generation. The stored-kind aliases extract and classification are also accepted. An unknown value is rejected with 422.</summary>
        public string? Operation { get; set; }

        /// <summary>Filter by execution lifecycle status: created, running, completed, failed, or canceled. Note the single-l spelling: a primitive execution is canceled, whereas a workflow run's status is cancelled. An unknown value is rejected with 422.</summary>
        public string? Status { get; set; }

        /// <summary>Filter by metadata equality: a JSON object of string key/value pairs (e.g. {"tenant":"acme"}). Pairs AND together.</summary>
        public string? Metadata { get; set; }

        /// <summary>Inclusive created_at lower bound (YYYY-MM-DD, UTC).</summary>
        public string? FromDate { get; set; }

        /// <summary>Inclusive created_at upper bound (YYYY-MM-DD, UTC).</summary>
        public string? ToDate { get; set; }

    }

    /// <summary>Request options for <see cref="UsageService.ListRunsAsync"/>: List Usage Runs</summary>
    public class UsageListRunsOptions : ListOptions
    {
        /// <summary>Filter to a single workflow id.</summary>
        public string? WorkflowId { get; set; }

        /// <summary>Filter by lifecycle status: pending, queued, running, completed, error, failed, awaiting_review, or cancelled.</summary>
        public string? Status { get; set; }

        /// <summary>Filter by trigger type: manual, api, schedule, webhook, email, or restart.</summary>
        public string? TriggerType { get; set; }

        /// <summary>Inclusive created_at lower bound (YYYY-MM-DD, UTC).</summary>
        public string? FromDate { get; set; }

        /// <summary>Inclusive created_at upper bound (YYYY-MM-DD, UTC).</summary>
        public string? ToDate { get; set; }

    }
}
