import copy
import json
from typing import Any, Literal, Optional, Self, Union, Type, MutableMapping, Tuple, MutableSequence
from ...utils.hashing import generate_blake2b_hash_from_string

import datetime
from pathlib import Path


from anthropic.types.message_param import MessageParam
from google.genai.types import ContentUnionDict  # type: ignore
from openai.types.chat.chat_completion_message_param import ChatCompletionMessageParam
from openai.types.responses.response_input_param import ResponseInputItemParam
from pydantic import BaseModel, Field, PrivateAttr, computed_field, model_validator

from .chat import convert_to_anthropic_format, convert_to_google_genai_format
from .chat import convert_to_openai_completions_api_format

from ...utils.json_schema import convert_json_schema_to_basemodel, expand_refs, load_json_schema
from .chat import convert_to_openai_responses_api_format
from ..standards import StreamingBaseModel
from ..chat import ChatCompletionRetabMessage

def add_reasoning_sibling_inplace(properties: dict[str, Any], field_name: str, reasoning_desc: str) -> None:
    """
    Add a reasoning sibling for a given property field_name into properties dict.
    We'll use the naming convention reasoning___<field_name>.
    If the field_name is 'root', we add 'reasoning___root'.
    """
    reasoning_key = f"reasoning___{field_name}"
    new_properties: dict[str, Any]
    if field_name == "root":
        new_properties = {reasoning_key: {"type": "string", "description": reasoning_desc}, **properties}
    else:
        # Insert reasoning_key just above the field_name
        new_properties = {}
        for key, value in properties.items():
            if key == field_name:
                new_properties[reasoning_key] = {"type": "string", "description": reasoning_desc}
            new_properties[key] = value
    properties.clear()
    properties.update(new_properties)


def _insert_reasoning_fields_inner(schema: dict[str, Any]) -> tuple[dict[str, Any], str | None]:
    """
    Inner function that returns (updated_schema, reasoning_desc_for_this_node).
    The parent caller (which handles 'properties') will add the sibling reasoning field if reasoning_desc_for_this_node is not None.
    """
    reasoning_desc = schema.pop("X-ReasoningPrompt", None)

    node_type = schema.get("type")

    # Process children recursively
    # If object: process properties
    if node_type == "object" or "$ref" in schema:
        if "properties" in schema and isinstance(schema["properties"], dict):
            new_props = {}
            for property_key, property_value in schema["properties"].items():
                updated_prop_schema, child_reasoning = _insert_reasoning_fields_inner(property_value)
                new_props[property_key] = updated_prop_schema
                if child_reasoning:
                    add_reasoning_sibling_inplace(new_props, property_key, child_reasoning)
                    # Add the reasoning field to required if the property is required
                    if "required" in schema and property_key in schema["required"]:
                        schema["required"].append(f"reasoning___{property_key}")
            schema["properties"] = new_props

        if "$defs" in schema and isinstance(schema["$defs"], dict):
            new_defs = {}
            for dk, dv in schema["$defs"].items():
                updated_def_schema, _ = _insert_reasoning_fields_inner(dv)
                new_defs[dk] = updated_def_schema
            schema["$defs"] = new_defs

    elif node_type == "array" and "items" in schema:
        # Recurse into items if present
        updated_items, item_reasoning = _insert_reasoning_fields_inner(schema["items"])
        schema["items"] = updated_items

        # If the item schema has a reasoning prompt, create a reasoning field inside the item
        if item_reasoning and updated_items.get("type") == "object":
            # Create reasoning field for array items
            if "properties" not in updated_items:
                updated_items["properties"] = {}

            # Add the reasoning field as first property
            reasoning_key = "reasoning___item"
            new_properties = {reasoning_key: {"type": "string", "description": item_reasoning}}

            # Add the rest of the properties
            for key, value in updated_items["properties"].items():
                new_properties[key] = value

            updated_items["properties"] = new_properties

            # Add to required if we have required fields
            if "required" in updated_items:
                updated_items["required"].insert(0, reasoning_key)
            else:
                updated_items["required"] = [reasoning_key]

    return schema, reasoning_desc


def _insert_quote_fields_inner(schema: dict[str, Any]) -> dict[str, Any]:
    """
    Inner function that processes a schema and adds source___ fields for leaf nodes with X-SourceQuote: true.
    Only applies to leaf fields, never to the root.
    """
    if not isinstance(schema, dict):
        return schema

    # Create a copy to avoid modifying the original
    new_schema = copy.deepcopy(schema)

    # Process children recursively
    if "properties" in new_schema and isinstance(new_schema["properties"], dict):
        new_props = {}
        for property_key, property_value in new_schema["properties"].items():
            updated_prop_schema_value = _insert_quote_fields_inner(property_value)
            has_quote_field = updated_prop_schema_value.get("X-SourceQuote") is True

            # Check if this property is a leaf with X-SourceQuote: true
            if has_quote_field:
                # Add the quote field
                quote_key = f"source___{property_key}"
                new_props[quote_key] = {"type": "string", "description": f"The exact quote from the source document that supports the extracted value for '{property_key}'."}

                # Add the quote field to required if the property is required
                if "required" in new_schema and property_key in new_schema["required"]:
                    # add the quote field to required just before the property_key
                    new_schema["required"].insert(new_schema["required"].index(property_key), quote_key)

                # Remove the X-SourceQuote field
                updated_prop_schema_value.pop("X-SourceQuote", None)

            new_props[property_key] = updated_prop_schema_value
        new_schema["properties"] = new_props

    elif "items" in new_schema:
        # Recurse into items if present
        updated_items = _insert_quote_fields_inner(new_schema["items"])
        new_schema["items"] = updated_items

    # Process $defs as well
    if "$defs" in new_schema and isinstance(new_schema["$defs"], dict):
        new_defs = {}
        for dk, dv in new_schema["$defs"].items():
            new_defs[dk] = _insert_quote_fields_inner(dv)
        new_schema["$defs"] = new_defs

    return new_schema


