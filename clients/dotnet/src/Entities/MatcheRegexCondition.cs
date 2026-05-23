namespace Retab
{

    /// <summary>Represents a matche regex condition.</summary>
    public class MatcheRegexCondition
    {
        public string? Kind { get; set; }
        public string Pattern { get; set; } = default!;
    }
}
