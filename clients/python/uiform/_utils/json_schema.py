import copy
import datetime
import json
import re
import types
from collections import defaultdict
from pathlib import Path
from typing import Annotated, Any, Callable, Literal, MutableMapping, MutableSequence, Optional, Tuple, Type, Union, cast, get_args, get_origin

import phonenumbers
import pycountry
import stdnum.eu.vat  # type: ignore
from email_validator import validate_email
from pydantic import BaseModel, BeforeValidator, Field, create_model
from pydantic.config import ConfigDict

from uiform._utils.mime import generate_blake2b_hash_from_string
from uiform.types.schemas.layout import Column, FieldItem, Layout, RefObject, Row, RowList

# **** Validation Functions ****

# 1) Special Objects


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


def validate_currency(currency_code: Any) -> Optional[str]:
    """
    Return the valid currency code (ISO 4217) or None if invalid.
    """
    if currency_code is None:
        return None
    currency_code = str(currency_code).strip()  # convert to str and trim
    if not currency_code:
        return None
    try:
        if pycountry.currencies.lookup(currency_code):
            return currency_code
    except LookupError:
        pass
    return None


def validate_country_code(v: Any) -> Optional[str]:
    """
    Return the valid country code (ISO 3166) or None if invalid.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        if pycountry.countries.lookup(v_str):
            return v_str
    except LookupError:
        pass
    return None


def validate_email_regex(v: Any) -> Optional[str]:
    """
    Return the string if it matches a basic email pattern, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    pattern = r"^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$"
    if re.match(pattern, v_str):
        return v_str.lower()
    return None


def validate_vat_number(v: Any) -> Optional[str]:
    """
    Return the VAT number if valid (EU format) else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        if stdnum.eu.vat.is_valid(v_str):
            return stdnum.eu.vat.validate(v_str)
    except:
        pass
    return None


def validate_phone_number(v: Any) -> Optional[str]:
    """
    Return E.164 phone number format if valid, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        phone_number = phonenumbers.parse(v_str, "FR")  # Default region: FR
        if phonenumbers.is_valid_number(phone_number):
            return phonenumbers.format_number(phone_number, phonenumbers.PhoneNumberFormat.E164)
    except phonenumbers.NumberParseException:
        pass
    return None


def validate_email_address(v: Any) -> Optional[str]:
    """
    Return the normalized email address if valid, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        return validate_email(v_str).normalized
    except:
        return None


def validate_frenchpostcode(v: Any) -> Optional[str]:
    """
    Return a 5-digit postcode if valid, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    # Zero-pad to 5 digits
    try:
        v_str = v_str.zfill(5)
        # Optionally check numeric
        if not v_str.isdigit():
            return None
        return v_str
    except:
        return None


def validate_packing_type(v: Any) -> Optional[str]:
    """
    Return the packing type if in the known set, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip().lower()
    # We'll store the valid set in lower for easy comparison
    valid_packing_types = {"box", "pallet", "container", "bag", "drum", "other"}
    if v_str in valid_packing_types:
        return v_str
    return None


def validate_un_code(v: Any) -> Optional[int]:
    """
    Return an integer UN code in range [0..3481], else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        val = int(float(v_str))  # handle numeric strings
        if 0 <= val <= 3481:
            return val
    except:
        pass
    return None


def validate_adr_tunnel_code(v: Any) -> Optional[str]:
    """
    Return a valid ADR tunnel code from a known set, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip().upper()  # unify for set comparison
    valid_codes = {"B", "B1000C", "B/D", "B/E", "C", "C5000D", "C/D", "C/E", "D", "D/E", "E", "-"}
    return v_str if v_str in valid_codes else None


def validate_un_packing_group(v: Any) -> Optional[str]:
    """
    Return a valid UN packing group (I, II, or III), else None.
    """
    if v is None:
        return None
    v_str = str(v).strip().upper()
    valid_groups = {"I", "II", "III"}
    return v_str if v_str in valid_groups else None


# 2) General Objects


def validate_integer(v: Any) -> Optional[int]:
    """
    Return an integer if parseable, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        return int(float(v_str))
    except:
        return None


