from typing import Any, Type, Optional, Literal, Annotated, Callable
import datetime
from pydantic import BaseModel, Field, ConfigDict, BeforeValidator
from pydantic import create_model
import copy
SCHEMA_TYPES = Literal["string", "integer", "number", "boolean", "array", "object"]

def merge_descriptions(outer_schema: dict[str, Any], inner_schema: dict[str, Any]) -> dict[str, Any]:
    """
    Merge descriptions from outer and inner schemas, giving preference to outer.
    Also merges X-ReasoningPrompt similarly.
    """
    merged = copy.deepcopy(inner_schema)

    # Outer description preferred if present
    if "description" in outer_schema:
        merged["description"] = outer_schema["description"]

    # Outer reasoning preferred if present
    if "X-ReasoningPrompt" in outer_schema:
        merged["X-ReasoningPrompt"] = outer_schema["X-ReasoningPrompt"]
    elif "X-ReasoningPrompt" in inner_schema:
        merged["X-ReasoningPrompt"] = inner_schema["X-ReasoningPrompt"]
    
    # Outer LLM Description preferred if present
    if "X-FieldPrompt" in outer_schema:
        merged["X-FieldPrompt"] = outer_schema["X-FieldPrompt"]
    elif "X-FieldPrompt" in inner_schema:
        merged["X-FieldPrompt"] = inner_schema["X-FieldPrompt"]

    return merged


def get_pydantic_primitive_field_type(type_: SCHEMA_TYPES | str, format_: str | None, is_nullable: bool = False, validator_func: Callable | None = None, enum_values: list[Any] | None = None) -> Any:
    python_base_type: Any
    
    if enum_values is not None:
        python_base_type = Literal[tuple(enum_values)]
    elif type_ == "string":
        if format_ in ("date", "iso-date"):
            python_base_type = datetime.date
        if format_ in ("time", "iso-time"):
            python_base_type = datetime.time
        if format_ in ("datetime", "iso-datetime"):
            python_base_type = datetime.datetime
        else:
            python_base_type = str
    elif type_ == "integer":
        python_base_type = int
    elif type_ == "number":
        python_base_type = float
    elif type_ == "boolean":
        python_base_type = bool
    elif type_ == "array":
        python_base_type = list
    elif type_ == "object":
        python_base_type = dict
    else:
        raise ValueError(f"Unsupported schema type: {type_}")
    
    field_kwargs: Any = {
        "json_schema_extra": {"format": format_}
    } if format_ is not None else {}

    final_type: Any = Annotated[python_base_type, Field(..., **field_kwargs)]
    final_type = Optional[final_type] if is_nullable or validator_func is not None else final_type
    if validator_func is not None:
        return Annotated[final_type, BeforeValidator(validator_func)]
    return final_type


def expand_refs(schema: dict[str, Any], definitions: dict[str, dict[str, Any]] | None = None) -> dict[str, Any]:
    """
    Recursively resolve $ref in the given schema.
    For each $ref, fetch the target schema, merge descriptions, and resolve further.
    """
    if not isinstance(schema, dict):
        return schema
    
    if definitions is None:
        definitions = schema.pop("$defs", {})

    assert isinstance(definitions, dict)

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
    for k, v in schema.items():
        if k in ["properties", "$defs"]:
            if isinstance(v, dict):
                new_dict = {}
                for pk, pv in v.items():
                    new_dict[pk] = expand_refs(pv, definitions)
                result[k] = new_dict
            else:
                result[k] = v
        elif k == "items":
            if isinstance(v, list):
                result[k] = [expand_refs(item, definitions) for item in v]
            else:
                result[k] = expand_refs(v, definitions)
        else:
            if isinstance(v, dict):
                result[k] = expand_refs(v, definitions)
            elif isinstance(v, list):
                new_list = []
                for item in v:
                    if isinstance(item, dict):
                        new_list.append(expand_refs(item, definitions))
                    else:
                        new_list.append(item)
                result[k] = new_list
            else:
                result[k] = v

    return result