def filter_auxiliary_fields(data: dict[str, Any], prefixes: list[str] = ["reasoning___", "source___"]) -> dict[str, Any]:
    """
    Recursively filters out fields that start with any of the prefixes in `prefixes` from the input data.
    """
    if not isinstance(data, dict):
        return data  # Base case: return non-dict values as is

    filtered: dict[str, Any] = {}
    for key, value in data.items():
        if not str(key).startswith(tuple(prefixes)):
            if isinstance(value, dict):
                filtered[key] = filter_auxiliary_fields(value, prefixes)
            elif isinstance(value, list):
                filtered[key] = [filter_auxiliary_fields(item, prefixes) if isinstance(item, dict) else item for item in value]
            else:
                filtered[key] = value

    return filtered


def create_reasoning_schema_without_ref_expansion(json_schema: dict[str, Any]) -> dict[str, Any]:
    """
    Create reasoning schema without expanding $refs, preserving the original reference structure.
    """
    # Work with the original schema, keeping $refs intact
    expanded_schema = copy.deepcopy(json_schema)

    # Insert reasoning fields.
    # We'll handle the root reasoning similarly: if root has reasoning, we add reasoning___root
    updated_schema, root_reasoning = _insert_reasoning_fields_inner(copy.deepcopy(expanded_schema))

    if root_reasoning:
        # Root is an object (assumed). Add reasoning___root at top-level properties
        if "properties" not in updated_schema:
            updated_schema["properties"] = {}
        add_reasoning_sibling_inplace(updated_schema["properties"], "root", root_reasoning)
        if "required" in updated_schema:
            updated_schema["required"].append("reasoning___root")

    # Insert quote fields for leaf nodes with X-SourceQuote: true
    updated_schema = _insert_quote_fields_inner(updated_schema)

    # Clean the schema (remove defaults, etc)
    updated_schema = clean_schema(updated_schema, remove_custom_fields=True)
    return updated_schema


def create_reasoning_schema(json_schema: dict[str, Any]) -> dict[str, Any]:
    # Resolve refs first to get expanded schema
    definitions = json_schema.get("$defs", {})
    resolved = expand_refs(copy.deepcopy(json_schema), definitions)
    # resolved.pop("$defs", None)

    expanded_schema = copy.deepcopy(resolved)

    # Insert reasoning fields.
    # We'll handle the root reasoning similarly: if root has reasoning, we add reasoning___root
    updated_schema, root_reasoning = _insert_reasoning_fields_inner(copy.deepcopy(expanded_schema))

    if root_reasoning:
        # Root is an object (assumed). Add reasoning___root at top-level properties
        if "properties" not in updated_schema:
            updated_schema["properties"] = {}
        add_reasoning_sibling_inplace(updated_schema["properties"], "root", root_reasoning)
        if "required" in updated_schema:
            updated_schema["required"].append("reasoning___root")

    # Insert quote fields for leaf nodes with X-SourceQuote: true
    updated_schema = _insert_quote_fields_inner(updated_schema)

    # Clean the schema (remove defaults, etc)
    updated_schema = clean_schema(updated_schema, remove_custom_fields=True)
    return updated_schema


def get_type_str(field_schema):
    """
    Recursively determine the type string for a given schema field.
    Handles 'anyOf' unions, enums, arrays, $ref references, and simple types.
    """
    if "$ref" in field_schema:
        return "reference"
    elif "anyOf" in field_schema:
        types = []
        for sub_schema in field_schema["anyOf"]:
            types.append(get_type_str(sub_schema))
        # Remove duplicates while preserving order
        seen = set()
        unique_types = []
        for t in types:
            if t not in seen:
                seen.add(t)
                unique_types.append(t)
        return " | ".join(unique_types)
    elif "enum" in field_schema:
        # Create a union of the literal enum values (as JSON strings)
        return " | ".join(json.dumps(val) for val in field_schema["enum"])
    elif "type" in field_schema:
        typ = field_schema["type"]
        if typ == "array" and "items" in field_schema:
            # For arrays, indicate the type of the items
            item_type = get_type_str(field_schema["items"])
            return f"array of {item_type}"
        return typ
    else:
        return "unknown"


def process_schema_field(field_name, field_schema, level, new_line_sep: str = "\n", field_name_prefix: str = ""):
    """
    Process a single field in the JSON schema.
    'level' indicates the header level (e.g., 3 for root, 4 for nested, etc.).
    Returns a markdown string representing the field.
    """
    md = ""
    field_name_complete = field_name_prefix + field_name

    # Handle $ref fields
    if "$ref" in field_schema:
        ref_value = field_schema["$ref"]
        header = "#" * level + f" {field_name_complete} (reference to {ref_value})"
        md += header + new_line_sep

        # Extract description (or use a placeholder if not provided)
        description = field_schema.get("description", None)
        if description is not None:
            md += f"<Description>\n{description}\n</Description>"
        else:
            md += f"<Description>Reference to {ref_value}</Description>"

        md += new_line_sep * 2
        return md

    # Extract type information
    type_str = get_type_str(field_schema)
    # md += f"**Type**: {type_str}{new_line_sep}"

    header = "#" * level + f" {field_name_complete} ({type_str})"
    md += header + new_line_sep

    # Extract description (or use a placeholder if not provided)
    description = field_schema.get("description", None)
    if description is not None:
        md += f"<Description>\n{description}\n</Description>"
    else:
        md += "<Description></Description>"

    md += new_line_sep * 2

    # If the field is an object with its own properties, process those recursively.
    if field_schema.get("type") == "object" and "properties" in field_schema:
        for sub_field_name, sub_field_schema in field_schema["properties"].items():
            md += process_schema_field(sub_field_name, sub_field_schema, level + 1, field_name_prefix=field_name_complete + ".")

    # If the field is an array and its items are objects with properties, process them.
    elif field_schema.get("type") == "array" and "items" in field_schema:
        items_schema = field_schema["items"]
        # Handle $ref in array items
        if "$ref" in items_schema:
            ref_value = items_schema["$ref"]
            md += "#" * (level + 1) + f" {field_name_complete}.* (reference to {ref_value})" + new_line_sep
            md += f"<Description>Array items reference {ref_value}</Description>" + new_line_sep * 2
        elif items_schema.get("type") == "object" and "properties" in items_schema:
            md += process_schema_field("*", items_schema, level + 1, field_name_prefix=field_name_complete + ".")

    return md