def validate_float(v: Any) -> Optional[float]:
    """
    Return a float if parseable, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        return float(v_str)
    except:
        return None


def validate_date(v: Union[str, datetime.date, None]) -> Optional[str]:
    """
    Return date in ISO format (YYYY-MM-DD) if valid, else None.
    """
    if v is None:
        return None

    # If it's already a date object
    if isinstance(v, datetime.date):
        return v.isoformat()

    # If it's a string
    v_str = str(v).strip()
    if not v_str:
        return None

    # Try ISO or a close variant
    try:
        return datetime.date.fromisoformat(v_str).isoformat()
    except ValueError:
        # Fallback to strptime
        try:
            return datetime.datetime.strptime(v_str, "%Y-%m-%d").date().isoformat()
        except ValueError:
            return None


def validate_time(v: Union[str, datetime.time, None]) -> Optional[str]:
    """
    Return time in ISO format (HH:MM[:SS]) if valid, else None.
    """
    if v is None:
        return None

    # If it's already a time object
    if isinstance(v, datetime.time):
        return v.isoformat()

    v_str = str(v).strip()
    if not v_str:
        return None

    # Try multiple formats
    time_formats = ["%H:%M:%S", "%H:%M", "%I:%M %p", "%I:%M:%S %p"]
    for fmt in time_formats:
        try:
            parsed = datetime.datetime.strptime(v_str, fmt).time()
            return parsed.isoformat()
        except ValueError:
            continue
    return None


def validate_bool(v: Any) -> bool:
    """
    Convert to bool if matches known true/false strings or actual bool.
    Otherwise return False.
    """
    if v is None:
        return False

    if isinstance(v, bool):
        return v

    try:
        v_str = str(v).strip().lower()
        true_values = {"true", "t", "yes", "y", "1"}
        false_values = {"false", "f", "no", "n", "0"}
        if v_str in true_values:
            return True
        elif v_str in false_values:
            return False
    except:
        pass

    return False


def validate_strold(v: Any) -> Optional[str]:
    """
    Return a stripped string unless it's empty or a known 'null' placeholder, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    # Treat these placeholders (and empty) as invalid
    if v_str.lower() in {"null", "none", "nan", ""}:
        return None
    return v_str


