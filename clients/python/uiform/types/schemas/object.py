from pydantic import BaseModel, Field, computed_field, model_validator, PrivateAttr
from typing import Any, Literal, cast, Self
import json
from pathlib import Path
import copy

from ..documents.create_messages import ChatCompletionUiformMessage
from ..documents.create_messages import convert_to_google_genai_format, convert_to_anthropic_format
from ..._utils.mime import generate_sha_hash_from_string
from ..._utils.json_schema import clean_schema, json_schema_to_inference_schema, json_schema_to_typescript_interface, expand_refs, create_reasoning_schema, schema_to_ts_type, convert_json_schema_to_basemodel, convert_basemodel_to_partial_basemodel, load_json_schema

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

    @property
    def inference_pydantic_model(self) -> type[BaseModel]:
        """Converts the structured output schema to a Pydantic model, with the LLMDescription and ReasoningDescription fields added.
        
        Returns:
            type[BaseModel]: A Pydantic model class generated from the schema.
        """
        return convert_json_schema_to_basemodel(self.inference_json_schema)

    @property
    def inference_json_schema(self) ->  dict[str, Any]:
        """Returns the schema formatted for structured output, with the LLMDescription and ReasoningDescription fields added.
        
        Returns:
            dict[str, Any]: The schema formatted for structured output processing.
        """
        inference_json_schema_ = json_schema_to_inference_schema(copy.deepcopy(self._reasoning_object_schema))
        assert isinstance(inference_json_schema_, dict), "Validation Error: The inference_json_schema is not a dict"
        return inference_json_schema_
    
    # This is a computed field, it is exposed when serializing the object
    @computed_field   # type: ignore
    @property
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
    @computed_field   # type: ignore
    @property
    def schema_version(self) -> str:
        """Returns the SHA1 hash of the complete schema.
        
        Returns:
            str: A SHA1 hash string representing the complete schema version.
        """
        return generate_sha_hash_from_string(json.dumps(self.json_schema, sort_keys=True).strip(), "sha1")
    
    @property
    def openai_messages(self) -> list[ChatCompletionMessageParam]:
        """Returns the messages formatted for OpenAI's API.
        
        Returns:
            list[ChatCompletionMessageParam]: List of messages in OpenAI's format.
        """
        return cast(list[ChatCompletionMessageParam], self.messages)
            

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
    def gemini_messages(self) -> content_types.ContentsType:
        """Returns the messages formatted for Google's Gemini API.
        
        Returns:
            content_types.ContentsType: Messages formatted for Gemini.
        """
        return convert_to_google_genai_format(self.messages)



    @property
    def inference_gemini_json_schema(self) -> dict[str, Any]:
        # Like OpenAI but does not accept "anyOf" typing, all fields must not be nullable
        inference_json_schema_ = copy.deepcopy(self.inference_json_schema)
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
        remove_optional_types_rec(inference_json_schema_)
        return inference_json_schema_


    @property
    def inference_typescript_interface(self) -> str:
        """Returns the TypeScript interface representation of the inference schema, that is more readable than the JSON schema.
        
        Returns:
            str: A string containing the TypeScript interface definition.
        """
        return json_schema_to_typescript_interface(self._reasoning_object_schema, add_field_description=True)

    @property
    def system_prompt(self) -> str:
        """Returns the system prompt combining custom prompt and TypeScript interface.
        
        Returns:
            str: The combined system prompt string.
        """
        return self.json_schema.get("X-SystemPrompt", "") + "\nThis is the expected output schema (as a TypeScript interface for better readability) with useful prompts added as comments bellow each field :\n\n" + self.inference_typescript_interface
    
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
            dict[str, Any]: The expanded schema with resolved references.
        """
        return expand_refs(copy.deepcopy(self.json_schema))

    @property
    def _reasoning_object_schema(self) -> dict[str, Any]:
        """Returns the schema with inference-specific modifications.
        
        Returns:
            dict[str, Any]: The modified schema with reasoning fields added to the structure.
        """
        inference_schema = create_reasoning_schema(copy.deepcopy(self._expanded_object_schema)) # Automatically populates the reasoning fields into the structure.
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


    
    
    
    def _get_pattern_attribute(self, pattern: str, attribute: Literal['X-FieldPrompt', 'X-ReasoningPrompt', 'type']) -> str | None:
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


    def _set_pattern_attribute(self, pattern: str, attribute: Literal['X-FieldPrompt', 'X-ReasoningPrompt', 'X-SystemPrompt'], value: str) -> None:
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
                ref_count = json.dumps(self.json_schema).count(ref)

                if ref_count > 1:
                    new_copy_names = [f"{ref_name}Copy{i+1}" for i in range(ref_count)]

                    # Get the nex copy name available
                    next_copy_name = next((name for name in new_copy_names if name not in definitions), None)
                    assert next_copy_name is not None, "Validation Error: No available copy name found"

                    # Create a copy of the definition
                    def_copy = copy.deepcopy(definitions[ref_name])
                    
                    # Change the title and name of the definition to avoid recursion
                    if "title" in def_copy:
                        def_copy["title"] = new_copy_names
                    if "name" in def_copy:
                        def_copy["name"] = new_copy_names

                    # Add the new copy name to the definitions
                    definitions[next_copy_name] = def_copy

                    # Replace the $ref with the new copy name
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
        json_schema: dict[str, Any] | None = data.get('json_schema', None)
        pydantic_model: type[BaseModel] | None = data.get('pydantic_model', None)

        # Check if either json_schema or pydantic_model is provided
        if json_schema and pydantic_model:
            raise ValueError("Cannot provide both json_schema and pydantic_model")
        
        if not json_schema and not pydantic_model:
            raise ValueError("Must provide either json_schema or pydantic_model")

        if json_schema:
            json_schema = load_json_schema(json_schema)
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
    

    def save(self, path: Path | str) -> None:
        """Save a JSON schema to a file.

        Args:
            json_schema: The JSON schema to save, can be a dict, Path, or string
            schema_path: Output path for the schema file
        """
        with open(path, 'w', encoding='utf-8') as f:
            json.dump(self.json_schema, f, ensure_ascii=False, indent=2)

