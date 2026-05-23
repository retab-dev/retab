namespace Retab
{

    /// <summary>Capture one experiment document from workflow execution provenance.</summary>
    public class ExperimentDocumentCaptureRequest
    {
        public string WorkflowRunId { get; set; } = default!;
        public string? StepId { get; set; }
    }
}