def validate_str(v: Any) -> Optional[str]:
    """
    Return a stripped string unless it's invalid (e.g., placeholders like 'null'), else None.
    Does NOT convert empty strings to None—leaves them as-is.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if v_str.lower() in {"null", "none", "nan"}:  # Only treat explicit placeholders as None
        return None
    return v_str  # Keep empty strings intact


def notnan(x: Any) -> bool:
    """
    Return False if x is None, 'null', 'nan', or x != x (NaN check).
    True otherwise.
    """
    if x is None:
        return False
    x_str = str(x).lower().strip()
    if x_str in {"null", "nan"}:
        return False
    # Check for actual float NaN (x != x)
    return not (x != x)


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

    # Outer LLM Description preferred if present
    if outer_schema.get("X-FieldPrompt", "").strip():
        merged["X-FieldPrompt"] = outer_schema["X-FieldPrompt"]
    elif inner_schema.get("X-FieldPrompt", "").strip():
        merged["X-FieldPrompt"] = inner_schema["X-FieldPrompt"]

    if not merged.get("X-FieldPrompt", "").strip():
        # delete it
        merged.pop("X-FieldPrompt", None)

    # System-Prompt
    if not merged.get("X-SystemPrompt", "").strip():
        # delete it
        merged.pop("X-SystemPrompt", None)

    return merged


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
    Inner function that processes a schema and adds quote___ fields for leaf nodes with X-ReferenceQuote: true.
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
            has_quote_field = updated_prop_schema_value.get("X-ReferenceQuote") is True

            # Check if this property is a leaf with X-ReferenceQuote: true
            if has_quote_field:
                # Add the quote field
                quote_key = f"quote___{property_key}"
                new_props[quote_key] = {"type": "string"}

                # Add the quote field to required if the property is required
                if "required" in new_schema and property_key in new_schema["required"]:
                    # add the quote field to required just before the property_key
                    new_schema["required"].insert(new_schema["required"].index(property_key), quote_key)

                # Remove the X-ReferenceQuote field
                updated_prop_schema_value.pop("X-ReferenceQuote", None)

            new_props[property_key] = updated_prop_schema_value
        new_schema["properties"] = new_props

    elif "items" in new_schema:
        # Recurse into items if present
        updated_items = _insert_quote_fields_inner(new_schema["items"])
        new_schema["items"] = updated_items

    return new_schema


def _rec_replace_description_with_llm_description(schema: dict[str, Any]) -> dict[str, Any]:
    """
    Recursively replace the description field with X-ReasoningPrompt if present.
    """
    if not isinstance(schema, dict):
        return schema

    new_schema = copy.deepcopy(schema)
    if "description" in new_schema or "X-FieldPrompt" in new_schema:
        new_schema["description"] = new_schema.pop("X-FieldPrompt", new_schema.get("description"))
        if new_schema["description"] is None:
            new_schema.pop("description")
        elif "default" in new_schema:
            new_schema["description"] += f"\nUser Provided a Default Value: {json.dumps(new_schema['default'])}"

    if "properties" in new_schema:
        new_schema["properties"] = {k: _rec_replace_description_with_llm_description(v) for k, v in new_schema["properties"].items()}

    if "items" in new_schema:
        new_schema["items"] = _rec_replace_description_with_llm_description(new_schema["items"])

    if "$defs" in new_schema:
        new_schema["$defs"] = {k: _rec_replace_description_with_llm_description(v) for k, v in new_schema["$defs"].items()}

    return new_schema


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

    # Insert quote fields for leaf nodes with X-ReferenceQuote: true
    updated_schema = _insert_quote_fields_inner(updated_schema)

    # Clean up $defs from inference_schema if desired (optional)
    # if "$defs" in updated_schema:
    #     updated_schema.pop("$defs", None)

    # Replace description with X-FieldPrompt if present
    updated_schema = _rec_replace_description_with_llm_description(updated_schema)

    # Clean the schema (remove defaults, etc)
    updated_schema = clean_schema(updated_schema, remove_custom_fields=True)
    return updated_schema


def cleanup_reasoning(output_data: Any, reasoning_preffix: str = "reasoning___") -> Any:
    """
    Recursively removes all reasoning key/values from the output data. Reasoning keys starts with 'reasoning___'.
    """
    if isinstance(output_data, dict):
        new_dict = {}
        for k, v in output_data.items():
            if not k.startswith(reasoning_preffix):
                new_dict[k] = cleanup_reasoning(v)
        return new_dict
    elif isinstance(output_data, list):
        return [cleanup_reasoning(item) for item in output_data]
    else:
        return output_data


# Other utils


def cast_all_leaves_from_json_schema_to_type(leaf: dict[str, Any], new_type: Literal["string", "boolean"], is_optional: bool = True) -> dict[str, Any]:
    new_leaf: dict[str, Any] = {}
    # new_leaf["description"] = "Here goes the suggestion, if any, or null."
    if leaf.get("type") == "object":
        new_leaf["type"] = "object"
        new_leaf["properties"] = {}
        for key, value in leaf["properties"].items():
            new_leaf["properties"][key] = cast_all_leaves_from_json_schema_to_type(value, new_type, is_optional=is_optional)
    elif leaf.get("type") == "array":
        new_leaf["type"] = "array"
        new_leaf["items"] = cast_all_leaves_from_json_schema_to_type(leaf["items"], new_type, is_optional=is_optional)
    else:
        if is_optional:
            new_leaf["anyOf"] = [{"type": new_type}, {"type": "null"}]
        else:
            new_leaf["type"] = new_type
    return new_leaf


SCHEMA_TYPES = Literal["string", "integer", "number", "boolean", "array", "object"]
# SCHEMA_STRING_DATE_FORMATS = Literal["date", "iso-date"]
# SCHEMA_STRING_TIME_FORMATS = Literal["time", "iso-time"]
# SCHEMA_STRING_DATETIME_FORMATS = Literal["datetime", "iso-datetime"]
# SCHEMA_STRING_CUSTOM_FORMATS = Literal["email", "phone-number", "vat-number"]


def get_pydantic_primitive_field_type(
    type_: SCHEMA_TYPES | str, format_: str | None, is_nullable: bool = False, validator_func: Callable | None = None, enum_values: list[Any] | None = None
) -> Any:
    python_base_type: Any

    if enum_values is not None:
        python_base_type = Literal[tuple(enum_values)]  # type: ignore
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

    field_kwargs: Any = {"json_schema_extra": {"format": format_}} if format_ is not None else {}

    final_type: Any = Annotated[python_base_type, Field(..., **field_kwargs)]
    final_type = Optional[final_type] if is_nullable or validator_func is not None else final_type
    if validator_func is not None:
        return Annotated[final_type, BeforeValidator(validator_func)]
    return final_type


# Defaultdict that returns a no-op lambda for unknown keys, then merges known validators
# Expansive coercion functions (can evolve on time)
KNOWN_COERCIONS: dict[tuple[str | None, str | None], Callable[[Any], Any]] = defaultdict(lambda: lambda x: x) | {
    # ("string", "iso-date"): validate_date,
    # ("string", "iso-time"): validate_time,
    # ("string", "email"): validate_email_address,
    # ("string", "phone-number"): validate_phone_number,
    # ("string", "vat-number"): validate_vat_number,
    ("integer", None): validate_integer,
    ("number", None): validate_float,
    ("boolean", None): validate_bool,
    ("string", None): validate_str,
}


def object_format_coercion(instance: dict[str, Any], schema: dict[str, Any]) -> dict[str, Any]:
    """
    Coerces an instance to conform to a JSON Schema, applying defaults and handling nullable fields.
    Converts empty strings to None only if the field is optional.
    """

    def recursive_coercion(_instance: Any, _schema: dict[str, Any]) -> Any:
        # 1. Handle object type
        if _schema.get("type") == "object":
            if not isinstance(_instance, dict):
                return _schema.get("default", {})
            coerced_instance = {}
            for prop_key, prop_schema in _schema.get("properties", {}).items():
                coerced_instance[prop_key] = recursive_coercion(_instance.get(prop_key), prop_schema)
            return coerced_instance

        # 2. Handle array type
        if _schema.get("type") == "array":
            if not isinstance(_instance, list):
                return _schema.get("default", [])
            return [recursive_coercion(value, _schema.get("items", {})) for value in _instance]

        # 3. Handle anyOf (optional fields)
        if "anyOf" in _schema:
            is_field_optional = any(sub.get("type") == "null" for sub in _schema["anyOf"])
            if is_field_optional and (_instance == "" or _instance is None):
                return None

            # Try to coerce with the first matching subschema
            for subschema in _schema["anyOf"]:
                # Skip null subschema for explicit coercion; handled above
                if subschema.get("type") == "null":
                    continue
                coerced_value = recursive_coercion(_instance, subschema)
                if coerced_value is not None:
                    return coerced_value
            return None  # If none match, return None

        # 4. Handle primitive types and known coercions
        schema_type = _schema.get("type")
        ## Custom Formats that are not supported by default should be supplied as X-format.
        schema_format = _schema.get("X-format") or _schema.get("format")

        # Use default if instance is None
        if _instance is None:
            _instance = _schema.get("default")

        # If schema type is null, just return None
        if schema_type == "null":
            return None

        # Apply known coercion
        if (schema_type, schema_format) in KNOWN_COERCIONS:
            return KNOWN_COERCIONS[(schema_type, schema_format)](_instance)

        return _instance  # Return as-is if no coercion is required

    expanded_schema = expand_refs(schema)
    coerced = recursive_coercion(instance, expanded_schema)
    return coerced if coerced is not None else {}


def flatten_dict(obj: Any, prefix: str = "", allow_empty_objects: bool = True) -> dict[str, Any]:
    items = []  # type: ignore
    if isinstance(obj, dict):
        if len(obj) == 0 and allow_empty_objects:
            # Keep empty dicts as dicts (so we can keep its structure)
            items.append((prefix, {}))
        else:
            for k, v in obj.items():
                new_key = f"{prefix}.{k}" if prefix else k
                items.extend(flatten_dict(v, new_key, allow_empty_objects=allow_empty_objects).items())

    elif isinstance(obj, list):
        if len(obj) == 0 and allow_empty_objects:
            # Keep empty lists as lists (so we can keep its structure)
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
            raise ValueError(f"'anyOf' structure not supported or doesn't match a single null type. Found: {sub_schemas}")

    # At this point, we expect a single 'type' in the property
    if "type" not in prop_schema:
        raise ValueError("Property schema must have a 'type' or a supported 'anyOf' pattern.")

    prop_type = prop_schema["type"]
    # Pop 'format' or 'X-format' if any
    prop_format = prop_schema.pop("format", None) or prop_schema.pop("X-format", None)
    enum_values = prop_schema.get("enum", None)

    return prop_type, prop_format, is_nullable, enum_values


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


def convert_json_schema_to_basemodelold(schema: dict[str, Any]) -> Type[BaseModel]:
    """
    Create a Pydantic BaseModel dynamically from a JSON Schema.
    Steps:
      1. Expand all refs.
      2. For each property, parse type info and create a suitable Pydantic field.
      3. Nested objects -> submodels, arrays -> list[type].
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
            "json_schema_extra": {k: v for k, v in prop_schema.items() if k.startswith("X-")},
        }

        # c) Determine the default or whether it's required
        if prop_name in required_fields:
            default_val = prop_schema.get("default", ...)
        else:
            default_val = prop_schema.get("default", None)

        # d) Dispatch based on prop_type
        if prop_type == "object":
            if "properties" not in prop_schema:
                raise ValueError(f"Schema for object '{prop_name}' must have 'properties' to build a submodel.")
            sub_model = convert_json_schema_to_basemodel(prop_schema)
            final_type = sub_model if not is_nullable else Optional[sub_model]

            field_definitions[prop_name] = (final_type, Field(default_val, **field_kwargs))

        elif prop_type == "array":
            # Handle arrays of both objects and primitive types
            items_schema = prop_schema.get("items", {})
            item_type, item_format, item_nullable, item_enum = extract_property_type_info(items_schema)

            if item_type == "object":
                # Handle array of objects
                sub_model = convert_json_schema_to_basemodel(items_schema)
                array_type = list[sub_model]  # type: ignore
            else:
                # Handle array of primitives
                item_python_type = get_pydantic_primitive_field_type(
                    item_type, item_format, is_nullable=item_nullable, validator_func=KNOWN_COERCIONS.get((item_type, item_format), None), enum_values=item_enum
                )
                array_type = list[item_python_type]  # type: ignore

            field_definitions[prop_name] = (array_type if not is_nullable else Optional[array_type], Field(default_val, **field_kwargs))

        else:
            # e) Primitive
            python_validator = KNOWN_COERCIONS.get((prop_type, prop_format), None)
            python_type = get_pydantic_primitive_field_type(prop_type, prop_format, is_nullable=is_nullable, validator_func=python_validator, enum_values=enum_values)

            # If the field can be null, or we have a validator that must accept None:
            field_definitions[prop_name] = (python_type, Field(default_val, **field_kwargs))

    # 5. Build the model class
    model_name: str = schema_expanded.get("title", "DynamicModel")
    model_config = ConfigDict(extra="forbid", json_schema_extra=x_keys) if x_keys else ConfigDict(extra="forbid")

    return create_model(
        model_name,
        __config__=model_config,
        __module__="__main__",
        **field_definitions,
    )


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


