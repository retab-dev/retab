namespace Retab
{
    using System;

    /// <summary>Represents a secret.</summary>
    public class Secret
    {
        public string Name { get; set; } = default!;

        /// <summary>When the secret was first created.</summary>
        public DateTimeOffset CreatedAt { get; set; }

        /// <summary>When the secret value was last updated.</summary>
        public DateTimeOffset UpdatedAt { get; set; }
        public string? CreatedBy { get; set; }
        public string? UpdatedBy { get; set; }

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