def extract_property_type_info(prop_schema: dict[str, Any]) -> tuple[str, Optional[str], bool, list[Any] | None]:
    """
    Extract the property type, possible 'format'/'X-format', and nullability from a property schema.
    - If an 'anyOf' with exactly one 'null' type is used, we unify it into a single schema 
      (i.e., prop_schema plus is_nullable=True).
    - This ensures 'enum', 'format', etc. are preserved from the non-null sub-schema.

    Returns:
        (prop_type, prop_format, is_nullable)
    """
    is_nullable = False

    if "anyOf" in prop_schema:
        sub_schemas = prop_schema["anyOf"]
        sub_types = [s.get("type") for s in sub_schemas if isinstance(s, dict)]

        # We only handle the scenario: anyOf: [{type=XYZ,...}, {type=null}]
        # If you have more complex unions, you'll need additional logic.
        if len(sub_schemas) == 2 and "null" in sub_types:
            # Identify the non-null sub-schema
            valid_sub = next(s for s in sub_schemas if s.get("type") != "null")
            is_nullable = True

            # Merge *everything* (enum, format, x-, etc.) from the valid_sub
            # into prop_schema.  This ensures we don't lose 'enum', 'format', etc.
            prop_schema.update(valid_sub)
            # Remove the anyOf now that it's merged
            prop_schema.pop("anyOf", None)
        else:
            raise ValueError(
                f"'anyOf' structure not supported or doesn't match a single null type. Found: {sub_schemas}"
            )
    
    # At this point, we expect a single 'type' in the property
    if "type" not in prop_schema:
        raise ValueError(
            "Property schema must have a 'type' or a supported 'anyOf' pattern."
        )

    prop_type = prop_schema["type"]
    # Pop 'format' or 'X-format' if any
    prop_format = prop_schema.pop("format", None) or prop_schema.pop("X-format", None)
    enum_values = prop_schema.get("enum", None)

    return prop_type, prop_format, is_nullable, enum_values

def convert_json_schema_to_basemodel(schema: dict[str, Any]) -> Type[BaseModel]:
    """
    Create a Pydantic BaseModel dynamically from a JSON Schema.
    Steps:
      1. Expand all refs.
      2. For each property, parse type info and create a suitable Pydantic field.
      3. Nested objects -> submodels, arrays of objects -> list[submodels].
      4. Keep 'enum' and 'format' in the final schema so Pydantic sees them in the
         generated model's JSON schema.
    """
    # 1. Expand references
    schema_expanded = expand_refs(copy.deepcopy(schema))

    # 2. Gather 'X-*' keys from the root for the config
    x_keys = {k: v for k, v in schema_expanded.items() if k.startswith("X-")}

    # 3. Prepare dynamic model fields
    field_definitions: Any = {}

    # 4. Get properties + required
    props = schema_expanded.get("properties", {})
    required_fields = set(schema_expanded.get("required", []))

    for prop_name, prop_schema in props.items():
        # a) Determine the python type, format, and nullability
        prop_type, prop_format, is_nullable, enum_values = extract_property_type_info(prop_schema)
        field_kwargs = {
            "description": prop_schema.get("description"),
            "title": prop_schema.get("title"),
            # Put all schema extras, including 'enum', 'format', 'X-...' etc. into json_schema_extra
            "json_schema_extra": {
                k: v for k, v in prop_schema.items()
                if k.startswith("X-")
            },
        }

        # c) Determine the default or whether it's required
        if prop_name in required_fields:
            default_val = prop_schema.get("default", ...)
        else:
            default_val = prop_schema.get("default", None)

        # d) Dispatch based on prop_type
        if prop_type == "object":
            if "properties" not in prop_schema:
                raise ValueError(
                    f"Schema for object '{prop_name}' must have 'properties' to build a submodel."
                )
            sub_model = convert_json_schema_to_basemodel(prop_schema)
            final_type = sub_model if not is_nullable else Optional[sub_model]

            field_definitions[prop_name] = (final_type, Field(default_val, **field_kwargs))

        elif prop_type == "array":
            # We only handle "array of objects" for simplicity
            items_schema = prop_schema.get("items", {})
            item_type, _, item_nullable, _ = extract_property_type_info(items_schema)

            if item_type != "object":
                raise ValueError(
                    f"Only arrays of type 'object' are currently supported. Got: {item_type}"
                )
            if item_nullable:
                raise ValueError(
                    "Array of nullable objects is not supported by this function."
                )

            # create sub-model for items
            sub_model = convert_json_schema_to_basemodel(items_schema)

            field_definitions[prop_name] = (
                list[sub_model] if not is_nullable else Optional[list[sub_model]], 
                Field(default_val, **field_kwargs)
            )

        else:
            # e) Primitive
            python_type = get_pydantic_primitive_field_type(prop_type, prop_format, is_nullable=is_nullable, enum_values=enum_values)

            # If the field can be null, or we have a validator that must accept None:
            field_definitions[prop_name] = (python_type, Field(default_val, **field_kwargs))

    # 5. Build the model class
    model_name: str = schema_expanded.get("title", "DynamicModel")
    model_config = ConfigDict(json_schema_extra=x_keys) if x_keys else None

    return create_model(
        model_name,
        __config__=model_config,
        __module__="__main__",
        **field_definitions,
    )

