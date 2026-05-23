namespace Retab
{

    /// <summary>Per-lifecycle counts for a batch of block-test runs.</summary>
    public class BlockTestLifecycleCounts
    {
        public long? Pending { get; set; }
        public long? Queued { get; set; }
        public long? Running { get; set; }
        public long? Completed { get; set; }
        public long? Error { get; set; }
        public long? Cancelled { get; set; }
    }
}