def json_schema_to_nlp_data_structure(schema: dict) -> str:
    """
    Receives a JSON schema and returns a markdown string that documents each field
    with its name, description, type (including unions and enums), and default value
    (if defined). Root-level fields use 3 hashtags, and nested fields add one hashtag
    per level. $ref references are preserved and shown as-is without expansion.
    Includes definitions from $defs section.
    """
    schema_title = schema.get("title", schema.get("name", "Schema"))
    md = f"## {schema_title} -- NLP Data Structure\n\n"

    # Add the description of the schema
    description = schema.get("description", None)
    if description is not None:
        md += f"<User Prompt>\n{description}\n</User Prompt>"

    # Assume the root schema is an object with properties.
    if schema.get("type") == "object" and "properties" in schema:
        for field_name, field_schema in schema["properties"].items():
            md += process_schema_field(field_name, field_schema, 3)
    else:
        md += process_schema_field("root", schema, 3)

    # Process definitions from $defs
    defs = schema.get("$defs", {})
    if defs:
        md += "\n## Definitions\n\n"
        for def_name, def_schema in defs.items():
            md += f"### {def_name}\n\n"

            # Add definition description if available
            def_description = def_schema.get("description", None)
            if def_description is not None:
                md += f"<Description>\n{def_description}\n</Description>\n\n"
            else:
                md += f"<Description>Definition for {def_name}</Description>\n\n"

            # Process definition properties if it's an object
            if def_schema.get("type") == "object" and "properties" in def_schema:
                for prop_name, prop_schema in def_schema["properties"].items():
                    md += process_schema_field(prop_name, prop_schema, 4, field_name_prefix=f"{def_name}.")
            else:
                # If it's not an object, show its type and description
                type_str = get_type_str(def_schema)
                md += f"**Type**: {type_str}\n\n"

    return md


def json_schema_to_strict_openai_schema(obj: Union[dict[str, Any], list[Any]]) -> Union[dict[str, Any], list[Any]]:
    # Gets a json supported by GPT Structured Output from a pydantic Basemodel

    if isinstance(obj, dict):
        new_obj: dict[str, Any] = copy.deepcopy(obj)

        # Remove some not-supported fields
        for key in ["default", "format", "X-FieldTranslation", "X-EnumTranslation"]:
            new_obj.pop(key, None)

        # Handle integer type
        if "type" in new_obj:
            if new_obj["type"] == "integer":
                new_obj["type"] = "number"
            elif isinstance(new_obj["type"], list):
                new_obj["type"] = ["number" if t == "integer" else t for t in new_obj["type"]]

        # Handle allOf
        if "allOf" in new_obj:
            subschemas = new_obj.pop("allOf")
            merged: dict[str, Any] = {}
            for subschema in subschemas:
                if "$ref" in subschema:
                    merged.update({"$ref": subschema["$ref"]})
                else:
                    merged.update(json_schema_to_strict_openai_schema(subschema))
            new_obj.update(merged)

        # Handle anyOf
        if "anyOf" in new_obj:
            new_obj["anyOf"] = [json_schema_to_strict_openai_schema(subschema) for subschema in new_obj["anyOf"]]

        # Handle enum (force type to string)
        if "enum" in new_obj:
            new_obj["enum"] = [str(e) for e in new_obj["enum"]]
            new_obj["type"] = "string"

        # Handle object type
        if new_obj.get("type") == "object" and "properties" in new_obj and isinstance(new_obj["properties"], dict):
            new_obj["required"] = list(new_obj["properties"].keys())
            new_obj["additionalProperties"] = False
            new_obj["properties"] = {k: json_schema_to_strict_openai_schema(v) for k, v in new_obj["properties"].items()}

        # Handle array type
        if new_obj.get("type") == "array" and "items" in new_obj:
            new_obj["items"] = json_schema_to_strict_openai_schema(new_obj["items"])

        # Handle defs
        if "$defs" in new_obj:
            new_obj["$defs"] = {k: json_schema_to_strict_openai_schema(v) for k, v in new_obj["$defs"].items()}

        return new_obj
    elif isinstance(obj, list):
        return [json_schema_to_strict_openai_schema(item) for item in obj]
    else:
        return obj


def json_schema_to_typescript_interface(
    schema: dict[str, Any],
    name: str = "RootInterface",
    definitions: Optional[dict[str, dict[str, Any]]] = None,
    processed_refs: Optional[dict[str, str]] = None,
    indent: int = 2,
    add_field_description: bool = False,
) -> str:
    """
    Convert a JSON Schema to a TypeScript interface.

    :param schema: The JSON schema as a dict.
    :param name: Name of the interface to generate.
    :param definitions: A dictionary of named schemas that can be referenced by $ref.
    :param processed_refs: A dict to keep track of processed $refs to avoid recursion.
    :param indent: Number of spaces for indentation in the output.
    :param add_field_description: If True, include field descriptions as comments.
    :return: A string containing the TypeScript interface.
    """
    if definitions is None:
        # Extract definitions from $defs if present
        definitions = schema.get("$defs", {})

    if processed_refs is None:
        processed_refs = {}

    # If we have a top-level object schema
    if schema.get("type") == "object" or "properties" in schema:
        interface_lines = [f"interface {name} {{"]
        indentation = " " * indent
        properties = schema.get("properties", {})
        required_fields = set(schema.get("required", []))

        for prop_name, prop_schema in properties.items():
            is_optional = prop_name not in required_fields
            field_ts = schema_to_ts_type(prop_schema, definitions or {}, processed_refs, indent, indent, add_field_description=add_field_description)
            optional_flag = "?" if is_optional else ""
            line = ""
            if add_field_description and "description" in prop_schema:
                desc = prop_schema["description"].replace("\n", f"\n{indentation}// ")
                line = f"{indentation}// {desc}\n"
            line += f"{indentation}{prop_name}{optional_flag}: {field_ts};"
            interface_lines.append(line)

        interface_lines.append("}")
        return "\n".join(interface_lines)
    else:
        # Otherwise, produce a type alias if it's not an object
        ts_type = schema_to_ts_type(schema, definitions or {}, processed_refs, indent, indent, add_field_description=add_field_description)
        return f"type {name} = {ts_type};"


