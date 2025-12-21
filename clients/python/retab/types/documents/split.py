from pydantic import BaseModel, Field
from ..mime import MIMEData


class Category(BaseModel):
    name: str = Field(..., description="The name of the category")
    description: str = Field(..., description="The description of the category")


class SplitRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to split")
    categories: list[Category] = Field(..., description="The categories to split the document into")
    model: str = Field(default="retab-small", description="The model to use to split the document")


class SplitResult(BaseModel):
    name: str = Field(..., description="The name of the category")
    start_page: int = Field(..., description="The start page of the category (1-indexed)")
    end_page: int = Field(..., description="The end page of the category (1-indexed, inclusive)")


class SplitResponse(BaseModel):
    splits: list[SplitResult] = Field(..., description="The list of document splits with their page ranges")



class SplitOutputSchema(BaseModel):
    """Schema for LLM structured output."""
    splits: list[SplitResult] = Field(
        ..., 
        description="List of document sections, each classified into one of the provided categories with their page ranges"
    )