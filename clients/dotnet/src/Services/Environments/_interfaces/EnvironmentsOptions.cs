namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="EnvironmentsService.ListAsync"/>: List Organization Environments</summary>
    public class EnvironmentsListOptions : ListOptions
    {
    }

    /// <summary>Request options for <see cref="EnvironmentsService.CreateAsync"/>: Create Organization Environment</summary>
    public class EnvironmentsCreateOptions : BaseOptions
    {
        public string Name { get; set; } = default!;

        public AuthStatusEnvironmentType? Type { get; set; }

    }
}
