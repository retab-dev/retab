namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using Newtonsoft.Json.Linq;

    /// <summary>
    /// JSON converter that deserializes discriminated union variants
    /// based on the "kind" property.
    /// </summary>
    public class ExperimentSummaryMetricsResponseDiscriminatorConverter : Newtonsoft.Json.JsonConverter
    {
        public override bool CanConvert(Type objectType) => objectType == typeof(object);

        public override object? ReadJson(Newtonsoft.Json.JsonReader reader, Type objectType, object? existingValue, Newtonsoft.Json.JsonSerializer serializer)
        {
            var jObject = JObject.Load(reader);
            var discriminatorValue = jObject["kind"]?.ToString();
            switch (discriminatorValue)
            {
                case "summary": return jObject.ToObject<ExperimentSummaryMetricsResponse>(serializer);
                case "by_document": return jObject.ToObject<ExperimentByDocumentMetricsResponse>(serializer);
                case "by_target": return jObject.ToObject<ExperimentByTargetMetricsResponse>(serializer);
                case "votes": return jObject.ToObject<ExperimentVotesMetricsResponse>(serializer);
                case "stale_metrics": return jObject.ToObject<ExperimentMetricsStaleError>(serializer);
                case "no_metrics": return jObject.ToObject<ExperimentMetricsMissingError>(serializer);
                default: return jObject.ToObject<object>(serializer);
            }
        }

        public override void WriteJson(Newtonsoft.Json.JsonWriter writer, object? value, Newtonsoft.Json.JsonSerializer serializer)
        {
            serializer.Serialize(writer, value);
        }
    }
}
