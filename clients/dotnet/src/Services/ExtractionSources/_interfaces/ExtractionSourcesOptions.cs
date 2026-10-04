namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="ExtractionSourcesService.GetAsync"/>: Get Extraction Sources</summary>
    public class ExtractionSourcesGetOptions : BaseOptions
    {
        /// <summary>Opt into progressive sources for this mode. Omit to retain the legacy synchronous response.</summary>
        public CreateSourcesRequestMode? Mode { get; set; }

        /// <summary>Expected job identity returned by POST. A changed extraction returns 409.</summary>
        public string? JobId { get; set; }

    }

    /// <summary>Request options for <see cref="ExtractionSourcesService.CreateAsync"/>: Create Extraction Sources</summary>
    public class ExtractionSourcesCreateOptions : BaseOptions
    {
        /// <summary>Return immediately when true. Otherwise wait up to 20 seconds, then return 202 with Location while the same durable job continues.</summary>
        public bool? Background { get; set; }

        /// <summary>Optional expected sources job identity. Returns 409 if the extraction changed.</summary>
        public string? JobId { get; set; }

        public CreateSourcesRequestMode? Mode { get; set; }

        /// <summary>Retry a terminal failed or incomplete computation. Concurrent requests join the current attempt; billing remains once per extraction.</summary>
        public bool? Retry { get; set; }

    }
}
