namespace Retab
{

    /// <summary>Represents a block exec file handle input.</summary>
    public class BlockExecFileHandleInput
    {
        public BlockExecFileRef Document { get; set; } = default!;
        public string? Type { get; set; } = "file";

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
