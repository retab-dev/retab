from pydantic import BaseModel, Field, computed_field, model_validator, PrivateAttr
from typing import Any, Literal, cast, Self
import json
import copy

from ..documents.create_messages import ChatCompletionUiformMessage
from ..documents.create_messages import convert_to_google_genai_format, convert_to_anthropic_format
from ..._utils.mime import generate_sha_hash_from_string
from ..._utils.json_schema import clean_schema, json_schema_to_structured_output_json_schema, json_schema_to_typescript_interface, expand_refs, create_inference_schema, schema_to_ts_type, convert_json_schema_to_basemodel, convert_basemodel_to_partial_basemodel

from openai.types.chat.chat_completion_message_param import ChatCompletionMessageParam
from openai.types.chat.completion_create_params import ResponseFormat

from anthropic.types.message_param import MessageParam
from anthropic._types import NotGiven

from google.generativeai.types import content_types # type: ignore



class Schema(BaseModel):

    id: str | None = None
    """A unique identifier for the document loading."""

    object: Literal["schema"] = "schema"
    """The type of object being preprocessed."""

    messages: list[ChatCompletionUiformMessage] = Field(default=[], exclude=True, repr=False)
    """A list of messages containing the system prompt and a user prompt."""

    created: int | None = None
    """The Unix timestamp (in seconds) of when the document was loaded."""

    json_schema: dict[str, Any] = {}
    """The JSON schema to use for loading."""

    pydantic_model: type[BaseModel] = Field(default=None, exclude=True, repr=False)     # type: ignore

    _partial_pydantic_model: type[BaseModel] = PrivateAttr()
    """The Pydantic model to use for loading."""

    @model_validator(mode="before")
    def validate_schema_and_model(cls, data: Any) -> Any:
        """Validate schema and model logic."""
        # Extract from data
        json_schema: dict[str, Any] | None = data.get('json_schema', None)
        pydantic_model: type[BaseModel] | None = data.get('pydantic_model', None)

        # Check if either json_schema or pydantic_model is provided
        if json_schema and pydantic_model:
            raise ValueError("Cannot provide both json_schema and pydantic_model")
        
        if not json_schema and not pydantic_model:
            raise ValueError("Must provide either json_schema or pydantic_model")

        if json_schema:
            data['pydantic_model'] = convert_json_schema_to_basemodel(json_schema)
            data['json_schema'] = json_schema
        if pydantic_model:
            data['pydantic_model'] = pydantic_model
            data['json_schema'] = pydantic_model.model_json_schema()


        return data

    @model_validator(mode="after")
    def model_after_validator(self) -> Self:
        # Validate Messages
        messages = getattr(self, "messages", [])
        self.messages = [ChatCompletionUiformMessage(role="system", content=self.system_prompt)] + messages

        # Set the partial_pydantic_model
        self._partial_pydantic_model = convert_basemodel_to_partial_basemodel(self.pydantic_model)

        return self


    @property
    def response_format_json(self) -> ResponseFormat:
        """Returns the JSON schema response format for OpenAI API, with the LLMDescription and ReasoningDescription fields added.
        
        Returns:
            ResponseFormat: A dictionary containing the JSON schema format specification.
        """
        return {
            "type": "json_schema",
            "json_schema": {
                "name": "document_preprocessing",
                "schema": self.structured_output_object_schema,
                "strict": True
            }
        }
    
    @property
    def response_format_pydantic(self) -> type[BaseModel]:
        """Converts the structured output schema to a Pydantic model, with the LLMDescription and ReasoningDescription fields added.
        
        Returns:
            type[BaseModel]: A Pydantic model class generated from the schema.
        """
        return convert_json_schema_to_basemodel(self.structured_output_object_schema)


    @property
    def response_format_json_gemini(self) -> ResponseFormat:
        # This will method does not allow nullable fields, every field is required and the anyOf is not supported.
        return {
            "type": "json_schema",
            "json_schema": {
                "name": "document_preprocessing",
                "schema": self.strict_gemini_object_schema,
                "strict": True
            }
        }

    # This is a computed field, it is exposed when serializing the object
    @property
    @computed_field
    def schema_data_version(self) -> str:
        """Returns the SHA1 hash of the schema data, ignoring all prompt/description/default fields.
        
        Returns:
            str: A SHA1 hash string representing the schema data version.
        """
        return generate_sha_hash_from_string(
            json.dumps(
                clean_schema(copy.deepcopy(self.json_schema), remove_custom_fields=True, fields_to_remove=["description", "default", "title", "required", "examples", "deprecated", "readOnly", "writeOnly"]),
                sort_keys=True).strip(), 
            "sha1")

    # This is a computed field, it is exposed when serializing the object
    @property
    @computed_field   
    def schema_version(self) -> str:
        """Returns the SHA1 hash of the complete schema.
        
        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        return generate_sha_hash_from_string(json.dumps(self.json_schema, sort_keys=True).strip(), "sha1")

    # Inner Properties not exposed when serializing
    @property
    def definitions(self) -> dict[str, Any]:
        """Returns the schema definitions ($defs) section.
        
        Returns:
            dict[str, Any]: A dictionary containing schema definitions.
        """
        if "$defs" in self.json_schema:
            return copy.deepcopy(self.json_schema["$defs"])
        return {}

    @property
    def expanded_object_schema(self) -> dict[str, Any]:
        """Returns the schema with all references expanded inline.
        
        Returns:
            dict[str, Any]: The expanded schema with resolved references.
        """
        return expand_refs(copy.deepcopy(self.json_schema))

    @property
    def inference_object_schema(self) -> dict[str, Any]:
        """Returns the schema with inference-specific modifications.
        
        Returns:
            dict[str, Any]: The modified schema with reasoning fields added to the structure.
        """
        inference_schema = create_inference_schema(copy.deepcopy(self.expanded_object_schema)) # Automatically populates the reasoning fields into the structure.
        assert isinstance(inference_schema, dict), "Validation Error: The structured_output_object_schema is not a dict"
        return inference_schema

    @property
    def structured_output_object_schema(self) ->  dict[str, Any]:
        """Returns the schema formatted for structured output, with the LLMDescription and ReasoningDescription fields added.
        
        Returns:
            dict[str, Any]: The schema formatted for structured output processing.
        """
        structured_output_object_schema_ = json_schema_to_structured_output_json_schema(copy.deepcopy(self.inference_object_schema))
        assert isinstance(structured_output_object_schema_, dict), "Validation Error: The structured_output_object_schema is not a dict"
        return structured_output_object_schema_
    
    @property
    def validation_object_schema(self) -> dict[str, Any]:
        """Returns a loose validation schema where all fields are optional.
        
        This schema ignores all 'required' properties, allowing partial data validation.
        
        Returns:
            dict[str, Any]: The modified schema for validation purposes.
        """
        # This ignores all 'required' properties (hence making all fields optional)
        # This is a 'loose' validation schema that allows for partial data to be validated.
        validation_object_schema_ = copy.deepcopy(self.inference_object_schema)
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
        
        rec_remove_required(validation_object_schema_)
        return validation_object_schema_


    @property
    def strict_gemini_object_schema(self) -> dict[str, Any]:
        # Like OpenAI but does not accept "anyOf" typing, all fields must not be nullable
        structured_output_object_schema_ = copy.deepcopy(self.structured_output_object_schema)
        def remove_optional_types_rec(schema: dict[str, Any]) -> None:
            if "properties" in schema:
                for prop_schema in schema["properties"].values():
                    remove_optional_types_rec(prop_schema)
            if "items" in schema:
                remove_optional_types_rec(schema["items"])
            if "$defs" in schema:
                for def_schema in schema["$defs"].values():
                    remove_optional_types_rec(def_schema)
            if "anyOf" in schema:
                any_of = schema.pop("anyOf")
                # Get the non-nullable type
                any_of_types = [s["type"] for s in any_of if s["type"] != "null"]
                # Set the non-nullable type
                if "type" not in schema and len(any_of_types) > 0:
                    schema["type"] = any_of_types[0]
            if "allOf" in schema:
                for allof_schema in schema["allOf"]:
                    remove_optional_types_rec(allof_schema)
        remove_optional_types_rec(structured_output_object_schema_)
        return structured_output_object_schema_


    @property
    def inference_typescript_interface(self) -> str:
        """Returns the TypeScript interface representation of the inference schema, that is more readable than the JSON schema.
        
        Returns:
            str: A string containing the TypeScript interface definition.
        """
        return json_schema_to_typescript_interface(self.inference_object_schema, add_field_description=True)

    @property
    def system_prompt(self) -> str:
        """Returns the system prompt combining custom prompt and TypeScript interface.
        
        Returns:
            str: The combined system prompt string.
        """
        return self.json_schema.get("X-SystemPrompt", "") + "\n" + self.inference_typescript_interface
    
    @property
    def title(self) -> str:
        """Returns the title of the schema.
        
        Returns:
            str: The schema title or 'NoTitle' if not specified.
        """
        return self.json_schema.get("title", "NoTitle")

    
    
    
    def get_pattern_attribute(self, pattern: str, attribute: Literal['description', 'X-FieldPrompt', 'X-ReasoningPrompt', 'type']) -> str | None:
        """
        Given a JSON Schema and a pattern (like "my_object.my_array.*.my_property"),
        navigate the schema and return the specified attribute of the identified node.
        """

        # Special case: "*" means the root schema itself
        current_schema = self.expanded_object_schema
        if pattern.strip() == "*":
            if attribute == "X-FieldPrompt":
                return current_schema.get(attribute) or current_schema.get("description")
            return current_schema.get(attribute)

        parts = pattern.split(".")
        index = 0  # Start at the first part

        while index < len(parts):
            part = parts[index]

            if part == "*":
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


    def set_pattern_attribute(self, pattern: str, attribute: Literal['description', 'X-FieldPrompt', 'X-ReasoningPrompt', 'X-SystemPrompt'], value: str) -> None:
        """Sets an attribute value at a specific path in the schema.
        
        Args:
            pattern (str): The path pattern to navigate the schema (e.g., "my_object.my_array.*.my_property")
            attribute (Literal): The attribute to set ('description', 'X-FieldPrompt', etc.)
            value (str): The value to set for the attribute
        """
        current_schema = self.json_schema
        definitions = self.json_schema["$defs"]
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
            if part == "*":
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
                if not ref.startswith("#/$defs/"):
                    # The reference is not to a known definition location
                    return
                ref_name = ref.split("/")[-1]
                if ref_name not in definitions:
                    # The reference is not found in the definitions
                    return 

                # Count how many times this ref is used in the entire schema
                ref_count = json.dumps(self.json_schema).count(ref)

                if ref_count > 1:
                    # Expand inline if the reference is used multiple times
                    def_copy = copy.deepcopy(definitions[ref_name])
                    current_schema.pop("$ref", None)  # Remove the $ref

                    # Merge def_copy into current_schema
                    for k, v in def_copy.items():
                        current_schema[k] = v

                    # Do not increment index; retry handling the current part
                else:
                    # Reference is used only once; directly navigate to the definition
                    current_schema = definitions[ref_name]
                    # Do not increment index; retry handling the current part
            else:
                # Cannot navigate further; invalid pattern
                return

        # Once we have navigated to the correct node, set the attribute
        current_schema[attribute] = value

    
    @property
    def openai_messages(self) -> list[ChatCompletionMessageParam]:
        """Returns the messages formatted for OpenAI's API.
        
        Returns:
            list[ChatCompletionMessageParam]: List of messages in OpenAI's format.
        """
        return cast(list[ChatCompletionMessageParam], self.messages)
            

    @property
    def anthropic_system_prompt(self) -> str | NotGiven:
        """Returns the system message in Anthropic's Claude format.
        
        Returns:
            str | NotGiven: The system prompt formatted for Claude or NotGiven if none exists.
        """
        return self.system_prompt

    @property
    def anthropic_messages(self) -> list[MessageParam]:
        """Returns the messages in Anthropic's Claude format.
        
        Returns:
            list[MessageParam]: List of messages formatted for Claude.
        """
        return convert_to_anthropic_format(self.messages)[1]
    
    @property
    def gemini_messages(self) -> content_types.ContentsType:
        """Returns the messages formatted for Google's Gemini API.
        
        Returns:
            content_types.ContentsType: Messages formatted for Gemini.
        """
        return convert_to_google_genai_format(self.messages)
