namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="EditTemplatesService.ListAsync"/>: List Templates</summary>
    public class EditTemplatesListOptions : ListOptions
    {
        public string? Name { get; set; }

        public string? SortBy { get; set; }

    }

    /// <summary>Request options for <see cref="EditTemplatesService.CreateAsync"/>: Create Template</summary>
    public class EditTemplatesCreateOptions : BaseOptions
    {
        /// <summary>Name of the template.</summary>
        public string Name { get; set; } = default!;

        /// <summary>The PDF document to use as the empty template.</summary>
        public MimeData Document { get; set; } = default!;

        /// <summary>Form fields to attach to the template.</summary>
        public List<FormField> FormFields { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="EditTemplatesService.UpdateAsync"/>: Update Template</summary>
    public class EditTemplatesUpdateOptions : BaseOptions
    {
        /// <summary>New name for the template.</summary>
        public string? Name { get; set; }

        /// <summary>Replacement list of form fields.</summary>
        public List<FormField>? FormFields { get; set; }

    }
}