def clean_schema(schema: dict[str, Any], remove_custom_fields: bool = False, fields_to_remove: list[str] = ["default", "minlength", "maxlength"]) -> dict[str, Any]:
    """
    Recursively remove specified fields from a JSON schema.

    Args:
        schema: The JSON schema to be cleaned.
        remove_custom_fields: If True, also remove fields starting with 'x-'.
        fields_to_remove: List of keys to remove (case-insensitive check).

    Returns:
        The resulting cleaned JSON schema.
    """
    schema = schema.copy()
    lower_fields_to_remove = [f.lower() for f in fields_to_remove]
    for key in list(schema.keys()):
        if not isinstance(key, str):
            continue

        lower_key = key.lower()

        conditions_to_remove = [
            # Empty keys
            not key,
            # Empty subschemas
            isinstance(schema[key], dict) and len(schema[key]) == 0,
            # Fields to remove
            lower_key in lower_fields_to_remove,
            # Custom fields
            remove_custom_fields and lower_key.startswith("x-"),
        ]

        if any(conditions_to_remove):
            schema.pop(key)
            continue

    if "properties" in schema:
        schema["properties"] = {
            prop_key: clean_schema(prop_schema, fields_to_remove=fields_to_remove, remove_custom_fields=remove_custom_fields)
            for prop_key, prop_schema in schema["properties"].items()
        }
    if "items" in schema:
        schema["items"] = clean_schema(schema["items"], fields_to_remove=fields_to_remove, remove_custom_fields=remove_custom_fields)
    if "$defs" in schema:
        schema["$defs"] = {k: clean_schema(v, fields_to_remove=fields_to_remove, remove_custom_fields=remove_custom_fields) for k, v in schema["$defs"].items()}
    if "allOf" in schema:
        schema["allOf"] = [clean_schema(subschema, fields_to_remove=fields_to_remove, remove_custom_fields=remove_custom_fields) for subschema in schema["allOf"]]
    if "anyOf" in schema:
        schema["anyOf"] = [clean_schema(subschema, fields_to_remove=fields_to_remove, remove_custom_fields=remove_custom_fields) for subschema in schema["anyOf"]]

    return schema


def generate_schema_data_id(json_schema: dict[str, Any]) -> str:
    """Generate a SHA1 hash ID for schema data, ignoring prompt/description/default fields.

    Args:
        json_schema: The JSON schema to generate an ID for

    Returns:
        str: A SHA1 hash string with "sch_data_id_" prefix
    """
    return "sch_data_id_" + generate_blake2b_hash_from_string(
        json.dumps(
            clean_schema(
                copy.deepcopy(json_schema),
                remove_custom_fields=True,
                fields_to_remove=["description", "default", "title", "required", "examples", "deprecated", "readOnly", "writeOnly"],
            ),
            sort_keys=True,
        ).strip()
    )


def generate_schema_id(json_schema: dict[str, Any]) -> str:
    """Generate a SHA1 hash ID for the complete schema.

    Args:
        json_schema: The JSON schema to generate an ID for

    Returns:
        str: A SHA1 hash string with "sch_id_" prefix
    """
    return "sch_id_" + generate_blake2b_hash_from_string(json.dumps(json_schema, sort_keys=True).strip())


def primitive_type_to_ts(t: Union[str, None]) -> str:
    """
    Convert a primitive JSON schema type to a TypeScript type.
    """
    if t == "string":
        return "string"
    elif t in ("integer", "number"):
        return "number"
    elif t == "boolean":
        return "boolean"
    elif t == "null":
        return "null"
    elif t is None:
        # no specific type given
        return "any"
    else:
        # fallback
        return "any"


def resolve_ref(ref: str, definitions: dict[str, dict[str, Any]]) -> Optional[dict[str, Any]]:
    """
    Resolve a $ref against the given definitions.
    The schema uses $defs. Ref format: "#/$defs/SomeDefinition"
    """
    if ref.startswith("#/$defs/"):
        key = ref[len("#/$defs/") :]
        return definitions.get(key)
    # No known resolution strategy
    return None


def inline_object(schema: dict[str, Any], definitions: dict[str, dict[str, Any]], processed_refs: dict[str, str], indent: int, increment: int, add_field_description: bool) -> str:
    """
    Inline an object type from a JSON schema into a TypeScript type.
    """
    properties = schema.get("properties", {})
    required_fields = set(schema.get("required", []))
    lines = ["{"]
    field_indentation = " " * (indent + increment)
    for prop_name, prop_schema in properties.items():
        is_optional = prop_name not in required_fields
        ts_type = schema_to_ts_type(prop_schema, definitions, processed_refs, indent + increment, increment, add_field_description)
        optional_flag = "?" if is_optional else ""
        line = ""
        if add_field_description and "description" in prop_schema:
            desc = prop_schema["description"].replace("\n", f"\n{field_indentation}// ")
            line = f"{field_indentation}// {desc}\n"
        line += f"{field_indentation}{prop_name}{optional_flag}: {ts_type};"
        lines.append(line)
    lines.append(" " * indent + "}")
    return "\n".join(lines)


def schema_to_ts_type(
    schema: dict[str, Any], definitions: dict[str, dict[str, Any]], processed_refs: dict[str, str], indent: int, increment: int, add_field_description: bool = False
) -> str:
    """
    Convert a JSON schema snippet to a TypeScript type (string).
    Handles objects, arrays, primitives, enums, oneOf/anyOf/allOf, and $ref.
    """

    # Handle $ref upfront
    if "$ref" in schema:
        ref = schema["$ref"]
        if ref in processed_refs:
            return processed_refs[ref]
        resolved = resolve_ref(ref, definitions)
        if resolved is None:
            return "any"
        processed_refs[ref] = ""  # to avoid recursion
        ts_type = schema_to_ts_type(resolved, definitions, processed_refs, indent, increment, add_field_description=add_field_description)
        processed_refs[ref] = ts_type
        return ts_type

    # Handle allOf, oneOf, anyOf
    if "allOf" in schema:
        # allOf means intersection of all subschemas
        subtypes = [schema_to_ts_type(s, definitions, processed_refs, indent, increment, add_field_description) for s in schema["allOf"]]
        return "(" + " & ".join(subtypes) + ")"

    if "oneOf" in schema:
        # oneOf means a union type
        subtypes = [schema_to_ts_type(s, definitions, processed_refs, indent, increment, add_field_description) for s in schema["oneOf"]]
        return "(" + " | ".join(subtypes) + ")"

    if "anyOf" in schema:
        # anyOf means a union type
        subtypes = [
            schema_to_ts_type(s, definitions, processed_refs, indent, increment, add_field_description)
            for s in schema["anyOf"]
            # Remove "null" from subtypes if it's present
            # if not (isinstance(s, dict) and s.get("type") == "null")
        ]
        if len(subtypes) == 1:
            return subtypes[0]

        return "(" + " | ".join(subtypes) + ")"

    # Handle enums
    if "enum" in schema:
        # Create a union of literal types
        enum_values = schema["enum"]
        ts_literals = []
        for val in enum_values:
            if isinstance(val, str):
                ts_literals.append(f'"{val}"')
            elif val is None:
                ts_literals.append("null")
            else:
                ts_literals.append(str(val).lower() if isinstance(val, bool) else str(val))
        return " | ".join(ts_literals)

    # Handle type
    schema_type = schema.get("type")
    if schema_type == "object" or "properties" in schema:
        # Inline object
        return inline_object(schema, definitions, processed_refs, indent, increment, add_field_description)
    elif schema_type == "array":
        # items define the type of array elements
        items_schema = schema.get("items", {})
        item_type = schema_to_ts_type(items_schema, definitions, processed_refs, indent + increment, increment, add_field_description)
        return f"Array<{item_type}>"
    else:
        # Primitive types or missing type
        if isinstance(schema_type, list):
            # union of multiple primitive types
            primitive_types = [primitive_type_to_ts(t) for t in schema_type]
            return "(" + " | ".join(primitive_types) + ")"
        else:
            # single primitive
            return primitive_type_to_ts(schema_type)


