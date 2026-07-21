namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using Newtonsoft.Json.Linq;

    /// <summary>
    /// JSON converter that deserializes discriminated union variants
    /// based on the "kind" property.
    /// </summary>
    public class ReviewAlwaysDiscriminatorConverter : Newtonsoft.Json.JsonConverter
    {
        public override bool CanConvert(Type objectType) => objectType == typeof(object);

        public override object? ReadJson(Newtonsoft.Json.JsonReader reader, Type objectType, object? existingValue, Newtonsoft.Json.JsonSerializer serializer)
        {
            var jObject = JObject.Load(reader);
            var discriminatorValue = jObject["kind"]?.ToString();
            switch (discriminatorValue)
            {
                case "all_of": return jObject.ToObject<ReviewAllOf>(serializer);
                case "always": return jObject.ToObject<ReviewAlways>(serializer);
                case "any_of": return jObject.ToObject<ReviewAnyOf>(serializer);
                case "any_required_field_null": return jObject.ToObject<ReviewAnyRequiredFieldNull>(serializer);
                case "any_split_pages_lt": return jObject.ToObject<ReviewAnySplitPagesLt>(serializer);
                case "boundary_confidence_lt": return jObject.ToObject<ReviewBoundaryConfidenceLt>(serializer);
                case "branch_in": return jObject.ToObject<ReviewBranchIn>(serializer);
                case "category_in": return jObject.ToObject<ReviewCategoryIn>(serializer);
                case "confidence_lt": return jObject.ToObject<ReviewConfidenceLt>(serializer);
                case "json_condition": return jObject.ToObject<ReviewJsonCondition>(serializer);
                case "split_count_neq": return jObject.ToObject<ReviewSplitCountNeq>(serializer);
                case "top_margin_lt": return jObject.ToObject<ReviewTopMarginLt>(serializer);
                case "validation_failed": return jObject.ToObject<ReviewValidationFailed>(serializer);
                default: return jObject.ToObject<object>(serializer);
            }
        }

        public override void WriteJson(Newtonsoft.Json.JsonWriter writer, object? value, Newtonsoft.Json.JsonSerializer serializer)
        {
            serializer.Serialize(writer, value);
        }
    }
}
