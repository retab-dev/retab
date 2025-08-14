import copy
import json
import types
from pathlib import Path
from typing import Any, Type, Union, cast, get_args, get_origin, Literal

from pydantic import BaseModel, Field, create_model
from pydantic.config import ConfigDict



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

def _convert_property_schema_to_type(prop_schema: dict[str, Any]) -> Any:
    """
    Convert a single JSON Schema property to a Python type annotation:
      - If 'enum' => Literal[...]
      - If 'type=object' => nested submodel
      - If 'type=array' => list[sub_type]
      - If 'type=string/integer/number/boolean' => str/int/float/bool
    """
    # If there's an enum, return a Literal of the enum values
    if "enum" in prop_schema:
        # Convert each enum value to the correct Python literal
        enum_values = prop_schema["enum"]
        return Literal[tuple(enum_values)]  # type: ignore

    # Otherwise check 'type'
    prop_type = prop_schema.get("type")

    if prop_type == "object":
        # Nested submodel
        # If 'properties' is missing, that might be an empty dict
        if "properties" in prop_schema:
            return convert_json_schema_to_basemodel(prop_schema)
        else:
            # fallback
            return dict

    if prop_type == "array":
        # Look for 'items' => sub-schema
        items_schema = prop_schema.get("items", {})
        item_type = _convert_property_schema_to_type(items_schema)
        return list[item_type]  # type: ignore

    if prop_type == "string":
        return str
    if prop_type == "boolean":
        return bool
    if prop_type == "integer":
        return int
    if prop_type == "number":
        return float

    # If the schema is "null" or unknown, fallback to object
    return object



def merge_descriptions(outer_schema: dict[str, Any], inner_schema: dict[str, Any]) -> dict[str, Any]:
    """
    Merge descriptions from outer and inner schemas, giving preference to outer.
    Also merges X-ReasoningPrompt similarly.
    """
    merged = copy.deepcopy(inner_schema)

    # Outer description preferred if present
    if outer_schema.get("description", "").strip():
        merged["description"] = outer_schema["description"]

    # Outer reasoning preferred if present
    if outer_schema.get("X-ReasoningPrompt", "").strip():
        merged["X-ReasoningPrompt"] = outer_schema["X-ReasoningPrompt"]
    elif inner_schema.get("X-ReasoningPrompt", "").strip():
        merged["X-ReasoningPrompt"] = inner_schema["X-ReasoningPrompt"]

    if not merged.get("X-ReasoningPrompt", "").strip():
        # delete it
        merged.pop("X-ReasoningPrompt", None)

    return merged

def expand_refs(schema: dict[str, Any], definitions: dict[str, dict[str, Any]] | None = None) -> dict[str, Any]:
    """
    Recursively resolve $ref in the given schema.
    For each $ref, fetch the target schema, merge descriptions, and resolve further.
    """
    if not isinstance(schema, dict):
        return schema

    # First, we will verify if this schema is expandable, we do this by checking if there are cyclic $refs (infinite loop)
    # If there are, we will return the schema as is

    if has_cyclic_refs(schema):
        print("Cyclic refs found, keeping it as is")
        return schema

    if definitions is None:
        definitions = schema.pop("$defs", {})

    assert isinstance(definitions, dict)

    if "allOf" in schema:
        # Some schemas (notably the one converted from a pydantic model) have allOf. We only accept one element in allOf
        if len(schema["allOf"]) != 1:
            raise ValueError(f"Property schema must have a single element in 'allOf'. Found: {schema['allOf']}")
        schema.update(schema.pop("allOf", [{}])[0])

    if "$ref" in schema:
        ref: str = schema["$ref"]
        if ref.startswith("#/$defs/"):
            def_name = ref.removeprefix("#/$defs/")
            if def_name not in definitions:
                raise ValueError(f"Reference {ref} not found in definitions.")
            target = definitions[def_name]
            merged = merge_descriptions(schema, target)
            merged.pop("$ref", None)
            return expand_refs(merged, definitions)
        else:
            raise ValueError(f"Unsupported reference format: {ref}")

    result: dict[str, Any] = {}
    for annotation, subschema in schema.items():
        if annotation in ["properties", "$defs"]:
            if isinstance(subschema, dict):
                new_dict = {}
                for pk, pv in subschema.items():
                    new_dict[pk] = expand_refs(pv, definitions)
                result[annotation] = new_dict
            else:
                result[annotation] = subschema
        elif annotation == "items":
            if isinstance(subschema, list):
                result[annotation] = [expand_refs(item, definitions) for item in subschema]
            else:
                result[annotation] = expand_refs(subschema, definitions)
        else:
            if isinstance(subschema, dict):
                result[annotation] = expand_refs(subschema, definitions)
            elif isinstance(subschema, list):
                new_list = []
                for item in subschema:
                    if isinstance(item, dict):
                        new_list.append(expand_refs(item, definitions))
                    else:
                        new_list.append(item)
                result[annotation] = new_list
            else:
                result[annotation] = subschema

    return result