def flatten_dict(obj: Any, prefix: str = "", allow_empty_objects: bool = True) -> dict[str, Any]:
    items = []  # type: ignore
    if isinstance(obj, dict):
        if len(obj) == 0 and allow_empty_objects and prefix != "":
            # Keep empty dicts as dicts (so we can keep its structure, but not if it's the root)
            items.append((prefix, {}))
        else:
            for k, v in obj.items():
                new_key = f"{prefix}.{k}" if prefix else k
                items.extend(flatten_dict(v, new_key, allow_empty_objects=allow_empty_objects).items())

    elif isinstance(obj, list):
        if len(obj) == 0 and allow_empty_objects and prefix != "":
            # Keep empty lists as lists (so we can keep its structure, but not if it's the root)
            items.append((prefix, []))
        else:
            for i, v in enumerate(obj):
                new_key = f"{prefix}.{i}"
                items.extend(flatten_dict(v, new_key, allow_empty_objects=allow_empty_objects).items())
    else:
        items.append((prefix, obj))
    return dict(items)


def convert_dict_to_list_recursively(_obj: Any, allow_lists: bool = True) -> Any:
    """
    Recursively converts dict[int, Any] to list[Any] if the keys are sequential integers starting from 0.
    Creates a copy of the input object rather than modifying it in place.
    """
    # Handle non-dict types
    if not isinstance(_obj, dict):
        return _obj

    # Create a copy to avoid modifying the original
    result = {}

    # Process all nested dictionaries first
    for key, value in _obj.items():
        result[key] = convert_dict_to_list_recursively(value, allow_lists=allow_lists)

    # Check if this dictionary should be converted to a list
    if result and all(isinstance(k, int) for k in result.keys()):
        # Check if keys are sequential starting from 0
        keys = sorted(result.keys())
        if allow_lists and keys[0] == 0 and keys[-1] == len(keys) - 1:
            # Convert to list
            return [result[i] for i in keys]
        else:
            # Sort the keys and convert to string
            return {str(i): result[i] for i in keys}

    return result


def unflatten_dict(obj: dict[str, Any], allow_lists: bool = True) -> Any:
    """
    Unflattens a dictionary by recursively converting keys with dots into nested dictionaries.
    After building the nested structure, converts dict[int, Any] to list[Any] if the keys
    are sequential integers starting from 0.

    Args:
        obj: The dictionary to unflatten.

    Returns:
        The unflattened dictionary with appropriate dict[int, Any] converted to list[Any].
    """
    # Handle empty input
    if not obj:
        return obj

    # Create a copy of the input object to avoid modifying it
    input_copy = dict(obj)

    # Optionally validate that the dict is indeed flat
    # Commented out to avoid potential equality issues with key ordering
    # assert flatten_dict(input_copy) == input_copy, "Dictionary is not flat"

    # First pass: build everything as nested dictionaries
    result = {}
    for key, value in input_copy.items():
        # Skip invalid keys
        if not isinstance(key, str):
            continue

        parts = key.split(".")
        # Filter out empty parts
        valid_parts = [p for p in parts if p]
        if not valid_parts:
            result[key] = value
            continue

        current = result

        for i, part in enumerate(valid_parts):
            # Check if the part is an integer (for list indices)
            try:
                # More robust integer parsing - handles negative numbers too
                if part.lstrip("-").isdigit():
                    part = int(part)
            except (ValueError, AttributeError):
                # If conversion fails, keep as string
                pass

            # If at the last part, set the value
            if i == len(valid_parts) - 1:
                current[part] = value
            else:
                # Create the container if it doesn't exist
                if part not in current:
                    current[part] = {}
                elif not isinstance(current[part], dict):
                    # Handle case where we're trying to nest under a non-dict
                    # This is a conflict - the path is both a value and used as a prefix
                    current[part] = {}

                current = current[part]

    # Second pass: convert appropriate dict[int, Any] to list[Any]
    return convert_dict_to_list_recursively(result, allow_lists=allow_lists)


