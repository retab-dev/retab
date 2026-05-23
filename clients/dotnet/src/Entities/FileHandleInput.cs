namespace Retab
{

    /// <summary>File reference for a handle input.</summary>
    public class FileHandleInput
    {
        public string? Type { get; set; }
        public MaterializedDocument Document { get; set; } = default!;
    }
}
