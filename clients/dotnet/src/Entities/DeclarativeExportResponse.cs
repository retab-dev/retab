namespace Retab
{

    /// <summary>Represents a declarative export response.</summary>
    public class DeclarativeExportResponse
    {
        public string WorkflowId { get; set; } = default!;
        public string YamlDefinition { get; set; } = default!;
    }
}
