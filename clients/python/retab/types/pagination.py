from pydantic import BaseModel


class ListMetadata(BaseModel):
    before: str | None
    after: str | None
