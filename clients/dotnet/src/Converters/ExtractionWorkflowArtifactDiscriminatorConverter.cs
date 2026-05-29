namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using Newtonsoft.Json.Linq;

    /// <summary>
    /// JSON converter that deserializes discriminated union variants
    /// based on the "operation" property.
    /// </summary>
    public class ExtractionWorkflowArtifactDiscriminatorConverter : Newtonsoft.Json.JsonConverter
    {
        public override bool CanConvert(Type objectType) => objectType == typeof(object);

        public override object? ReadJson(Newtonsoft.Json.JsonReader reader, Type objectType, object? existingValue, Newtonsoft.Json.JsonSerializer serializer)
        {
            var jObject = JObject.Load(reader);
            var discriminatorValue = jObject["operation"]?.ToString();
            switch (discriminatorValue)
            {
                case "extraction": return jObject.ToObject<ExtractionWorkflowArtifact>(serializer);
                case "split": return jObject.ToObject<SplitWorkflowArtifact>(serializer);
                case "classification": return jObject.ToObject<ClassificationWorkflowArtifact>(serializer);
                case "parse": return jObject.ToObject<ParseWorkflowArtifact>(serializer);
                case "edit": return jObject.ToObject<EditWorkflowArtifact>(serializer);
                case "partition": return jObject.ToObject<PartitionWorkflowArtifact>(serializer);
                case "conditional_evaluation": return jObject.ToObject<ConditionalEvaluation>(serializer);
                case "review_trigger_evaluation": return jObject.ToObject<ReviewEvaluation>(serializer);
                case "while_loop_termination": return jObject.ToObject<WhileLoopTermination>(serializer);
                case "api_call_invocation": return jObject.ToObject<ApiCallInvocation>(serializer);
                case "function_invocation": return jObject.ToObject<FunctionInvocation>(serializer);
                default: return jObject.ToObject<object>(serializer);
            }
        }

        public override void WriteJson(Newtonsoft.Json.JsonWriter writer, object? value, Newtonsoft.Json.JsonSerializer serializer)
        {
            serializer.Serialize(writer, value);
        }
    }
}