def convert_json_schema_to_basemodel(schema: dict[str, Any]) -> Type[BaseModel]:
    """
    Create a Pydantic BaseModel dynamically from a JSON Schema:
      - Expand refs
      - For each property, figure out if it's required
      - Convert 'type': 'object' => nested submodel
      - Convert 'enum' => Literal
      - 'array' => list[submodel or primitive]
      - Primitives => str, int, float, bool
      - Preserves anyOf/oneOf structure for nullable fields
    """
    # 1) Expand references (inlines $refs)
    schema_expanded = expand_refs(copy.deepcopy(schema))

    # 2) Figure out model name
    model_name = schema_expanded.get("title", "DynamicModel")

    # 3) Collect any X-* keys for model config
    x_keys = {k: v for k, v in schema_expanded.items() if k.startswith("X-")}
    model_config = ConfigDict(extra="forbid", json_schema_extra=x_keys) if x_keys else ConfigDict(extra="forbid")

    # 4) Build up the field definitions
    properties = schema_expanded.get("properties", {})
    required_props = set(schema_expanded.get("required", []))

    field_definitions = {}
    for prop_name, prop_schema in properties.items():
        # If property is required => default=...
        # Else => default=None
        if prop_name in required_props:
            default_val = prop_schema.get("default", ...)
        else:
            default_val = prop_schema.get("default", None)

        # We also keep 'description', 'title', 'X-...' and everything else
        # that's needed to preserve schema structure for round-trip conversion
        field_kwargs = {
            "description": prop_schema.get("description"),
            "title": prop_schema.get("title"),
        }

        # Include all original schema structure for proper round-trip conversion
        schema_extra = {}
        for k, v in prop_schema.items():
            if k not in {"description", "title", "default"} and not k.startswith("$"):
                schema_extra[k] = v

        if schema_extra:
            field_kwargs["json_schema_extra"] = schema_extra

        # Handle anyOf for nullable types specially
        if "anyOf" in prop_schema:
            # Check if it's a standard nullable pattern: [type, null]
            sub_schemas = prop_schema["anyOf"]
            null_schemas = [s for s in sub_schemas if s.get("type") == "null"]
            non_null_schemas = [s for s in sub_schemas if s.get("type") != "null"]

            if len(null_schemas) == 1 and len(non_null_schemas) == 1:
                # Standard nullable field pattern
                non_null_schema = non_null_schemas[0]
                inner_type = _convert_property_schema_to_type(non_null_schema)
                python_type = Union[inner_type, None]
            else:
                # More complex anyOf structure - preserve it in schema_extra
                python_type = object

            field_definitions[prop_name] = (python_type, Field(default_val, **field_kwargs))
            continue

        # Convert to a Python type annotation
        python_type = _convert_property_schema_to_type(prop_schema)

        # If a field is not in `required`, we typically wrap it in `Optional[...]`
        if prop_name not in required_props and not is_already_optional(python_type):
            python_type = Union[python_type, None]

        field_definitions[prop_name] = (python_type, Field(default_val, **field_kwargs))

    # 5) Build the dynamic model
    return create_model(
        model_name,
        __config__=model_config,
        __module__="__main__",
        **field_definitions,
    )  # type: ignore


