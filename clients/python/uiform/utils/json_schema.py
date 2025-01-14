from typing import Any, Optional, Union, Literal, Callable, Type, Annotated, Optional, get_origin, get_args
from pydantic import BaseModel, BeforeValidator, Field, create_model
from pydantic.config import ConfigDict
from collections import defaultdict

from pathlib import Path
import copy
import json
import datetime

import stdnum.eu.vat  # type: ignore
import phonenumbers
import datetime
from email_validator import validate_email
import pycountry
import re

# **** Validation Functions ****

# 1) Special Objects

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
    pattern = r'^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$'
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
    # We’ll store the valid set in lower for easy comparison
    valid_packing_types = {'box', 'pallet', 'container', 'bag', 'drum', 'other'}
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
    valid_codes = {
        'B', 'B1000C', 'B/D', 'B/E', 'C', 'C5000D', 'C/D', 'C/E',
        'D', 'D/E', 'E', '-'
    }
    return v_str if v_str in valid_codes else None


def validate_un_packing_group(v: Any) -> Optional[str]:
    """
    Return a valid UN packing group (I, II, or III), else None.
    """
    if v is None:
        return None
    v_str = str(v).strip().upper()
    valid_groups = {'I', 'II', 'III'}
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
    if v_str.lower() in {'null', 'none', 'nan', ''}:
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
    if v_str.lower() in {'null', 'none', 'nan'}:  # Only treat explicit placeholders as None
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
            line = f"{indentation}{prop_name}{optional_flag}: {field_ts};"
            comment_identation = " " * (len(line.split("\n")[-1]) + 2)
            if add_field_description and "description" in prop_schema:
                desc = prop_schema["description"].replace("\n", f"\n{comment_identation}// ")
                line += f"  // {desc}"
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
        line = f"{field_indentation}{prop_name}{optional_flag}: {ts_type};"
        field_comment_identation = " " * (len(line.split("\n")[-1]) + 2)
        if add_field_description and "description" in prop_schema:
            desc = prop_schema["description"].replace("\n", f"\n{field_comment_identation}// ")
            line += f"  // {desc}"
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




def json_schema_to_structured_output_json_schema(obj: Union[dict[str, Any], list[Any]]) -> Union[dict[str, Any], list[Any]]:
    # Gets a json supported by GPT Structured Output from a pydantic Basemodel

    if isinstance(obj, dict):
        new_obj: dict[str, Any] = {}
        for key, value in obj.items():
            # Remove 'default' and 'format' fields
            if key in ['default', 'format']:
                continue
            
            # Switch 'integer' for 'number' ('integer' isn't a supported type)
            if key == 'type':
                if value == 'integer':
                    new_obj[key] = 'number'
                elif isinstance(value, list):
                    new_obj[key] = ['number' if t == 'integer' else t for t in value]
                else:
                    new_obj[key] = value

            # Remove 'allOf' field
            elif key == 'allOf':
                merged: dict[str, Any] = {}
                for subschema in value:
                    if '$ref' in subschema:
                        merged.update({'$ref': subschema['$ref']})
                    else:
                        merged.update(json_schema_to_structured_output_json_schema(subschema))
                new_obj.update(merged)

            else:
                new_obj[key] = json_schema_to_structured_output_json_schema(value)
        
        # Add 'required' and 'additionalProperties' if 'properties' is present
        if 'properties' in new_obj:
            new_obj['required'] = list(new_obj['properties'].keys())
            new_obj['additionalProperties'] = False
        
        if '$ref' in new_obj:
            return {'$ref': new_obj['$ref']}
        
        return new_obj
    elif isinstance(obj, list):
        return [json_schema_to_structured_output_json_schema(item) for item in obj]
    else:
        return obj


def clean_schema(schema: dict[str, Any], remove_custom_fields: bool = False, fields_to_remove: list[str] = ["default"]) -> dict[str, Any]:
    """
    Recursively remove all default values from a JSON schema.
    """
    schema = schema.copy()
    for key in list(schema.keys()):
        if not isinstance(key, str):
            # Make sure we're only removing keys that are strings
            continue
        lower_key = key.lower()
        if lower_key in fields_to_remove or key in fields_to_remove:
            schema.pop(key)
        if remove_custom_fields and lower_key.startswith("x-"):
            schema.pop(key)

    if "properties" in schema:
        schema["properties"] = {
            prop_key: clean_schema(prop_schema, fields_to_remove=fields_to_remove, remove_custom_fields=remove_custom_fields)
            for prop_key, prop_schema in schema["properties"].items()
        }
    if "items" in schema:
        schema["items"] = clean_schema(schema["items"], fields_to_remove=fields_to_remove, remove_custom_fields=remove_custom_fields)
    if "$defs" in schema:
        schema["$defs"] = {
            k: clean_schema(v, fields_to_remove=fields_to_remove, remove_custom_fields=remove_custom_fields)
            for k, v in schema["$defs"].items()
        }
    if "allOf" in schema:
        schema["allOf"] = [
            clean_schema(subschema, fields_to_remove=fields_to_remove, remove_custom_fields=remove_custom_fields)
            for subschema in schema["allOf"]
        ]
    if "anyOf" in schema:
        schema["anyOf"] = [
            clean_schema(subschema, fields_to_remove=fields_to_remove, remove_custom_fields=remove_custom_fields)
            for subschema in schema["anyOf"]
        ]

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

    elif node_type == "array":
        # Recurse into items if present
        if "items" in schema:
            updated_items, _ = _insert_reasoning_fields_inner(schema["items"])
            schema["items"] = updated_items

    return schema, reasoning_desc

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

