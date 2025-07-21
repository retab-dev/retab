import collections.abc
import json
import os
import types
import typing
import enum
import sys
import inspect
from datetime import datetime, date
from typing import Any, Type, get_args, get_origin, Union, Literal, is_typeddict
from typing_extensions import is_typeddict as is_typeddict_ext
import typing_extensions
from pydantic_core import PydanticUndefined
from pydantic import BaseModel, EmailStr
import PIL.Image

to_compile: list[tuple[str, Type, bool]] = []

def to_camel_case(snake_str: str) -> str:
    components = snake_str.split('_')
    return ''.join(x.title() for x in components)

names = {}
def is_named(cls: Type) -> bool:
    return (cls.__module__, cls.__name__) in names
def get_class_name(cls: Type) -> str:
    key = (cls.__module__, cls.__name__)
    if key in names:
        return names[key]
    name = cls.__name__
    parts = cls.__module__.split('.')
    while name in names.values():
        name = f"{to_camel_case(parts.pop())}{name}"
    names[key] = name
    return name


def is_base_model(field_type: Type) -> bool:
    return getattr(field_type, "__name__", None) in ["BaseModel", "GenericModel", "ConfigDict", "Generic"]

def type_to_zod(field_type: Any, put_names: bool = True, ts: bool = False) -> str:
    origin = get_origin(field_type) or field_type
    optional = False

    def make_union(args):
        return args[0] if len(args) <= 1 else "z.union([" + ", ".join(args) + "])"

    def make_ts_union(args):
        return args[0] if len(args) <= 1 else " | ".join(args)

    if isinstance(field_type, typing.ForwardRef):
        return type_to_zod(typing._eval_type(field_type, globals(), locals(), []), ts=ts)
    elif origin is typing.Annotated or origin is typing.Required or origin is typing_extensions.Required:
        return type_to_zod(get_args(field_type)[0], put_names, ts=ts)
    if origin is Union or origin is types.UnionType:
        args = [x for x in get_args(field_type)]
        if types.NoneType in args:
            args.remove(types.NoneType)
            optional = True
        typename = make_union([type_to_zod(x) for x in args])
        ts_typename = make_ts_union([type_to_zod(x, ts=True) for x in args])
    elif issubclass(origin, BaseModel) or is_typeddict(origin) or is_typeddict_ext(origin):
        if put_names:
            name = get_class_name(origin)
            typename = "Z" + name
            ts_typename = name
            to_compile.append((name, field_type, True))
        else:
            typename = "("
            ts_typename = ""
            based = origin.__bases__
            for i in range(0, len(based)):
                if is_base_model(based[i]) or based[i] is dict:
                    break
                typename += "Z" + get_class_name(based[i]) + ".schema).merge("
                ts_typename += get_class_name(based[i]) + " & "

            typename += "z.object({\n"
            ts_typename += "{\n"
            props = [(n, f.annotation, f.default) for n, f in origin.model_fields.items() if not f.exclude] if issubclass(origin, BaseModel) else \
                    [(n, f, PydanticUndefined) for n, f in origin.__annotations__.items()]

            for field_name, field, default in props:
                if field_name not in origin.__annotations__.keys():
                    continue
                ts_compiled = type_to_zod(field, ts=True)
                default_str = ""
                if default is not PydanticUndefined and default is not None:
                    if isinstance(default, BaseModel):
                        default_str = f".default({json.dumps(default.model_dump(mode="json", exclude_unset=True))})"
                    else:
                        default_str = f".default({json.dumps(default)})"
                typename += f"    {field_name}: {type_to_zod(field)}{default_str},\n"
                ts_typename += f"    {field_name}{"?" if ts_compiled.endswith(" | undefined") or default is not PydanticUndefined else ""}: {ts_compiled},\n"
            typename += "}))"
            ts_typename += "}"
    elif origin is list or origin is typing.List or origin is collections.abc.Sequence or origin is collections.abc.Iterable:
        typename = "z.array(" + type_to_zod(get_args(field_type)[0]) + ")"
        ts_typename = "Array<" + type_to_zod(get_args(field_type)[0], ts=True) + ">"
    elif origin is tuple:
        args = get_args(field_type)
        typename = "z.tuple([" + ", ".join([type_to_zod(x) for x in args]) + "])"
        ts_typename = "[" + ", ".join([type_to_zod(x, ts=True) for x in args]) + "]"
    elif origin is dict:
        if len(get_args(field_type)) == 2:
            typename = "z.record(" + type_to_zod(get_args(field_type)[0]) + ", " + type_to_zod(get_args(field_type)[1]) + ")"
            ts_typename = "{[key: " + type_to_zod(get_args(field_type)[0], ts=True) + "]: " + type_to_zod(get_args(field_type)[1], ts=True) + "}"
        else:
            typename = "z.record(z.any())"
            ts_typename = "{[key: string]: any}"
    elif origin is Literal:
        typename = make_union(["z.literal(" + json.dumps(x) + ")" for x in get_args(field_type)])
        ts_typename = make_ts_union([json.dumps(x) for x in get_args(field_type)])
    elif isinstance(field_type, typing.TypeVar):
        typename = "z.any()"
        ts_typename = "any"
    elif isinstance(field_type, type) and issubclass(field_type, enum.Enum):
        typename = "z.any()"
        ts_typename = "any"
    elif field_type is str or field_type is date or field_type is datetime:
        typename = "z.string()"
        ts_typename = "string"
    elif field_type is int or field_type is float:
        typename = "z.number()"
        ts_typename = "number"
    elif field_type is bool:
        typename = "z.boolean()"
        ts_typename = "boolean"
    elif field_type is typing.Any:
        typename = "z.any()"
        ts_typename = "any"
    elif field_type is bytes or field_type is PIL.Image.Image or field_type is typing.BinaryIO or origin is typing.IO or origin is typing_extensions.IO:
        typename = "z.instanceof(Uint8Array)"
        ts_typename = "Uint8Array"
    elif field_type is EmailStr:
        typename = "z.string().email()"
        ts_typename = "string"
    elif field_type is os.PathLike:
        typename = "z.string()"
        ts_typename = "string"
    elif field_type is object:
        typename = "z.object({}).passthrough()"
        ts_typename = "object"
    else:
        raise ValueError(f"Unsupported type: {field_type} ({origin})")
    if ts:
        return ts_typename if not optional else ts_typename + " | null | undefined"
    else:
        return typename if not optional else typename + ".nullable().optional()"
    