def is_basemodel_subclass(t: Any) -> bool:
    return isinstance(t, type) and issubclass(t, BaseModel)


def is_already_optional(t: Any) -> bool:
    """Return True if type t is Optional[...] or includes None in a Union."""
    return (get_origin(t) in {Union, types.UnionType}) and type(None) in get_args(t)


def convert_basemodel_to_partial_basemodel(base_model: Type[BaseModel]) -> Type[BaseModel]:
    """
    Convert a BaseModel class to a new BaseModel class where all fields are Optional.
    Handles nested BaseModels, lists, and unions recursively.
    """
    field_definitions: Any = {}
    maybe_optional_type: Any
    for field_name, field_info in base_model.model_fields.items():
        field_type = field_info.annotation

        # Handle nested BaseModel
        if is_basemodel_subclass(field_type):
            partial_nested = convert_basemodel_to_partial_basemodel(cast(Type[BaseModel], field_type))
            maybe_optional_type = Union[partial_nested, None]
        else:
            origin = get_origin(field_type)
            args = get_args(field_type)

            # Handle list[...] or tuple[...]
            if origin in (list, tuple) and args:
                inner_type = args[0]
                if is_basemodel_subclass(inner_type):
                    # Recursively convert the inner model
                    partial_inner = convert_basemodel_to_partial_basemodel(inner_type)
                    container_type = list if origin is list else tuple
                    new_type = container_type[partial_inner]  # type: ignore
                else:
                    new_type = field_type  # type: ignore
                maybe_optional_type = Union[new_type, None]  # type: ignore

            # Handle Union types
            elif origin in {Union, types.UnionType}:
                new_union_args: list[type] = []
                for arg in args:
                    if is_basemodel_subclass(arg):
                        new_union_args.append(convert_basemodel_to_partial_basemodel(arg))
                    else:
                        new_union_args.append(arg)
                # Make sure the union has None in it (to enforce optional)
                if type(None) not in new_union_args:
                    new_union_args.append(type(None))
                maybe_optional_type = Union[tuple(new_union_args)]  # type: ignore

            # Any other type - wrap in Optional unless already optional
            else:
                if is_already_optional(field_type):
                    maybe_optional_type = field_type
                else:
                    maybe_optional_type = Union[field_type, None]  # type: ignore

        field_definitions[field_name] = (cast(type, maybe_optional_type), None)

    # Dynamically create a new model
    return create_model(f"Partial{base_model.__name__}", __config__=base_model.model_config, __module__="__main__", **field_definitions)



def filter_auxiliary_fields(data: dict[str, Any], prefixes: list[str] = ["reasoning___", "quote___"]) -> dict[str, Any]:
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


def filter_auxiliary_fields_json(data: str, prefixes: list[str] = ["reasoning___", "quote___"]) -> dict[str, Any]:
    """
    Recursively filters out fields that start with any of the prefixes in `prefixes` from the input JSON data.
    """
    data_dict = json.loads(data)
    return filter_auxiliary_fields(data_dict, prefixes)


def load_json_schema(json_schema: Union[dict[str, Any], Path, str]) -> dict[str, Any]:
    """
    Load a JSON schema from either a dictionary or a file path.

    Args:
        json_schema: Either a dictionary containing the schema or a path to a JSON file

    Returns:
        dict[str, Any]: The loaded JSON schema

    Raises:
        JSONDecodeError: If the schema file contains invalid JSON
        FileNotFoundError: If the schema file doesn't exist
    """
    if isinstance(json_schema, (str, Path)):
        with open(json_schema) as f:
            return json.load(f)
    return json_schema


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
