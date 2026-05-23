namespace Retab
{

    /// <summary>Represents a category.</summary>
    public class Category
    {

        /// <summary>The name of the category</summary>
        public string Name { get; set; } = default!;

        /// <summary>Stable machine key used by workflow classifier output handles</summary>
        public string? HandleKey { get; set; }

        /// <summary>The description of the category</summary>
        public string? Description { get; set; }
    }
}
