from pydantic import BaseModel, Field
from ..mime import MIMEData
from .split import Category


class ClassifyRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to classify")
    categories: list[Category] = Field(..., description="The categories to classify the document into")
    model: str = Field(default="retab-small", description="The model to use for classification")


class ClassifyResult(BaseModel):
    reasoning: str = Field(..., description="The reasoning for the classification decision")
    classification: str = Field(..., description="The category name that the document belongs to")


class ClassifyResponse(BaseModel):
    result: ClassifyResult = Field(..., description="The classification result with reasoning")


class ClassifyOutputSchema(BaseModel):
    """Schema for LLM structured output."""
    reasoning: str = Field(
        ..., 
        description="Step-by-step reasoning explaining why this document belongs to the chosen category"
    )
    classification: str = Field(
        ...,
        description="The category name that this document belongs to"
    )

