from pydantic import BaseModel, Field
from ..mime import MIMEData


class Category(BaseModel):
    name: str = Field(..., description="The name of the category")
    description: str = Field(..., description="The description of the category")
    partition_key: str | None = Field(default=None, description="The key to partition the category")


class SplitRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to split")
    categories: list[Category] = Field(..., description="The categories to split the document into")
    model: str = Field(default="retab-small", description="The model to use to split the document")


class Partition(BaseModel):
    key: str = Field(..., description="The partition key value (e.g., property ID, invoice number)")
    pages: list[int] = Field(..., description="The pages of the partition (1-indexed)")
    first_page_y_start: float = Field(default=0.0, description="The y coordinate of the first page of the partition")
    last_page_y_end: float = Field(default=1.0, description="The y coordinate of the last page of the partition")

class SplitResult(BaseModel):
    name: str = Field(..., description="The name of the category")
    pages: list[int] = Field(..., description="The pages of the category (1-indexed)")
    partitions: list[Partition] = Field(default_factory=list, description="The partitions of the category")


class SplitResponse(BaseModel):
    splits: list[SplitResult] = Field(..., description="The list of document splits with their page ranges")


class SplitOutputItem(BaseModel):
    """Internal schema item for LLM structured output validation."""
    name: str = Field(..., description="The name of the category")
    start_page: int = Field(..., description="The start page of the category (1-indexed)")
    end_page: int = Field(..., description="The end page of the category (1-indexed, inclusive)")


class SplitOutputSchema(BaseModel):
    """Schema for LLM structured output."""
    splits: list[SplitOutputItem] = Field(
        ...,
        description="List of document sections, each classified into one of the provided categories with their page ranges"
    )