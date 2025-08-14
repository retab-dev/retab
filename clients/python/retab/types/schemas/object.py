import copy
import datetime
import json
from pathlib import Path
from typing import Any, Literal, Self

from anthropic.types.message_param import MessageParam
from google.genai.types import ContentUnionDict  # type: ignore
from openai.types.chat.chat_completion_message_param import ChatCompletionMessageParam
from openai.types.responses.response_input_param import ResponseInputItemParam
from pydantic import BaseModel, Field, PrivateAttr, computed_field, model_validator

from ...utils.chat import convert_to_anthropic_format, convert_to_google_genai_format
from ...utils.chat import convert_to_openai_format as convert_to_openai_completions_api_format
from ...utils.json_schema import (
    convert_basemodel_to_partial_basemodel,
    convert_json_schema_to_basemodel,
    create_reasoning_schema,
    expand_refs,
    generate_schema_data_id,
    generate_schema_id,
    json_schema_to_nlp_data_structure,
    json_schema_to_strict_openai_schema,
    json_schema_to_typescript_interface,
    load_json_schema,
    schema_to_ts_type,
)
from ...utils.responses import convert_to_openai_format as convert_to_openai_responses_api_format
from ...types.standards import StreamingBaseModel
from ..chat import ChatCompletionRetabMessage


class PartialSchema(BaseModel):
    """Response from the Generate Schema API -- A partial Schema object with no validation"""

    object: Literal["schema"] = "schema"
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    json_schema: dict[str, Any] = {}
    strict: bool = True


class PartialSchemaChunk(StreamingBaseModel):
    object: Literal["schema.chunk"] = "schema.chunk"
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    delta_json_schema_flat: dict[str, Any] = {}
    delta_flat_deleted_keys: list[str] = []


# class PartialSchemaStreaming(StreamingBaseModel, PartialSchema): pass


