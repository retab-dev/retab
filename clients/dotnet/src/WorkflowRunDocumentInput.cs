// @oagen-ignore-file

using System;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Retab
{
    /// <summary>Document input for a workflow start-document block.</summary>
    [JsonConverter(typeof(WorkflowRunDocumentInputJsonConverter))]
    public sealed class WorkflowRunDocumentInput
    {
        internal object Value { get; }

        private WorkflowRunDocumentInput(object value)
        {
            this.Value = value ?? throw new ArgumentNullException(nameof(value));
        }

        public static WorkflowRunDocumentInput FromMimeData(MimeData document)
            => new WorkflowRunDocumentInput(document);

        public static WorkflowRunDocumentInput FromFileRef(FileRef document)
            => new WorkflowRunDocumentInput(document);

        public static implicit operator WorkflowRunDocumentInput(MimeData document)
            => FromMimeData(document);

        public static implicit operator WorkflowRunDocumentInput(FileRef document)
            => FromFileRef(document);
    }

    public sealed class WorkflowRunDocumentInputJsonConverter : JsonConverter<WorkflowRunDocumentInput>
    {
        public override WorkflowRunDocumentInput? Read(ref Utf8JsonReader reader, Type typeToConvert, JsonSerializerOptions options)
        {
            using var document = JsonDocument.ParseValue(ref reader);
            var json = document.RootElement.GetRawText();
            if (document.RootElement.TryGetProperty("id", out _))
            {
                var fileRef = JsonSerializer.Deserialize<FileRef>(json, options);
                return fileRef == null ? null : WorkflowRunDocumentInput.FromFileRef(fileRef);
            }

            var mimeData = JsonSerializer.Deserialize<MimeData>(json, options);
            return mimeData == null ? null : WorkflowRunDocumentInput.FromMimeData(mimeData);
        }

        public override void Write(Utf8JsonWriter writer, WorkflowRunDocumentInput value, JsonSerializerOptions options)
        {
            JsonSerializer.Serialize(writer, value.Value, value.Value.GetType(), options);
        }
    }
}
