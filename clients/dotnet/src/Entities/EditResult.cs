namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents an edit result.</summary>
    public class EditResult
    {

        /// <summary>Filled form fields (positions, descriptions, and filled values).</summary>
        public List<FormField> FormData { get; set; } = default!;

        /// <summary>PDF with the filled form values rendered in.</summary>
        public MimeData FilledDocument { get; set; } = default!;
    }
}
