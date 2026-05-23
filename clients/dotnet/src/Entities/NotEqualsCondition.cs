namespace Retab
{

    /// <summary>Represents a not equals condition.</summary>
    public class NotEqualsCondition
    {
        public string? Kind { get; set; }
        public object Expected { get; set; } = default!;
    }
}
