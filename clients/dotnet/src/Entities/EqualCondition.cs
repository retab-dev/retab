namespace Retab
{

    /// <summary>Represents an equal condition.</summary>
    public class EqualCondition
    {
        public string? Kind { get; set; }
        public object Expected { get; set; } = default!;
    }
}
