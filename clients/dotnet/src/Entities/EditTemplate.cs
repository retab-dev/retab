namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>A reusable edit template: an empty PDF and the `form_fields` defined on it.</summary>
    public class EditTemplate
    {

        /// <summary>Unique identifier of the template.</summary>
        public string Id { get; set; } = default!;

        /// <summary>Name of the template.</summary>
        public string Name { get; set; } = default!;

        /// <summary>File information for the empty PDF template.</summary>
        public BlockExecFileRef File { get; set; } = default!;

        /// <summary>Form fields attached to the template.</summary>
        public List<FormField>? FormFields { get; set; }

        /// <summary>Number of form fields in the template.</summary>
        public long? FieldCount { get; set; }

        /// <summary>Timestamp of creation.</summary>
        public DateTimeOffset? CreatedAt { get; set; }

        /// <summary>Timestamp of last update.</summary>
        public DateTimeOffset? UpdatedAt { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();
    }
}
