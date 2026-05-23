namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Gate fires if ANY child predicate fires. Evaluated in list order;</summary>
    /// <remarks>
    /// `triggered_by` reports the first match (decision: first-match wins).
    /// </remarks>
    public class ReviewAnyOf
    {
        public string? Kind { get; set; }
        public List<object> Predicates { get; set; } = default!;
    }
}
