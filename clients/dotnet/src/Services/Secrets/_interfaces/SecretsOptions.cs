namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="SecretsService.CreateAsync"/>: Create Secret</summary>
    public class SecretsCreateOptions : BaseOptions
    {
        public string Name { get; set; } = default!;

        public string Value { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="SecretsService.UpdateAsync"/>: Set Secret</summary>
    public class SecretsUpdateOptions : BaseOptions
    {
        public string Value { get; set; } = default!;

    }
}
