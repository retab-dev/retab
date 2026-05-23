namespace Retab
{

    /// <summary>Represents a b box.</summary>
    public class BBox
    {

        /// <summary>Left coordinate of the bounding box, relative to page width (0.0 = left edge, 1.0 = right edge).</summary>
        public double Left { get; set; }

        /// <summary>Top coordinate of the bounding box, relative to page height (0.0 = top edge, 1.0 = bottom edge).</summary>
        public double Top { get; set; }

        /// <summary>Width of the bounding box, relative to page width (0.0–1.0).</summary>
        public double Width { get; set; }

        /// <summary>Height of the bounding box, relative to page height (0.0–1.0).</summary>
        public double Height { get; set; }

        /// <summary>1-based index of the page where this field appears.</summary>
        public long Page { get; set; }
    }
}
