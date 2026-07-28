namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="ConsensusService.CreateAsync"/>: Create Consensus</summary>
    public class ConsensusCreateOptions : BaseOptions
    {
        public bool? IncludeAlignment { get; set; }

        public List<Dictionary<string, object>>? Inputs { get; set; }

        public Dictionary<string, object>? JsonSchema { get; set; }

    }
}
