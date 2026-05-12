from typing import Dict, List, Literal, Optional, Union

from pydantic import Field as PydanticField
from retab.types.base import RetabBaseModel


# Terminal items
class FieldItem(RetabBaseModel):
    type: Literal["field"]
    name: str
    size: Optional[int] = None


class RefObject(RetabBaseModel):
    type: Literal["object"]
    size: Optional[int] = None
    name: Optional[str] = None
    ref: str = PydanticField(..., alias="$ref")


# Recursive items
class Column(RetabBaseModel):
    type: Literal["column"]
    size: int
    items: List[Union["Row", FieldItem, RefObject, "RowList"]] = PydanticField(default_factory=list)
    name: Optional[str] = None

    model_config = {"arbitrary_types_allowed": True}


class Row(RetabBaseModel):
    type: Literal["row"]
    name: Optional[str] = None
    items: List[Column | FieldItem | RefObject]

    model_config = {"arbitrary_types_allowed": True}


class RowList(RetabBaseModel):
    type: Literal["rowList"]
    name: Optional[str] = None
    items: List[Column | FieldItem | RefObject] = PydanticField(default_factory=list)

    model_config = {"arbitrary_types_allowed": True}


# Root Layout type
class Layout(RetabBaseModel):
    # Use alias "$defs" for the definitions
    defs: Dict[str, Column] = PydanticField(default_factory=dict, alias="$defs")
    type: Literal["column"]
    size: int
    items: List[Row | RowList | FieldItem | RefObject] = PydanticField(default_factory=list)

    model_config = {"arbitrary_types_allowed": True}


Column.model_rebuild()
