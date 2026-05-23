namespace Retab
{

    /// <summary>Document identity used by compact metric responses.</summary>
    public class ExperimentMetricDocumentRef
    {
        public string Id { get; set; } = default!;
        public string Filename { get; set; } = default!;
    }
}
