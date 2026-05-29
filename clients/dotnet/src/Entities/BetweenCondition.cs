namespace Retab
{

    /// <summary>Represents a between condition.</summary>
    public class BetweenCondition
    {
        public string? Kind { get; set; }
        public OneOf.OneOf<long, double> Lower { get; set; } = default!;
        public OneOf.OneOf<long, double> Upper { get; set; } = default!;
        public bool? Inclusive { get; set; }

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
