namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using Newtonsoft.Json.Linq;

    /// <summary>
    /// JSON converter that deserializes discriminated union variants
    /// based on the "status" property.
    /// </summary>
    public class PendingWorkflowExperimentRunDiscriminatorConverter : Newtonsoft.Json.JsonConverter
    {
        public override bool CanConvert(Type objectType) => objectType == typeof(object);

        public override object? ReadJson(Newtonsoft.Json.JsonReader reader, Type objectType, object? existingValue, Newtonsoft.Json.JsonSerializer serializer)
        {
            var jObject = JObject.Load(reader);
            var discriminatorValue = jObject["status"]?.ToString();
            switch (discriminatorValue)
            {
                case "cancelled": return jObject.ToObject<CancelledTerminal>(serializer);
                case "completed": return jObject.ToObject<CompletedBlockExecutionLifecycle>(serializer);
                case "error": return jObject.ToObject<ErrorWorkflowEvalRun>(serializer);
                case "pending": return jObject.ToObject<PendingBlockExecutionLifecycle>(serializer);
                case "queued": return jObject.ToObject<QueuedBlockExecutionLifecycle>(serializer);
                case "running": return jObject.ToObject<RunningBlockExecutionLifecycle>(serializer);
                default: return jObject.ToObject<object>(serializer);
            }
        }

        public override void WriteJson(Newtonsoft.Json.JsonWriter writer, object? value, Newtonsoft.Json.JsonSerializer serializer)
        {
            serializer.Serialize(writer, value);
        }
    }
}
