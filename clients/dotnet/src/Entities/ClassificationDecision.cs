namespace Retab
{

    /// <summary>Represents a classification decision.</summary>
    public class ClassificationDecision
    {

        /// <summary>The reasoning for the classification decision</summary>
        public string Reasoning { get; set; } = default!;

        /// <summary>The category name that the document belongs to</summary>
        public string Category { get; set; } = default!;
    }
}