def has_cyclic_refs(schema: dict[str, Any]) -> bool:
    """Check if the JSON Schema contains cyclic references.

    The function recursively traverses all nested objects and arrays in the schema.
    It follows any "$ref" that points to a definition (i.e. "#/$defs/<name>")
    and uses DFS with a current-path stack to detect cycles.
    """
    definitions = schema.get("$defs", {})
    if not definitions:
        return False

    # Memoize results for each definition to avoid repeated work.
    memo: dict[str, bool] = {}

    def dfs(def_name: str, stack: set[str]) -> bool:
        """Perform DFS on a definition (by name) using 'stack' to detect cycles."""
        if def_name in stack:
            return True
        if def_name in memo:
            return memo[def_name]

        # Add to current path and traverse the definition.
        stack.add(def_name)
        node = definitions.get(def_name)
        if node is None:
            # No such definition, so nothing to do.
            stack.remove(def_name)
            memo[def_name] = False
            return False

        result = traverse(node, stack)
        stack.remove(def_name)
        memo[def_name] = result
        return result

    def traverse(node: Any, stack: set[str]) -> bool:
        """Recursively traverse an arbitrary JSON Schema node."""
        if isinstance(node, dict):
            # If we see a "$ref", try to follow it.
            if "$ref" in node:
                ref = node["$ref"]
                if ref.startswith("#/$defs/"):
                    target = ref[len("#/$defs/") :]
                    if dfs(target, stack):
                        return True
            # Recursively check all values in the dictionary.
            for key, value in node.items():
                # Skip "$ref" as it has already been processed.
                if key == "$ref":
                    continue
                if traverse(value, stack):
                    return True
        elif isinstance(node, list):
            for item in node:
                if traverse(item, stack):
                    return True
        return False

    # Start DFS on each top-level definition.
    for def_name in definitions:
        if dfs(def_name, set()):
            return True

    return False


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

    _partial_pydantic_model: type[BaseModel] | None = PrivateAttr(default=None)
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
        reasoning_schema = create_reasoning_schema_without_ref_expansion(self.json_schema)
        return json_schema_to_nlp_data_structure(reasoning_schema)

    @property
    def developer_system_prompt(self) -> str:
        return """# General Instructions

You are an expert in data extraction and structured outputs.

Given a **JSON schema** and a **document**, you must:

1. Extract all relevant data from the document, adhering strictly to the schema.
2. Output the extracted data in exact schema format.
3. Ensure all values are UTF-8 encodable strings; avoid bytes, binary, base64, or non-UTF-8 data.
4. Be exhaustive: Fully populate objects and arrays (including nested ones) with available data.

---

## Date and Time Formatting

Extract dates, times, or datetimes in ISO 8601 format only (e.g., "2023-12-25", "14:30:00", "2023-12-25T14:30:00Z", "2023-12-25T14:30:00+02:00").

Avoid non-ISO formats like "12/25/2023" or "Dec 25, 2023 at 2:30 PM".

---

## Handling Missing and Nullable Fields

Nullable fields are indicated by `anyOf` including a `null` type in the schema. Fields with a single type (e.g., "string", "number") are non-nullable.

- User instructions take precedence over schema rules.
- For nullable fields with missing data: Use JSON `null`.
- For non-nullable fields with missing data: Use type-appropriate placeholders:
  - String: `""`
  - Number: `0`
  - Boolean: `false`
  - Array: `[]`
  - Object: See [Nullable Nested Objects](#nullable-nested-objects)
- Never use the string `"null"` for missing data.

### Decision Flow
- Data missing? → Nullable? → Yes: `null` → No: Placeholder.

### Examples

- Nullable string: `{"email": null}` (not `""`)
- Non-nullable string: `{"email": ""}` (not `"null"`)
- Nullable number: `{"age": null}` (not `0`)
- Non-nullable number: `{"age": 0}` (not `"null"`)
- Nullable boolean: `{"isActive": null}` (not `false`)
- Non-nullable boolean: `{"isActive": false}` (not `"null"`)
- Nullable array: `{"tags": null}` (not `[]`)
- Non-nullable array: `{"tags": []}` (not `"null"`)

---

## Nullable Nested Objects

- Nullable object missing entirely: Set to `null`.
- Non-nullable object missing entirely: Include the object with all fields following missing-data rules.
- Partial data: Include all fields, using `null` or placeholders as appropriate. No incomplete objects.

### Examples

- Nullable object missing: `{"address": null}`
- Non-nullable object missing: `{"address": {"street": "", "zipCode": 0, "city": ""}}`
- Partial: `{"address": {"street": "", "zipCode": 0, "city": "Paris"}}` (not `{"address": {"city": "Paris"}}`)

---

## Reasoning Fields

The schema includes reasoning fields (`reasoning___*`) for documenting extraction logic. These are internal and omitted from final outputs.

Naming:
- Root: `reasoning___root`
- Nested object: `reasoning___[objectname]`
- Array: `reasoning___[arrayname]`
- Array item: `reasoning___item`
- Leaf: `reasoning___[attributename]`

Include:
- Evidence: Direct quotes or references from the document.
- Justification: Why data was selected/rejected.
- Transformations: Any calculations, conversions, or normalizations.
- Alternatives: Considered options and rejection reasons.
- Confidence/Assumptions: Level of certainty and any assumptions.

### Example (Leaf Field)
"Company name 'ACME Corp' from page 1 top-right letterhead and page 3 signature. Matches 'Client: ACME Corp'. High confidence; rejected as sender due to explicit label."

### Array Reasoning Example
"Invoice items from page 2, lines 12–17 under 'Invoice Items' header:
1. Office Supplies: qty 5, $4.99, total $24.95 (line 12)
2. Printer Paper: qty 1, $5.99, total $5.99 (line 13)
No ambiguities."

### Array Item Example
"Line 12: 'Office Supplies x5 $4.99ea $24.95'. Qty * price = total verified. Consistent format; high confidence."

---

## Source Quote Fields

The schema may include source quote fields (`source___*`) for capturing exact quotes from the document that support extracted values. These fields appear as siblings to the fields they document.

Naming:
- `source___[fieldname]` for each field marked with X-SourceQuote in the schema

Guidelines:
- Extract the exact verbatim text from the document that supports the extracted value.
- Include surrounding context when helpful for verification.
- For missing data, use an empty string `""`.
- These fields are internal and omitted from final outputs.

### Example
If extracting a company name with source quote:
```json
{
  "source___company_name": "Registered Office: ACME Corporation Ltd",
  "company_name": "ACME Corporation Ltd"
}
```

---

## Extraction Principles

- **Transparency**: Justify every decision with evidence.
- **Precision**: Base extractions on direct document quotes.
- **Conservatism**: Use `null` or placeholders for missing/ambiguous data; avoid invention.
- **Structure Integrity**: Preserve full schema, populating all elements per rules."""

    @property
    def schema_system_prompt(self) -> str:
        return (
            self.inference_nlp_data_structure #+ "\n---\n" + "## Expected output schema as a TypeScript interface for better readability:\n\n" + self.inference_typescript_interface
        )

    @property
    def system_prompt(self) -> str:
        """Returns the system prompt combining custom prompt and TypeScript interface.

        Returns:
            str: The combined system prompt string.
        """
        return self.developer_system_prompt + "\n\n" + self.schema_system_prompt
 
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
            # Load and normalize the JSON Schema, but defer heavy BaseModel generation
            # to first use to avoid blocking the event loop during initialization.
            json_schema = load_json_schema(json_schema)
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
        # Defer partial model creation until/if it's needed to avoid blocking.
        try:
            self._partial_pydantic_model  # type: ignore[attr-defined]
        except AttributeError:
            # Initialize lazily; will be computed by caller if required.
            object.__setattr__(self, "_partial_pydantic_model", None)  # type: ignore[attr-defined]

        return self

    # Lazily compute heavy artifacts to avoid blocking the event loop at init-time
    def __getattribute__(self, name: str) -> Any:  # type: ignore[override]
        if name == "pydantic_model":
            # Access underlying dict directly to avoid recursion
            d = object.__getattribute__(self, "__dict__")
            model = d.get("pydantic_model", None)
            if model is None:
                # Build and cache on first access
                json_schema = object.__getattribute__(self, "json_schema")
                from retab.utils.json_schema import convert_json_schema_to_basemodel as _conv
                model = _conv(json_schema)
                object.__setattr__(self, "pydantic_model", model)
            return model
        if name == "_partial_pydantic_model":
            d = object.__getattribute__(self, "__dict__")
            partial = d.get("_partial_pydantic_model", None)
            if partial is None:
                # Derive from full model lazily
                base_model = object.__getattribute__(self, "pydantic_model")
                from retab.utils.json_schema import convert_basemodel_to_partial_basemodel as _to_partial
                partial = _to_partial(base_model)
                object.__setattr__(self, "_partial_pydantic_model", partial)
            return partial
        return object.__getattribute__(self, name)

    def save(self, path: Path | str) -> None:
        """Save a JSON schema to a file.

        Args:
            json_schema: The JSON schema to save, can be a dict, Path, or string
            schema_path: Output path for the schema file
        """
        with open(path, "w", encoding="utf-8") as f:
            json.dump(self.json_schema, f, ensure_ascii=False, indent=2)


