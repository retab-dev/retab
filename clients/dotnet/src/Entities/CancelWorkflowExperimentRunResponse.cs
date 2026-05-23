namespace Retab
{

    /// <summary>Represents a cancel workflow experiment run response.</summary>
    public class CancelWorkflowExperimentRunResponse
    {
        public string Id { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(PendingWorkflowExperimentRunDiscriminatorConverter))]
        public object Lifecycle { get; set; } = default!;
    }
}