def filter_auxiliary_fields(data: dict[str, Any], prefixes: list[str] = ["reasoning___", "quote___"]) -> dict[str, Any]:
    """
    Recursively filters out fields that start with any of the prefixes in `prefixes` from the input data.
    """
    if not isinstance(data, dict):
        return data  # Base case: return non-dict values as is

    filtered: dict[str, Any] = {}
    for key, value in data.items():
        if not key.startswith(tuple(prefixes)):
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


def get_all_paths(schema: dict[str, Any]) -> list[str]:
    """
    Extract all possible JSON pointer paths from a JSON Schema.

    This function traverses a JSON Schema and generates a list of all possible paths
    that could exist in a document conforming to that schema. For arrays, it uses '*'
    as a wildcard index.

    Args:
        schema (dict[str, Any]): The JSON Schema to analyze

    Returns:
        list[str]: A list of dot-notation paths (e.g. ["person.name", "person.addresses.*.street"])

    Example:
        >>> schema = {
        ...     "type": "object",
        ...     "properties": {
        ...         "name": {"type": "string"},
        ...         "addresses": {
        ...             "type": "array",
        ...             "items": {
        ...                 "type": "object",
        ...                 "properties": {
        ...                     "street": {"type": "string"}
        ...                 }
        ...             }
        ...         }
        ...     }
        ... }
        >>> get_all_paths(schema)
        ['name', 'addresses', 'addresses.*.street']
    """
    paths: list[str] = []

    def _traverse(current_schema: dict[str, Any], current_path: str = "") -> None:
        if any(key in current_schema for key in ["oneOf", "allOf"]):
            raise ValueError("OneOf and AllOf are not supported yet.")

        # Handle array type schemas
        # if current_schema.get("type") == "array":
        if "items" in current_schema:
            paths.append(f"{current_path}")
            _traverse(current_schema["items"], f"{current_path}.*")
            return

        # Handle object type schemas
        if "properties" in current_schema:
            for prop_name, prop_schema in current_schema["properties"].items():
                new_path = f"{current_path}.{prop_name}" if current_path else prop_name

                # If property is a leaf node (has type but no properties/items)
                if not any(key in prop_schema for key in ["properties", "items"]):
                    paths.append(new_path)
                else:
                    _traverse(prop_schema, new_path)

        # Handle $ref schemas
        elif "$ref" in current_schema:
            # Skip refs for now since we don't have access to the full schema with definitions
            pass

        # Handle anyOf/oneOf/allOf schemas

        elif any(key in current_schema for key in ["anyOf", "oneOf", "allOf"]):
            # Take first schema as representative for path generation
            for key in ["anyOf", "oneOf", "allOf"]:
                if key in current_schema and current_schema[key]:
                    _traverse(current_schema[key][0], current_path)
                    break

    _traverse(schema)
    return paths


