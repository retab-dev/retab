namespace Retab
{

    /// <summary>Represents a category.</summary>
    public class Category
    {

        /// <summary>The name of the category</summary>
        public string Name { get; set; } = default!;

        /// <summary>Stable machine key used by workflow classifier output handles</summary>
        public string? HandleKey { get; set; }

        /// <summary>The description of the category</summary>
        public string? Description { get; set; }

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
