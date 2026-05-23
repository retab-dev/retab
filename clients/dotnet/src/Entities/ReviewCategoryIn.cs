namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Gate when the predicted category is in `categories` (e.g., review fraud alerts).</summary>
    public class ReviewCategoryIn
    {
        public string? Kind { get; set; }
        public List<string> Categories { get; set; } = default!;
    }
}
