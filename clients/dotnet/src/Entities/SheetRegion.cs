namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a sheet region.</summary>
    public class SheetRegion
    {
        public long? ColEnd { get; set; }
        public long? ColStart { get; set; }
        public List<long>? HeaderRows { get; set; }
        public long RowEnd { get; set; }
        public long RowStart { get; set; }
        public long SheetIndex { get; set; }
        public string SheetName { get; set; } = default!;

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
