from typing import Any, List
from pydantic import BaseModel


class ListMetadata(BaseModel):
    before: str | None
    after: str | None


class PaginatedList(BaseModel):
    data: List[Any]
    list_metadata: ListMetadata
