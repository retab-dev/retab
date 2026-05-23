namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a parse output.</summary>
    public class ParseOutput
    {

        /// <summary>Text content of each page (1-indexed order)</summary>
        public List<string> Pages { get; set; } = default!;

        /// <summary>Concatenated text content of the full document</summary>
        public string Text { get; set; } = default!;
    }
}
