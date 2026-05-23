namespace Retab
{

    /// <summary>Run started by an inbound email message.</summary>
    public class EmailTrigger
    {
        public string? Type { get; set; }

        /// <summary>Sender email address</summary>
        public string Sender { get; set; } = default!;

        /// <summary>Email subject line</summary>
        public string? Subject { get; set; }
    }
}
