namespace Retab
{

    /// <summary>Gate when (top1_prob - top2_prob) &lt; `margin` — model was torn.</summary>
    public class ReviewTopMarginLt
    {
        public string? Kind { get; set; }
        public double Margin { get; set; }
    }
}