def create_inference_schema(raw_schema: dict[str, Any]) -> dict[str, Any]:
    # Resolve refs first to get expanded schema
    definitions = raw_schema.get("$defs", {})
    resolved = expand_refs(copy.deepcopy(raw_schema), definitions)

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

    # Clean up $defs from inference_schema if desired (optional)
    if "$defs" in updated_schema:
        updated_schema.pop("$defs", None)

    # Replace description with X-FieldPrompt if present
    updated_schema = _rec_replace_description_with_llm_description(updated_schema)

    # Pop X-SystemPrompt if present
    updated_schema.pop("X-SystemPrompt", None)
    
    # Clean the schema (remove defaults, etc)
    updated_schema = clean_schema(updated_schema)
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

def cast_all_leaves_from_json_schema_to_type(leaf: dict[str, Any], new_type: Literal['string', 'boolean'], is_optional: bool = True) -> dict[str, Any]:
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
            new_leaf['type'] = new_type
    return new_leaf


SCHEMA_TYPES = Literal["string", "integer", "number", "boolean", "array", "object"]
# SCHEMA_STRING_DATE_FORMATS = Literal["date", "iso-date"]
# SCHEMA_STRING_TIME_FORMATS = Literal["time", "iso-time"]
# SCHEMA_STRING_DATETIME_FORMATS = Literal["datetime", "iso-datetime"]
# SCHEMA_STRING_CUSTOM_FORMATS = Literal["email", "phone-number", "vat-number"]

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


# Defaultdict that returns a no-op lambda for unknown keys, then merges known validators
# Expansive coercion functions (can evolve on time)
KNOWN_COERCIONS: dict[tuple[str | None, str | None], Callable[[Any], Any]] = defaultdict(lambda: lambda x: x) | {
    ("string", "iso-date"): validate_date,
    ("string", "iso-time"): validate_time,
    ("string", "email"): validate_email_address,
    ("string", "phone-number"): validate_phone_number,
    ("string", "vat-number"): validate_vat_number,
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

def flatten_dict(obj: Any, prefix: str = '') -> dict[str, Any]:
    items = []  # type: ignore
    if isinstance(obj, dict):
        for k, v in obj.items():
            new_key = f"{prefix}.{k}" if prefix else k
            items.extend(flatten_dict(v, new_key).items())
    elif isinstance(obj, list):
        for i, v in enumerate(obj):
            new_key = f"{prefix}.{i}"
            items.extend(flatten_dict(v, new_key).items())
    else:
        items.append((prefix, obj))
    return dict(items)


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
            python_validator = KNOWN_COERCIONS.get((prop_type, prop_format), None)
            python_type = get_pydantic_primitive_field_type(prop_type, prop_format, is_nullable=is_nullable, validator_func=python_validator, enum_values=enum_values)

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



def convert_basemodel_to_partial_basemodel(base_model: Type[BaseModel]) -> Type[BaseModel]:
    # Prepare the fields for the new model
    fields = {}
    for field_name, field_info in base_model.model_fields.items():
        field_type = field_info.annotation

        # Check if the field type is a Pydantic model (BaseModel subclass)
        if isinstance(field_type, type) and issubclass(field_type, BaseModel):
            # Recursively make nested models optional
            optional_field_type = Optional[convert_basemodel_to_partial_basemodel(field_type)]
        else:
            # Handle lists of nested models or other complex types
            origin = get_origin(field_type)
            if origin in (list, tuple):
                inner_type = get_args(field_type)[0]
                # Check if the inner type is a BaseModel subclass and apply recursively if so
                if isinstance(inner_type, type) and issubclass(inner_type, BaseModel):
                    optional_inner_type = Optional[origin[convert_basemodel_to_partial_basemodel(inner_type)]]
                    optional_field_type = optional_inner_type
                else:
                    optional_field_type = Optional[field_type]
            else:
                # Make the field optional if it's not already
                optional_field_type = Optional[field_type] if origin is not Optional else field_type

        # Assign field type with default None
        fields[field_name] = (optional_field_type, None)

    # Create the new model with the optional fields
    optional_model = create_model(      # type: ignore
        f"Partial{base_model.__name__}",
        **fields,
        __base__=BaseModel
    )
    return optional_model


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
