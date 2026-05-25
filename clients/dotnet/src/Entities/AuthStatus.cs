namespace Retab
{

    /// <summary>Represents an auth status.</summary>
    public class AuthStatus
    {
        public bool? Authenticated { get; set; }
        public string AuthMethod { get; set; } = default!;
        public AuthStatusEnvironment? Environment { get; set; }
        public AuthStatusKey? Key { get; set; }
    }
}
