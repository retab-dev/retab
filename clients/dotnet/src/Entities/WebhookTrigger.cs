namespace Retab
{

    /// <summary>Run started by an inbound webhook.</summary>
    public class WebhookTrigger
    {
        public string? Type { get; set; }

        /// <summary>ID of the webhook configuration, when known</summary>
        public string? WebhookId { get; set; }
    }
}
