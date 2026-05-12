from pydantic import Field
from retab.types.base import RetabBaseModel

class RetabUsage(RetabBaseModel):
    """Usage information for document processing."""
    credits: float = Field(..., description="Credits consumed for processing")
