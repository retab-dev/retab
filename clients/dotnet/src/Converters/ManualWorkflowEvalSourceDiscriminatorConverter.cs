namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using Newtonsoft.Json.Linq;

    /// <summary>
    /// JSON converter that deserializes discriminated union variants
    /// based on the "type" property.
    /// </summary>
    public class ManualWorkflowEvalSourceDiscriminatorConverter : Newtonsoft.Json.JsonConverter
    {
        public override bool CanConvert(Type objectType) => objectType == typeof(object);

        public override object? ReadJson(Newtonsoft.Json.JsonReader reader, Type objectType, object? existingValue, Newtonsoft.Json.JsonSerializer serializer)
        {
            var jObject = JObject.Load(reader);
            var discriminatorValue = jObject["type"]?.ToString();
            switch (discriminatorValue)
            {
                case "manual": return jObject.ToObject<ManualWorkflowEvalSource>(serializer);
                case "run_step": return jObject.ToObject<RunStepWorkflowEvalSource>(serializer);
                default: return jObject.ToObject<object>(serializer);
            }
        }

        public override void WriteJson(Newtonsoft.Json.JsonWriter writer, object? value, Newtonsoft.Json.JsonSerializer serializer)
        {
            serializer.Serialize(writer, value);
        }
    }
}
