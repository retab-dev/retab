namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a declarative plan change.</summary>
    public class DeclarativePlanChange
    {
        public object? Before { get; set; }
        public object? After { get; set; }
        public object? BeforeSensitive { get; set; }
        public object? AfterSensitive { get; set; }
        public List<DeclarativePlanFieldChange>? FieldChanges { get; set; }

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
