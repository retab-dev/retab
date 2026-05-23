namespace Retab
{

    /// <summary>JSON payload for a handle input. ``data`` is the raw JSON value.</summary>
    public class JsonHandleInput
    {
        public string? Type { get; set; }
        public object? Data { get; set; }
    }
}
