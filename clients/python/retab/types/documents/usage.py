from pydantic import BaseModel, Field

class RetabUsage(BaseModel):
    """Usage information for document processing."""
    credits: float = Field(..., description="Credits consumed for processing")
