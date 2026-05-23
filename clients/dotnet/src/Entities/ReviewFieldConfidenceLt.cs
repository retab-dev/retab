namespace Retab
{

    /// <summary>Gate when the field at `path` has confidence below `threshold`.</summary>
    public class ReviewFieldConfidenceLt
    {
        public string? Kind { get; set; }

        /// <summary>JSONPath-style path, e.g. '$.invoice.total' or 'invoice.total'</summary>
        public string Path { get; set; } = default!;
        public double Threshold { get; set; }
    }
}
