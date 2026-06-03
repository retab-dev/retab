namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="TablesService.CreateAsync"/>: Create Table</summary>
    public class TablesCreateOptions : BaseOptions
    {
        public string Name { get; set; } = default!;

        public string File { get; set; } = default!;

        public string? ColumnSchemaOverrides { get; set; }

    }

    /// <summary>Request options for <see cref="TablesService.ReplaceAsync"/>: Replace Table</summary>
    public class TablesReplaceOptions : BaseOptions
    {
        public string File { get; set; } = default!;

        public string? ColumnSchemaOverrides { get; set; }

    }

    /// <summary>Request options for <see cref="TablesService.UpdateAsync"/>: Update Table</summary>
    public class TablesUpdateOptions : BaseOptions
    {
        public string? Name { get; set; }

        public Dictionary<string, object>? Metadata { get; set; }

    }

    /// <summary>Request options for <see cref="TablesService.ProfileAsync"/>: Profile Table</summary>
    public class TablesProfileOptions : BaseOptions
    {
        public List<string>? Select { get; set; }

    }

    /// <summary>Request options for <see cref="TablesService.QueryAsync"/>: Query Table</summary>
    public class TablesQueryOptions : BaseOptions
    {
    }

    /// <summary>Request options for <see cref="TablesService.ValidateAsync"/>: Validate Table</summary>
    public class TablesValidateOptions : BaseOptions
    {
        public List<string>? RequiredColumns { get; set; }

        public Dictionary<string, WorkflowTableValidationColumnRule>? Columns { get; set; }

        public List<List<string>>? Unique { get; set; }

    }
}
