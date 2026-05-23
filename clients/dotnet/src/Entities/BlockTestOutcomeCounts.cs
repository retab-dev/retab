namespace Retab
{

    /// <summary>Per-outcome counts. Only completed runs contribute to these buckets.</summary>
    public class BlockTestOutcomeCounts
    {
        public long? Passed { get; set; }
        public long? Failed { get; set; }
        public long? Blocked { get; set; }
    }
}
