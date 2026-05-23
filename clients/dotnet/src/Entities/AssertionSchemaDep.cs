namespace Retab
{

    /// <summary>Single-rule schema dependency for Level 2 drift detection.</summary>
    public class AssertionSchemaDep
    {
        public string? OutputHandleId { get; set; }
        public string SchemaPath { get; set; } = default!;
        public string SubtreeHash { get; set; } = default!;
        public bool? DependsOnRoot { get; set; }
    }
}
