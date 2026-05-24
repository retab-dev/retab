namespace Retab
{

    /// <summary>Run started by an inbound email.</summary>
    public class EmailTrigger
    {
        public string? Type { get; set; }

        /// <summary>Sender email address, when known</summary>
        public string? Sender { get; set; }

        /// <summary>Email subject, when known</summary>
        public string? Subject { get; set; }
    }
}
