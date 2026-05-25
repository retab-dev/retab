namespace Retab
{

    /// <summary>Represents an auth status environment.</summary>
    public class AuthStatusEnvironment
    {
        public string Id { get; set; } = default!;
        public string? Name { get; set; }
        public AuthStatusEnvironmentType? Type { get; set; }
    }
}
