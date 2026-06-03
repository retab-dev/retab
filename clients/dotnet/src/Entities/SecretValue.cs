namespace Retab
{
    using System;

    /// <summary>Represents a secret value.</summary>
    public class SecretValue
    {
        public string Name { get; set; } = default!;
        public string Value { get; set; } = default!;

        /// <summary>When the secret value was last updated.</summary>
        public DateTimeOffset UpdatedAt { get; set; }

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
