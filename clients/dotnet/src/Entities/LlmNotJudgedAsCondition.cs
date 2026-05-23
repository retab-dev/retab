namespace Retab
{

    /// <summary>Represents a llm not judged as condition.</summary>
    public class LlmNotJudgedAsCondition
    {
        public string? Kind { get; set; }
        public string Rubric { get; set; } = default!;
        public string? ExpectedLabel { get; set; }
    }
}
