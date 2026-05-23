namespace Retab
{

    /// <summary>Public/shared file reference used across SDK and customer-facing APIs.</summary>
    public class FileRef
    {

        /// <summary>ID of the file</summary>
        public string Id { get; set; } = default!;

        /// <summary>Filename of the file</summary>
        public string Filename { get; set; } = default!;

        /// <summary>MIME type of the file</summary>
        public string MimeType { get; set; } = default!;
    }
}
