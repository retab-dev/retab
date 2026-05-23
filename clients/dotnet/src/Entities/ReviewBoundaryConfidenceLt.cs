namespace Retab
{

    /// <summary>Gate when any split boundary's confidence is below `threshold`.</summary>
    public class ReviewBoundaryConfidenceLt
    {
        public string? Kind { get; set; }
        public double Threshold { get; set; }
    }
}