def convert_schema_to_layout(schema: dict[str, Any]) -> dict[str, Any]:
    """
    Convert a JSON Schema (represented as a Python dict) into a Layout object.
    """
    # Get the definitions from the schema (or empty dict if not provided)
    defs = schema.get("$defs", {})
    converted_defs: dict[str, Column] = {}

    def is_object_schema(sch: dict[str, Any]) -> bool:
        return "properties" in sch and isinstance(sch.get("properties"), dict)

    def extract_ref(sch: dict[str, Any]) -> Optional[str]:
        return sch.get("$ref")

    def extract_ref_schema(ref: Optional[str], defs: dict[str, dict[str, Any]]) -> Optional[dict[str, Any]]:
        if not ref:
            return None
        ref_name = ref.split("/")[-1]
        return defs.get(ref_name)

    def is_object_via_any_of(sch: dict[str, Any]) -> bool:
        any_of = sch.get("anyOf")
        if isinstance(any_of, list):
            return any((extract_ref(option) and extract_ref_schema(extract_ref(option), defs)) or is_object_schema(option) for option in any_of)
        return False

    def property_is_object(prop_schema: dict[str, Any]) -> bool:
        ref = extract_ref(prop_schema)
        if ref:
            ref_schema = extract_ref_schema(ref, defs)
            return bool(ref_schema)
        return is_object_schema(prop_schema) or is_object_via_any_of(prop_schema)

    def property_is_array(prop_schema: dict[str, Any]) -> bool:
        return prop_schema.get("type") == "array"

    def handle_ref_object(prop_name: str, ref: str) -> RefObject:
        ref_name = ref.split("/")[-1]
        if ref_name not in converted_defs:
            ref_schema = extract_ref_schema(ref, defs)
            if ref_schema and is_object_schema(ref_schema):
                result = handle_object(ref_name, ref_schema, drop_name=True)
                assert isinstance(result, Column)
                converted_defs[ref_name] = result
        return RefObject(type="object", size=None, **{"$ref": ref})

    def handle_object(prop_name: str, object_schema: dict[str, Any], drop_name: bool = False) -> Union[RefObject, Column]:
        ref = extract_ref(object_schema)
        if ref:
            return handle_ref_object(prop_name, ref)
        else:
            props = object_schema.get("properties")
            if not props:
                # If no properties, try anyOf (skipping null types)
                any_of = object_schema.get("anyOf")
                if isinstance(any_of, list):
                    for option in any_of:
                        if option.get("type") != "null":
                            props = option.get("properties")
                            if props:
                                break
            if not props:
                props = {}
            items: list[Row | RowList | FieldItem | RefObject] = []
            for p_name, p_schema in props.items():
                if property_is_object(p_schema):
                    # Wrap object properties in a row
                    items.append(Row(type="row", name=p_name, items=[handle_object(p_name, p_schema)]))
                elif property_is_array(p_schema):
                    items.append(handle_array_items(p_name, p_schema))
                else:
                    items.append(FieldItem(type="field", name=p_name, size=1))
            if drop_name:
                return Column(type="column", size=1, items=items)
            else:
                return Column(type="column", size=1, items=items, name=prop_name)

    def handle_array_items(prop_name: str, array_schema: dict[str, Any]) -> RowList:
        items_schema = array_schema.get("items", {})
        row_items: list[Column | FieldItem | RefObject] = []
        if property_is_object(items_schema):
            row_items.append(handle_object(prop_name, items_schema))
        else:
            row_items.append(FieldItem(type="field", name=prop_name, size=1))
        return RowList(type="rowList", name=prop_name, items=row_items)

    # Process definitions from $defs
    for definition_name, definition_schema in defs.items():
        if is_object_schema(definition_schema):
            result = handle_object(definition_name, definition_schema, drop_name=True)
            assert isinstance(result, Column)
            converted_defs[definition_name] = result

    # Process top-level properties
    top_level_props = schema.get("properties", {})
    top_level_items: list[Row | RowList | FieldItem | RefObject] = []
    for prop_name, prop_schema in top_level_props.items():
        if property_is_object(prop_schema):
            top_level_items.append(Row(type="row", name=prop_name, items=[handle_object(prop_name, prop_schema)]))
        elif property_is_array(prop_schema):
            top_level_items.append(handle_array_items(prop_name, prop_schema))
        else:
            top_level_items.append(FieldItem(type="field", name=prop_name, size=1))

    return Layout(type="column", size=1, items=top_level_items, **{"$defs": converted_defs}).model_dump(by_alias=True)


