namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="SchemasService.GenerateAsync"/>: Generate Schema From Examples</summary>
    public class SchemasGenerateOptions : BaseOptions
    {
        public List<MimeData> Documents { get; set; } = default!;

        public string? Model { get; set; }

        public string? Instructions { get; set; }

        /// <summary>Resolution of the image sent to the LLM</summary>
        public long? ImageResolutionDpi { get; set; }

        public bool? Stream { get; set; }

    }
}
