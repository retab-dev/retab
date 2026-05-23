namespace Retab
{

    /// <summary>Represents a contain condition.</summary>
    public class ContainCondition
    {
        public string? Kind { get; set; }
        public object Expected { get; set; } = default!;
    }
}