### Json Schema to NLP Data Structure


def get_type_str(field_schema):
    """
    Recursively determine the type string for a given schema field.
    Handles 'anyOf' unions, enums, arrays, and simple types.
    """
    if "anyOf" in field_schema:
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
        if items_schema.get("type") == "object" and "properties" in items_schema:
            md += process_schema_field("*", items_schema, level + 1, field_name_prefix=field_name_complete + ".")

    return md


def json_schema_to_nlp_data_structure(schema: dict) -> str:
    """
    Receives a JSON schema (without $defs or $ref) and returns a markdown string
    that documents each field with its name, description, type (including unions and enums),
    and default value (if defined). Root-level fields use 3 hashtags, and nested fields
    add one hashtag per level.
    """
    schema_title = schema.get("title", schema.get("name", "Schema"))
    md = f"## {schema_title} -- NLP Data Structure\n\n"
    # Assume the root schema is an object with properties.
    if schema.get("type") == "object" and "properties" in schema:
        for field_name, field_schema in schema["properties"].items():
            md += process_schema_field(field_name, field_schema, 3)
    else:
        md += process_schema_field("root", schema, 3)
    return md


def nlp_data_structure_to_field_descriptions(nlp_data_structure: str) -> dict:
    """
    This function updates the JSON schema with the descriptions from the NLP data structure.

    Args:
        schema: The original JSON schema dictionary
        nlp_data_structure: A markdown string created by json_schema_to_nlp_data_structure, potentially with updated descriptions

    Returns:
        A new schema with updated descriptions from the NLP data structure
    """

    # Pattern to match headers and extract field_name and type
    # Example: "### field_name (type)" or "#### parent.child (type)"
    header_pattern = re.compile(r"^(#+)\s+([^\s(]+)\s*\(([^)]*)\)")

    # Pattern to extract description between tags
    description_pattern = re.compile(r"<Description>(.*?)</Description>", re.DOTALL)

    # Split the markdown by lines
    lines = nlp_data_structure.split("\n")

    # Process the markdown to extract field names and descriptions
    field_descriptions = {}

    i = 0
    while i < len(lines):
        line = lines[i]

        # Check if this line is a header
        header_match = header_pattern.match(line)
        if header_match:
            field_path = header_match.group(2)  # Field name or path

            # Look for description in subsequent lines until next header
            desc_start = i + 1
            while desc_start < len(lines) and not header_pattern.match(lines[desc_start]):
                desc_start += 1

            # Extract description from the block of text
            description_block = "\n".join(lines[i + 1 : desc_start])
            desc_match = description_pattern.search(description_block)
            if desc_match:
                description_text = desc_match.group(1).strip()
                field_descriptions[field_path] = description_text

            i = desc_start - 1  # Will be incremented in the loop

        i += 1
    return field_descriptions


##### JSON Schema Sanitization  #####

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


import copy
import json
from .mime import generate_blake2b_hash_from_string


def compute_schema_data_id(json_schema: dict[str, Any]) -> str:
    """Returns the schema_data_id for a given JSON schema.

    The schema_data_id is a hash of the schema data, ignoring all prompt/description/default fields
    and other non-structural metadata.

    Args:
        json_schema: The JSON schema to compute the ID for

    Returns:
        str: A hash string representing the schema data version with "sch_data_id_" prefix
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
