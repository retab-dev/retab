namespace Retab
{

    /// <summary>Structured context for a single container in the hierarchy.</summary>
    public class ContainerContextData
    {

        /// <summary>Container ID (e.g., 'while_loop-abc')</summary>
        public string ContainerId { get; set; } = default!;

        /// <summary>Iteration index (0-based)</summary>
        public long Iteration { get; set; }

        /// <summary>Whether this container represents a parallel item</summary>
        public bool? IsParallel { get; set; }

        /// <summary>Parallel item index if is_parallel</summary>
        public long? ParallelItemIndex { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();
    }
}
