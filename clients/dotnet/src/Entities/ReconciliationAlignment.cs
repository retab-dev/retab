namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a reconciliation alignment.</summary>
    public class ReconciliationAlignment
    {
        public List<Dictionary<string, object>>? AlignedInputs { get; set; }
        public List<ReconciliationPathAlignment>? PathAlignments { get; set; }
        public long ReferenceIndex { get; set; }

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
