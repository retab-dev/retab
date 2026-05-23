namespace Retab
{

    /// <summary>Terminal: the executed block raised. ``message`` is the executor's</summary>
    /// <remarks>
    /// error string.
    /// </remarks>
    public class ErrorBlockExecutionLifecycle
    {
        public string? Status { get; set; }

        /// <summary>Human-readable error message</summary>
        public string Message { get; set; } = default!;
    }
}
