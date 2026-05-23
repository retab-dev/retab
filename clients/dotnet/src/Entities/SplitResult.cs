namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a split result.</summary>
    public class SplitResult
    {

        /// <summary>The name of the subdocument</summary>
        public string Name { get; set; } = default!;

        /// <summary>The pages of the subdocument (1-indexed)</summary>
        public List<long> Pages { get; set; } = default!;
    }
}
