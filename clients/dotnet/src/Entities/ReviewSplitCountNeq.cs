namespace Retab
{

    /// <summary>Gate when the number of resulting splits != `expected`.</summary>
    public class ReviewSplitCountNeq
    {
        public string? Kind { get; set; }
        public long Expected { get; set; }
    }
}