# SET of names of python builtin types starting with a capital
builtin_types = {
    "Any",
    "BaseModel",
    "NoneType",
    "Literal",
    "Union",
    "List",
    "Sequence",
    "ConfigDict",
    "Optional",
}

if __name__ == "__main__":
    modules = []
    for root, dirs, files in os.walk("retab/types"):
        for module in files:
            if module[-3:] != '.py':
                continue
            full_name = os.path.join(root, module[:-3]).replace(os.path.sep, '.')
            __import__(full_name, locals(), globals())
            modules.append(full_name)


    for module_name in modules:
        for name, obj in inspect.getmembers(sys.modules[module_name]):
            if name[0] != name[0].lower() and isinstance(obj, type) and name not in builtin_types:
                objname = get_class_name(obj)
                to_compile.append((objname, obj, False))

    print("import * as z from 'zod';\n")
    
    defined = set()
    while len(to_compile) > 0:
        name, model, necessary = to_compile.pop(0)
        if name in defined:
            continue
        defined.add(name)
        try:
            compiled = type_to_zod(model, False)
            compiled_ts = type_to_zod(model, False, ts=True)
        except Exception as e:
            if not necessary:
                print(f"Skipping {name} {model} due to error: {e}", file=sys.stderr)
                continue
            print(f"Error compiling {name} {model}", file=sys.stderr)
            raise e
        print("export const Z" + name + " = z.lazy(() => " + compiled + ");")
        print("export type " + name + " = z.infer<typeof Z" + name + ">;\n")