class Schema(PartialSchema):
    """A full Schema object with validation."""

    object: Literal["schema"] = "schema"
    """The type of object being preprocessed."""

    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    """The timestamp of when the schema was created."""

    json_schema: dict[str, Any] = {}
    """The JSON schema to use for loading."""

    # This is a computed field, it is exposed when serializing the object
    @computed_field  # type: ignore
    @property
    def data_id(self) -> str:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.

        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        return generate_schema_data_id(self.json_schema)

    # This is a computed field, it is exposed when serializing the object
    @computed_field  # type: ignore
    @property
    def id(self) -> str:
        """Returns the SHA1 hash of the complete schema.

        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        return generate_schema_id(self.json_schema)

    pydantic_model: type[BaseModel] = Field(default=None, exclude=True, repr=False)  # type: ignore

    _partial_pydantic_model: type[BaseModel] = PrivateAttr()
    """The Pydantic model to use for loading."""

    @property
    def inference_pydantic_model(self) -> type[BaseModel]:
        """Converts the structured output schema to a Pydantic model, with the LLMDescription and ReasoningDescription fields added.

        Returns:
            type[BaseModel]: A Pydantic model class generated from the schema.
        """
        return convert_json_schema_to_basemodel(self.inference_json_schema)

    @property
    def inference_json_schema(self) -> dict[str, Any]:
        """Returns the schema formatted for structured output, with the LLMDescription and ReasoningDescription fields added.

        Returns:
            dict[str, Any]: The schema formatted for structured output processing.
        """
        if self.strict:
            inference_json_schema_ = json_schema_to_strict_openai_schema(copy.deepcopy(self._reasoning_object_schema))
            assert isinstance(inference_json_schema_, dict), "Validation Error: The inference_json_schema is not a dict"
            return inference_json_schema_
        else:
            return copy.deepcopy(self._reasoning_object_schema)

    @property
    def openai_messages(self) -> list[ChatCompletionMessageParam]:
        """Returns the messages formatted for OpenAI's API.

        Returns:
            list[ChatCompletionMessageParam]: List of messages in OpenAI's format.
        """
        return convert_to_openai_completions_api_format(self.messages)

    @property
    def openai_responses_input(self) -> list[ResponseInputItemParam]:
        """Returns the messages formatted for OpenAI's Responses API.

        Returns:
            list[ResponseInputItemParam]: List of messages in OpenAI's Responses API format.
        """
        return convert_to_openai_responses_api_format(self.messages)

    @property
    def anthropic_system_prompt(self) -> str:
        """Returns the system message in Anthropic's Claude format.

        Returns:
            str : The system prompt formatted for Claude.
        """
        return "Return your response as a JSON object following the provided schema." + self.system_prompt

    @property
    def anthropic_messages(self) -> list[MessageParam]:
        """Returns the messages in Anthropic's Claude format.

        Returns:
            list[MessageParam]: List of messages formatted for Claude.
        """
        return convert_to_anthropic_format(self.messages)[1]

    @property
    def gemini_system_prompt(self) -> str:
        return convert_to_google_genai_format(self.messages)[0]

    @property
    def gemini_messages(self) -> list[ContentUnionDict]:
        """Returns the messages formatted for Google's Gemini API."""
        return convert_to_google_genai_format(self.messages)[1]

    @property
    def inference_gemini_json_schema(self) -> dict[str, Any]:
        # Like OpenAI but does not accept "anyOf" typing, all fields must not be nullable
        inference_json_schema_ = copy.deepcopy(self._reasoning_object_schema)

        def json_schema_to_gemini_schema(schema: dict[str, Any]) -> None:
            if "$defs" in schema:
                for def_schema in schema["$defs"].values():
                    json_schema_to_gemini_schema(def_schema)
            if "anyOf" in schema:
                any_of = schema.pop("anyOf")
                is_nullable = any(s.get("type") == "null" for s in any_of)
                # Get the non-null subschemas
                non_null_schemas = [s for s in any_of if s.get("type") != "null"]

                if non_null_schemas:
                    subschema = non_null_schemas[0]
                    json_schema_to_gemini_schema(subschema)
                    # Take the first non-null subschema and merge it into the parent schema
                    schema.update(subschema)
                else:
                    raise ValueError("No non-null subschemas found within anyOf")

                if is_nullable and schema.get("type") not in ["object", "array"]:
                    schema["nullable"] = True

            if "allOf" in schema:
                for allof_schema in schema["allOf"]:
                    json_schema_to_gemini_schema(allof_schema)

            if schema.get("type") == "object" and "properties" in schema:
                for prop_schema in schema["properties"].values():
                    json_schema_to_gemini_schema(prop_schema)
                schema["propertyOrdering"] = schema["required"] = list(schema["properties"].keys())

            if schema.get("type") == "array" and "items" in schema:
                json_schema_to_gemini_schema(schema["items"])
            # Remove not allowed fields
            for key in ["additionalProperties", "format"]:
                schema.pop(key, None)

        json_schema_to_gemini_schema(inference_json_schema_)
        return inference_json_schema_

    @property
    def inference_typescript_interface(self) -> str:
        """Returns the TypeScript interface representation of the inference schema, that is more readable than the JSON schema.

        Returns:
            str: A string containing the TypeScript interface definition.
        """
        return json_schema_to_typescript_interface(self._reasoning_object_schema, add_field_description=False)

    @property
    def inference_nlp_data_structure(self) -> str:
        """Returns the NLP data structure representation of the inference schema, that is more readable than the JSON schema.

        Returns:
            str: A string containing the NLP data structure definition.
        """
        from ...utils.json_schema import create_reasoning_schema_without_ref_expansion

        reasoning_schema = create_reasoning_schema_without_ref_expansion(self.json_schema)
        return json_schema_to_nlp_data_structure(reasoning_schema)

    @property
    def developer_system_prompt(self) -> str:
        return """
# SYSTEM — STRICT JSON-ONLY OUTPUT FOR STRUCTURED DATA EXTRACTION

## Role

You are an expert in data extraction and structured outputs.
You will be given:
	1.	A JSON Schema
	2.	A source document

Your task is to extract all relevant data from the document according to the schema and return exactly one valid JSON object that conforms to the schema and the rules below.

## Absolute Output Contract

	1.	Output exactly one JSON object and nothing else.
	2.	Output must be JSON only — no prose, no Markdown, no code fences, no backticks, no comments.
	3.	The JSON must be syntactically valid and UTF-8 encodable:
	•	Proper quoting/escaping
	•	No trailing commas
	•	No invalid characters
	4.	If the schema defines required fields, include them; use null for missing or ambiguous leaf values as per the rules below.

## Schema Compliance

	•	Follow the provided JSON Schema exactly.
	•	Do not add fields not defined by the schema; respect additionalProperties rules if present.
	•	Types must match the schema:
	•	string → JSON string
	•	number / integer → JSON number (not quoted)
	•	boolean → true / false
	•	object / array → exact structure per schema
	•	Only coerce to string when the schema type is string.
	•	Preserve field names exactly as in the schema.

## Date/Time Normalization


For string-typed date/time/datetime fields:
	•	Use ISO-8601:
	•	Date: YYYY-MM-DD
	•	Time: HH:MM:SS (24h)
	•	Datetime UTC: YYYY-MM-DDTHH:MM:SSZ
	•	Datetime with offset: YYYY-MM-DDTHH:MM:SS±HH:MM
	•	Do not invent time zones. If the source has no timezone, return a naive datetime (no Z or offset).
	•	Do not use non-ISO formats.

Correct Examples:

~~~
{"date": "2023-12-25"}
{"time": "14:30:00"}
{"datetime": "2023-12-25T14:30:00Z"}
{"datetime_with_tz": "2023-12-25T14:30:00+02:00"}
~~~

Incorrect Examples:

~~~
{"date": "12/25/2023"}
{"time": "2:30 PM"}
{"datetime": "Dec 25, 2023 at 2:30 PM"}
~~~


## Missing and Nullable Data


Nullable Leaf Attributes
	•	If valid data is missing or not explicitly present, set leaf attributes to null.
	•	Do NOT use empty strings (""), placeholders, or fabricated data.

Correct:

~~~
{"email": null}
~~~

Incorrect:

~~~
{"email": ""}
~~~

Nullable Nested Objects
	•	If an entire nested object's data is missing, do not set the object itself to null.
	•	Keep the object structure intact, with each leaf set to null.
	•	This preserves structure and communicates exactly which fields lack data.

Correct (all missing):

~~~
{
  "address": {
    "street": null,
    "zipCode": null,
    "city": null
  }
}
~~~

Incorrect (all missing):

~~~
{"address": null}
~~~

Correct (partially missing):

~~~
{
  "address": {
    "street": null,
    "zipCode": null,
    "city": "Paris"
  }
}
~~~

Incorrect (partially missing):

~~~
{"address": {"city": "Paris"}}
~~~

Arrays
	•	If no items are present, return an empty array [], not null.
	•	For object elements, preserve each element's full structure with missing leaves set to null.

## Reasoning Fields


Some schemas define reasoning fields for traceability and explanation, named with the prefix reasoning___:

Reasoning Field Type	Field Naming Pattern
Root Object	reasoning___root
Nested Objects	reasoning___[objectname]
Array Fields	reasoning___[arrayname]
Array Elements	reasoning___item
Leaf Attributes	reasoning___[attributename]

Rules for Reasoning Fields
	•	Must be plain JSON strings (no Markdown, no bullet lists).
	•	Only appear if present in the schema.
	•	Should include, where applicable:
	•	Explicit evidence: short, escaped quotes from the document.
	•	Decision justification: why you chose/rejected data.
	•	Calculations/normalizations done.
	•	Alternative interpretations considered/rejected.
	•	Confidence and assumptions.

## Detailed Reasoning Examples


Array Reasoning (reasoning___[arrayname]):

"Identified itemized invoice section headed 'Invoice Items' (page 2, lines 12-17). Extracted: 1) Office Supplies, qty 5, unit $4.99, total $24.95; 2) Printer Paper, qty 1, unit $5.99, total $5.99; 3) Stapler, qty 1, unit $4.07, total $4.07."

Array Item Reasoning (reasoning___item):

"From line 12: 'Office Supplies x5 $4.99ea $24.95'. Quantity (5) x unit price ($4.99) matches total ($24.95)."


## Principles for Accurate Extraction

	•	Transparency: Document and justify every extraction in reasoning fields when present.
	•	Precision: Quote directly from the source.
	•	Conservatism: Use null when data is missing or ambiguous — never guess.
	•	Structure Preservation: Match the schema exactly, maintain full structure.

## Normalization & Precision

	•	Trim leading/trailing whitespace.
	•	Preserve meaningful internal spaces and diacritics.
	•	For numbers: remove thousand separators, exclude units or currency symbols.
	•	If ambiguous, set to null and explain in reasoning.

## Disambiguation

	•	Prefer explicitly labeled data.
	•	If multiple conflicting values, choose the most clearly labeled and justify in reasoning.
	•	If unresolved, set to null and document in reasoning.

## Validation Checklist (before output)

	1.	Output is exactly one JSON object.
	2.	Field names match schema exactly (including reasoning___*).
	3.	Types match schema; numbers and booleans are not quoted.
	4.	All strings are UTF-8 encodable and escaped.
	5.	Dates/times follow ISO-8601.
	6.	Missing leaves are null; empty arrays are []; nested objects have null leaves.
	7.	No extra fields, trailing commas, prose, Markdown, or code fences.

## Fail-Safe

	•	Never break the JSON-only rule.
	•	If the document is unreadable or highly ambiguous:
	•	Still return a valid JSON object matching the schema.
	•	Fill known values.
	•	Set uncertain or missing leaves to null.
	•	Place any explanation only in reasoning fields if present.

⸻


"""

    @property
    def user_system_prompt(self) -> str | None:
        return self.json_schema.get("X-SystemPrompt", None)

    @property
    def schema_system_prompt(self) -> str:
        return (
            self.inference_nlp_data_structure + "\n---\n" + "## Expected output schema as a TypeScript interface for better readability:\n\n" + self.inference_typescript_interface
        )

    @property
    def system_prompt(self) -> str:
        """Returns the system prompt combining custom prompt and TypeScript interface.

        Returns:
            str: The combined system prompt string.
        """
        return self.developer_system_prompt + "\n\n" + (self.user_system_prompt + "\n\n" if self.user_system_prompt else "") + self.schema_system_prompt

    @property
    def title(self) -> str:
        """Returns the title of the schema.

        Returns:
            str: The schema title or 'NoTitle' if not specified.
        """
        return self.json_schema.get("title", "NoTitle")

    @property
    def _expanded_object_schema(self) -> dict[str, Any]:
        """Returns the schema with all references expanded inline.

        Returns:
            dict[str, Any]: The expanded schema with resolved references. If the schema is not expandable, it is returned as is.
        """
        return expand_refs(copy.deepcopy(self.json_schema))

    @property
    def _reasoning_object_schema(self) -> dict[str, Any]:
        """Returns the schema with inference-specific modifications.

        Returns:
            dict[str, Any]: The modified schema with reasoning fields added to the structure.
        """
        inference_schema = create_reasoning_schema(copy.deepcopy(self._expanded_object_schema))  # Automatically populates the reasoning fields into the structure.
        assert isinstance(inference_schema, dict), "Validation Error: The inference_json_schema is not a dict"
        return inference_schema

    @property
    def _validation_object_schema(self) -> dict[str, Any]:
        """Returns a loose validation schema where all fields are optional.

        This schema ignores all 'required' properties, allowing partial data validation.

        Returns:
            dict[str, Any]: The modified schema for validation purposes.
        """
        # This ignores all 'required' properties (hence making all fields optional)
        # This is a 'loose' validation schema that allows for partial data to be validated.
        _validation_object_schema_ = copy.deepcopy(self._reasoning_object_schema)

        def rec_remove_required(schema: dict[str, Any]) -> None:
            if "required" in schema:
                schema.pop("required")
            if "properties" in schema:
                for prop_schema in schema["properties"].values():
                    rec_remove_required(prop_schema)
            if "items" in schema:
                rec_remove_required(schema["items"])
            if "$defs" in schema:
                for def_schema in schema["$defs"].values():
                    rec_remove_required(def_schema)
            if "anyOf" in schema:
                for anyof_schema in schema["anyOf"]:
                    rec_remove_required(anyof_schema)
            if "allOf" in schema:
                for allof_schema in schema["allOf"]:
                    rec_remove_required(allof_schema)

        rec_remove_required(_validation_object_schema_)
        return _validation_object_schema_

    def _get_pattern_attribute(self, pattern: str, attribute: Literal["X-FieldPrompt", "X-ReasoningPrompt", "type"]) -> str | None:
        """
        Given a JSON Schema and a pattern (like "my_object.my_array.*.my_property"),
        navigate the schema and return the specified attribute of the identified node.
        """

        # Special case: "*" means the root schema itself
        current_schema = self._expanded_object_schema
        if pattern.strip() == "*":
            if attribute == "X-FieldPrompt":
                return current_schema.get(attribute) or current_schema.get("description")
            return current_schema.get(attribute)

        parts = pattern.split(".")
        index = 0  # Start at the first part

        while index < len(parts):
            part = parts[index]

            if part == "*" or part.isdigit():
                # Handle wildcard case for arrays
                if "items" in current_schema:
                    current_schema = current_schema["items"]
                    index += 1  # Move to the next part
                else:
                    # Invalid use of "*" for the current schema
                    return None
            elif "properties" in current_schema and part in current_schema["properties"]:
                # Handle normal property navigation
                current_schema = current_schema["properties"][part]
                index += 1  # Move to the next part
            else:
                # If we encounter a structure without "properties" or invalid part
                return None

        # At this point, we've navigated to the target node
        if attribute == "X-FieldPrompt":
            return current_schema.get(attribute) or current_schema.get("description")
        elif attribute == "type":
            # Convert schema type to TypeScript type
            return schema_to_ts_type(current_schema, {}, {}, 0, 0, add_field_description=False)
        return current_schema.get(attribute)

    def _set_pattern_attribute(self, pattern: str, attribute: Literal["X-FieldPrompt", "X-ReasoningPrompt", "X-SystemPrompt", "description"], value: str) -> None:
        """Sets an attribute value at a specific path in the schema.

        Args:
            pattern (str): The path pattern to navigate the schema (e.g., "my_object.my_array.*.my_property")
            attribute (Literal): The attribute to set ('description', 'X-FieldPrompt', etc.)
            value (str): The value to set for the attribute
        """
        current_schema = self.json_schema
        definitions = self.json_schema.get("$defs", {})
        parts = pattern.split(".")
        path_stack: list[tuple[str, Any]] = []  # Keep track of how we navigated the schema

        if pattern.strip() == "*":
            # Special case: "*" means the root schema itself
            current_schema[attribute] = value
            return
        assert attribute != "X-SystemPrompt", "Cannot set the X-SystemPrompt attribute other than at the root schema."

        index = 0  # Index for the parts list
        while index < len(parts):
            part = parts[index]
            if part == "*" or part.isdigit():
                # Handle the array case
                if "items" in current_schema:
                    current_schema = current_schema["items"]
                    path_stack.append(("items", None))
                    index += 1  # Move to the next part
                else:
                    return  # Invalid pattern for the current schema

            elif "properties" in current_schema and part in current_schema["properties"]:
                # Handle the properties case
                current_schema = current_schema["properties"][part]
                path_stack.append(("properties", part))
                index += 1  # Move to the next part
            elif "$ref" in current_schema:
                # Handle the $ref case
                ref = current_schema["$ref"]
                assert isinstance(ref, str), "Validation Error: The $ref is not a string"
                assert ref.startswith("#/$defs/"), "Validation Error: The $ref is not a definition reference"
                ref_name = ref.split("/")[-1]
                assert ref_name in definitions, "Validation Error: The $ref is not a definition reference"

                # Count how many times this ref is used in the entire schema
                ref_count = json.dumps(self.json_schema).count(f'"{ref}"')

                if ref_count > 1:
                    # Create a unique copy name by appending a number
                    copy_num = 1
                    next_copy_name = f"{ref_name}Copy{copy_num}"
                    while next_copy_name in definitions:
                        copy_num += 1
                        next_copy_name = f"{ref_name}Copy{copy_num}"

                    # Create a copy of the definition
                    def_copy = copy.deepcopy(definitions[ref_name])

                    # Change the title and name of the definition
                    if "title" in def_copy:
                        def_copy["title"] = f"{def_copy['title']} Copy {copy_num}"
                    if "name" in def_copy:
                        def_copy["name"] = next_copy_name

                    # Add the new copy to definitions
                    definitions[next_copy_name] = def_copy

                    # Update the reference
                    current_schema["$ref"] = f"#/$defs/{next_copy_name}"
                    ref_name = next_copy_name
                # Reference is used only once or a copy is created; directly navigate to the definition
                current_schema = definitions[ref_name]
            else:
                # Cannot navigate further; invalid pattern
                return

        # Once we have navigated to the correct node, set the attribute
        current_schema[attribute] = value

    @model_validator(mode="before")
    def validate_schema_and_model(cls, data: Any) -> Any:
        """Validate schema and model logic."""
        # Extract from data
        json_schema: dict[str, Any] | None = data.get("json_schema", None)
        pydantic_model: type[BaseModel] | None = data.get("pydantic_model", None)

        # Check if either json_schema or pydantic_model is provided
        if json_schema and pydantic_model:
            raise ValueError("Cannot provide both json_schema and pydantic_model")

        if not json_schema and not pydantic_model:
            raise ValueError("Must provide either json_schema or pydantic_model")

        if json_schema:
            json_schema = load_json_schema(json_schema)
            data["pydantic_model"] = convert_json_schema_to_basemodel(json_schema)
            data["json_schema"] = json_schema
        if pydantic_model:
            data["pydantic_model"] = pydantic_model
            data["json_schema"] = pydantic_model.model_json_schema()

        return data

    @property
    def messages(self) -> list[ChatCompletionRetabMessage]:
        return [ChatCompletionRetabMessage(role="developer", content=self.system_prompt)]

    @model_validator(mode="after")
    def model_after_validator(self) -> Self:
        # Set the partial_pydantic_model
        self._partial_pydantic_model = convert_basemodel_to_partial_basemodel(self.pydantic_model)

        return self

    def save(self, path: Path | str) -> None:
        """Save a JSON schema to a file.

        Args:
            json_schema: The JSON schema to save, can be a dict, Path, or string
            schema_path: Output path for the schema file
        """
        with open(path, "w", encoding="utf-8") as f:
            json.dump(self.json_schema, f, ensure_ascii=False, indent=2)
