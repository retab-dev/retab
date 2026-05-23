namespace Retab
{

    /// <summary>Optional request payload for cancel workflow command idempotency.</summary>
    public class CancelWorkflowRequest
    {

        /// <summary>Optional idempotency key for deduplicating cancel commands</summary>
        public string? CommandId { get; set; }
    }
}
