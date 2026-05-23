namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using Newtonsoft.Json.Linq;

    /// <summary>
    /// JSON converter that deserializes discriminated union variants
    /// based on the "kind" property.
    /// </summary>
    public class ExistConditionDiscriminatorConverter : Newtonsoft.Json.JsonConverter
    {
        public override bool CanConvert(Type objectType) => objectType == typeof(object);

        public override object? ReadJson(Newtonsoft.Json.JsonReader reader, Type objectType, object? existingValue, Newtonsoft.Json.JsonSerializer serializer)
        {
            var jObject = JObject.Load(reader);
            var discriminatorValue = jObject["kind"]?.ToString();
            switch (discriminatorValue)
            {
                case "all_items_match": return jObject.ToObject<AllItemsMatchCondition>(serializer);
                case "any_item_matches": return jObject.ToObject<AnyItemMatchesCondition>(serializer);
                case "array_contains": return jObject.ToObject<ArrayContainsCondition>(serializer);
                case "between": return jObject.ToObject<BetweenCondition>(serializer);
                case "contains": return jObject.ToObject<ContainCondition>(serializer);
                case "ends_with": return jObject.ToObject<EndsWithCondition>(serializer);
                case "equals": return jObject.ToObject<EqualCondition>(serializer);
                case "exists": return jObject.ToObject<ExistCondition>(serializer);
                case "json_schema_valid": return jObject.ToObject<JsonSchemaValidCondition>(serializer);
                case "length_compare": return jObject.ToObject<LengthCompareCondition>(serializer);
                case "llm_judged_as": return jObject.ToObject<LlmJudgedAsCondition>(serializer);
                case "llm_not_judged_as": return jObject.ToObject<LlmNotJudgedAsCondition>(serializer);
                case "matches_regex": return jObject.ToObject<MatcheRegexCondition>(serializer);
                case "not_contains": return jObject.ToObject<NotContainsCondition>(serializer);
                case "not_equals": return jObject.ToObject<NotEqualsCondition>(serializer);
                case "not_exists": return jObject.ToObject<NotExistsCondition>(serializer);
                case "number_compare": return jObject.ToObject<NumberCompareCondition>(serializer);
                case "object_contains": return jObject.ToObject<ObjectContainsCondition>(serializer);
                case "similarity_gte": return jObject.ToObject<SimilarityGteCondition>(serializer);
                case "split_iou_gte": return jObject.ToObject<SplitIouCondition>(serializer);
                case "starts_with": return jObject.ToObject<StartWithCondition>(serializer);
                default: return jObject.ToObject<object>(serializer);
            }
        }

        public override void WriteJson(Newtonsoft.Json.JsonWriter writer, object? value, Newtonsoft.Json.JsonSerializer serializer)
        {
            serializer.Serialize(writer, value);
        }
    }
}
