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

        /// <summary>If true, run asynchronously: returns immediately with status 'queued'. Poll GET /v1/schemas/generate/{schema_generation_id} until status is terminal.</summary>
        public bool? Background { get; set; }

    }
}