def validate_json_against_schema(
    data: Any,
    schema: dict[str, Any],
    return_instance: bool = False,
) -> Union[None, BaseModel]:
    """
    Validate *data* against *schema*.

    Parameters
    ----------
    data
        A JSON‑serialisable Python object (dict / list / primitives).
    schema
        A JSON‑Schema dict (can contain $defs / $ref – they’ll be expanded
        by ``convert_json_schema_to_basemodel``).
    return_instance
        • ``False`` (default): only validate; raise if invalid; return ``None``.
        • ``True``: on success, return the fully‑validated Pydantic instance
          (handy for downstream type‑safe access).

    Raises
    ------
    pydantic.ValidationError
        If *data* does not conform to *schema*.

    Examples
    --------
    >>> validate_json_against_schema({"foo": 1}, my_schema)        # just checks
    >>> obj = validate_json_against_schema(data, schema, True)     # typed access
    >>> print(obj.foo + 5)
    """
    # 1) Build a Pydantic model on‑the‑fly from the JSON‑Schema
    Model: Type[BaseModel] = convert_json_schema_to_basemodel(schema)

    # 2) Let Pydantic do the heavy lifting
    instance = Model.model_validate(data)  # <- raises ValidationError if bad

    return instance if return_instance else None


SchemaPath = Tuple[Union[str, int], ...]  # e.g. ('address', 'city') or ('items', 3)


def _pick_subschema(schemas: list[dict[str, Any]], value: Any) -> dict[str, Any]:
    """
    Return the first subschema in *schemas* that
      • explicitly allows the Python type of *value*, or
      • has no "type" at all (acts as a wildcard).

    Fallback: the first subschema (so we *always* return something).
    """
    pytypes_to_json = {
        str: "string",
        int: "integer",
        float: "number",
        bool: "boolean",
        type(None): "null",
        dict: "object",
        list: "array",
    }
    jstype = pytypes_to_json.get(type(value))

    for sub in schemas:
        allowed = sub.get("type")
        if allowed is None or allowed == jstype or (isinstance(allowed, list) and jstype in allowed):
            return sub
    return schemas[0]  # last resort


def __sanitize_instance(instance: Any, schema: dict[str, Any], path: SchemaPath = ()) -> Any:
    """
    Return a **new** instance where every string that violates ``maxLength``
    has been sliced to that length.  Mutates nothing in‑place.
    """

    # ------------- unwrap anyOf ------------------------------------
    if "anyOf" in schema:
        schema = _pick_subschema(schema["anyOf"], instance)
        # (We recurse *once*; nested anyOfs will be handled the same way)

    # ------------- objects -----------------
    if schema.get("type") == "object" and isinstance(instance, MutableMapping):
        props = schema.get("properties", {})
        return {k: __sanitize_instance(v, props.get(k, {}), path + (k,)) for k, v in instance.items()}

    # ------------- arrays ------------------
    if schema.get("type") == "array" and isinstance(instance, MutableSequence):
        item_schema = schema.get("items", {})
        return [__sanitize_instance(v, item_schema, path + (i,)) for i, v in enumerate(instance)]

    # ------------- primitive strings -------
    if schema.get("type") == "string" and isinstance(instance, str):
        max_len = schema.get("maxLength")
        if max_len is not None and len(instance) > max_len:
            print("=" * 100)
            _path = ".".join(map(str, path)) or "<root>"
            print(
                f"Trimmed {_path} from {len(instance)}→{max_len} characters",
            )
            print("=" * 100)
            return instance[:max_len]

    # ------------- all other primitives ----
    return instance


def sanitize(instance: Any, schema: dict[str, Any]) -> Any:
    expanded_schema = expand_refs(schema)
    return __sanitize_instance(instance, expanded_schema)
 
