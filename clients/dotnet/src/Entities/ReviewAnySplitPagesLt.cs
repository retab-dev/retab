namespace Retab
{

    /// <summary>Gate when any resulting split has fewer than `min_pages` pages.</summary>
    public class ReviewAnySplitPagesLt
    {
        public string? Kind { get; set; }
        public long MinPages { get; set; }
    }
}
