namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Represents an edit template.</summary>
    public class EditTemplate
    {

        /// <summary>Unique identifier of the template.</summary>
        public string Id { get; set; } = default!;

        /// <summary>Name of the template.</summary>
        public string Name { get; set; } = default!;

        /// <summary>File information for the empty PDF template.</summary>
        public FileRef File { get; set; } = default!;

        /// <summary>Form fields attached to the template.</summary>
        public List<FormField>? FormFields { get; set; }

        /// <summary>Number of form fields in the template.</summary>
        public long? FieldCount { get; set; }

        /// <summary>Timestamp of creation.</summary>
        public DateTimeOffset CreatedAt { get; set; }

        /// <summary>Timestamp of last update.</summary>
        public DateTimeOffset UpdatedAt { get; set; }
    }
}
