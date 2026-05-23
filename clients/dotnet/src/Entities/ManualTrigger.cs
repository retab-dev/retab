namespace Retab
{

    /// <summary>Manual run started by a user from the dashboard.</summary>
    public class ManualTrigger
    {
        public string? Type { get; set; }

        /// <summary>User who started the run, when known</summary>
        public string? UserId { get; set; }
    }
}
