namespace Retab
{

    /// <summary>Represents an auth status response.</summary>
    public class AuthStatusResponse
    {
        public bool? Authenticated { get; set; }
        public string AuthMethod { get; set; } = default!;
        public string? OrganizationId { get; set; }
        public AuthStatusEnvironment? Environment { get; set; }
        public AuthStatusKey? Key { get; set; }
    }
}
