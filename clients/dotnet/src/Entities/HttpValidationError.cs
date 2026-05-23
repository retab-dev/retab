namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a http validation error.</summary>
    public class HttpValidationError
    {
        public List<ValidationError>? Detail { get; set; }
    }
}
