namespace Retab
{

    /// <summary>Represents a not contains condition.</summary>
    public class NotContainsCondition
    {
        public string? Kind { get; set; }
        public object Expected { get; set; } = default!;
    }
}
