namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="SecretsService.CreateAsync"/>: Secret.Create</summary>
    public class SecretsCreateOptions : BaseOptions
    {
        public string Name { get; set; } = default!;

        public string Value { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="SecretsService.UpdateAsync"/>: Secret.Set</summary>
    public class SecretsUpdateOptions : BaseOptions
    {
        public string Value { get; set; } = default!;

    }
}