"""
# General Instructions

You are an expert in data extraction and structured data outputs.

When provided with a **JSON schema** and a **document**, you must:

1. Carefully extract all relevant data from the provided documents according to the given schema.
2. Return extracted data strictly formatted according to the provided schema.
3. Make sure that the extracted values are **UTF-8** encodable strings.
4. Avoid generating bytes, binary data, base64 encoded data, or other non-UTF-8 encodable data.
5. Be comprehensive, do not miss any data - Objects must be fully populated if available, arrays must be fully populated if available (even if nested).
---

## Date and Time Formatting

When extracting date, time, or datetime values:

- **Always use ISO format** for dates and times (e.g., "2023-12-25", "14:30:00", "2023-12-25T14:30:00")

**Examples:**

// Correct ISO formats:
{"date": "2023-12-25"}
{"time": "14:30:00"}
{"datetime": "2023-12-25T14:30:00Z"}
{"datetime_with_tz": "2023-12-25T14:30:00+02:00"}

// Incorrect formats:
{"date": "12/25/2023"}
{"time": "2:30 PM"}
{"datetime": "Dec 25, 2023 at 2:30 PM"}

---

## Handling Missing and Nullable Fields

### General Rules
First of all: nullable fields can be identified by the presence of anyOf with a null type in the JSON Schema.
If a field has just a specific type (like "string", "number", "boolean", "array", "object"), it is not nullable.

- **User instructions override all schema rules.**  
- If a field is explicitly defined as **nullable** and the data is missing, set its value to `null` (the JSON null type).  
- If a field is **not nullable** and the data is missing:
  - Do **not** use `"null"` (string). <- NEVER ever use `"null"` to represent missing data.
  - Instead, provide a valid placeholder according to the field's type:  
    - **string** → `""` (empty string)  
    - **number** → `0`  
    - **boolean** → `false`  
    - **array** → `[]`  
    - **object** → see [Nullable Nested Objects](#nullable-nested-objects)  
    - (other types: follow their respective neutral/empty valid value)

### Quick Decision Flow
Missing data?  
→ Is the field nullable?  
    → Yes → use `null`  
    → No → use type-specific placeholder (e.g. `""`, `0`, `false`, `[]`, `{}`)

### **Prohibited Values**
- Never use the literal string `"null"`.

---

### Examples

#### Nullable string field:
// Correct:
{"email": null}

// Incorrect:
{"email": ""}

#### **Non-nullable** string field:
// Correct:
{"email": ""}

// Incorrect:
{"email": "null"}

#### Nullable number field:
// Correct:
{"age": null}

// Incorrect:
{"age": 0}

#### Non-nullable number field:
// Correct:
{"age": 0}

// Incorrect:
{"age": "null"}

#### Nullable boolean field:
// Correct:
{"isActive": null}

// Incorrect:
{"isActive": false}

#### Non-nullable boolean field:
// Correct:
{"isActive": false}

// Incorrect:
{"isActive": "null"}

#### Nullable array field:
// Correct:
{"tags": null}

// Incorrect:
{"tags": []}

#### Non-nullable array field:
// Correct:
{"tags": []}

// Incorrect:
{"tags": "null"}

---

### Nullable Nested Objects

- If a nested object is explicitly **nullable** and the entire object's data is missing, set the object itself to `null`.  
- If the object is **not nullable** and data is missing, preserve the object structure and apply the missing-data rules to its fields.  
- Partially filled objects are **not allowed**:  
  - If you provide a value for one field, you must include **all other fields** as well, filling them according to the rules (e.g. `null` if nullable, or the appropriate placeholder if non-nullable).  

**Examples:**

#### Nullable object (entirely missing):
// Correct:
{
  "address": null
}

#### Non-nullable object (entirely missing):
// Correct:
{
  "address": {
    "street": null,
    "zipCode": null,
    "city": null
  }
}

#### Partially known object:
// Correct:
{
  "address": {
    "street": null,
    "zipCode": null,
    "city": "Paris"
  }
}

// Incorrect (missing other fields):
{
  "address": {
    "city": "Paris"
  }
}

---

## Reasoning Fields

Your schema includes special reasoning fields (`reasoning___*`) used exclusively to document your extraction logic. These fields are for detailed explanations and will not appear in final outputs.

| Reasoning Field Type | Field Naming Pattern       |
|----------------------|----------------------------|
| Root Object          | `reasoning___root`         |
| Nested Objects       | `reasoning___[objectname]` |
| Array Fields         | `reasoning___[arrayname]`  |
| Array Elements       | `reasoning___item`         |
| Leaf Attributes      | `reasoning___[attributename]` |

You MUST include these details explicitly in your reasoning fields:

- **Explicit Evidence**: Quote specific lines or phrases from the document confirming your extraction.
- **Decision Justification**: Clearly justify why specific data was chosen or rejected.
- **Calculations/Transformations**: Document explicitly any computations, unit conversions, or normalizations.
- **Alternative Interpretations**: Explicitly describe any alternative data interpretations considered and why you rejected them.
- **Confidence and Assumptions**: Clearly state your confidence level and explicitly articulate any assumptions.

**Example Reasoning:**

> Found company name 'ACME Corp' explicitly stated in the top-right corner of page 1, matching standard letterhead format. Confirmed by matching signature block ('ACME Corp') at bottom of page 3. Confidence high. Alternative interpretation (e.g., sender's name) explicitly rejected due to explicit labeling 'Client: ACME Corp' on page 1.

---

## Detailed Reasoning Examples

### Array Reasoning (`reasoning___[arrayname]`)

- Explicitly describe how the entire array was identified.
- List explicitly all extracted items with clear details and source references.

**Example:**

~~~markdown
Identified itemized invoice section clearly demarcated by header "Invoice Items" (page 2, lines 12–17). Extracted items explicitly listed:

1. Office Supplies, quantity 5, unit price $4.99, total $24.95 (line 12)
2. Printer Paper, quantity 1, unit price $5.99, total $5.99 (line 13)
3. Stapler, quantity 1, unit price $4.07, total $4.07 (line 14)

No ambiguity detected.
~~~

### Array Item Reasoning (`reasoning___item`)

Explicitly document evidence for each individual item:

~~~markdown
Extracted explicitly from line 12: 'Office Supplies x5 $4.99ea $24.95'. Quantity (5 units) multiplied explicitly by unit price ($4.99) matches listed total ($24.95). Format consistent across invoice, high confidence.
~~~

---

## Principles for Accurate Extraction

When performing extraction, explicitly follow these core principles:

- **Transparency**: Explicitly document and justify every extraction decision.
- **Precision**: Always verify explicitly using direct quotes from the source document.
- **Conservatism**: Set explicitly fields as `null` when data is explicitly missing or ambiguous—never fabricate or guess.
- **Structure Preservation**: Always maintain explicitly the full schema structure, even when entire nested objects lack data (leaf attributes as null).


---"""