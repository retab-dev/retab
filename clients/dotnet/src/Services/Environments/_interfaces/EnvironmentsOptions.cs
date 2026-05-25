namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="EnvironmentsService.CreateAsync"/>: Create Organization Environment</summary>
    public class EnvironmentsCreateOptions : BaseOptions
    {
        public string Name { get; set; } = default!;

        public AuthStatusEnvironmentType? Type { get; set; }

    }

    /// <summary>Request options for <see cref="EnvironmentsService.UpdateAsync"/>: Update Organization Environment</summary>
    public class EnvironmentsUpdateOptions : BaseOptions
    {
        public string? Name { get; set; }

    }
}
