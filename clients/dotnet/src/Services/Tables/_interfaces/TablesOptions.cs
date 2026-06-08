namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="TablesService.ListAsync"/>: Table.List</summary>
    public class TablesListOptions : BaseOptions
    {
        /// <summary>Only return tables belonging to this project.</summary>
        public string? ProjectId { get; set; }

    }

    /// <summary>Request options for <see cref="TablesService.CreateAsync"/>: Table.Create</summary>
    public class TablesCreateOptions : BaseOptions
    {
        public string Name { get; set; } = default!;

        public string File { get; set; } = default!;

        public string? ColumnSchemaOverrides { get; set; }

        public string? ProjectId { get; set; }

    }

    /// <summary>Request options for <see cref="TablesService.ReplaceAsync"/>: Table.Replace</summary>
    public class TablesReplaceOptions : BaseOptions
    {
        public string File { get; set; } = default!;

        public string? ColumnSchemaOverrides { get; set; }

    }

    /// <summary>Request options for <see cref="TablesService.UpdateAsync"/>: Table.Update</summary>
    public class TablesUpdateOptions : BaseOptions
    {
        public string? Name { get; set; }

        public Dictionary<string, object>? Metadata { get; set; }

    }

    /// <summary>Request options for <see cref="TablesService.ProfileAsync"/>: Table.Get Profile</summary>
    public class TablesProfileOptions : BaseOptions
    {
        public List<string>? Select { get; set; }

    }

    /// <summary>Request options for <see cref="TablesService.QueryAsync"/>: Table.Query</summary>
    public class TablesQueryOptions : BaseOptions
    {
    }

    /// <summary>Request options for <see cref="TablesService.ValidateAsync"/>: Table.Validate</summary>
    public class TablesValidateOptions : BaseOptions
    {
        public List<string>? RequiredColumns { get; set; }

        public Dictionary<string, WorkflowTableValidationColumnRule>? Columns { get; set; }

        public List<List<string>>? Unique { get; set; }

    }
}
