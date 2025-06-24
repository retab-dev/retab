import datetime
from typing import Any, Optional
from pydantic import BaseModel, Field
from .ai_models import Amount


class PredictionMetadata(BaseModel):
    extraction_id: Optional[str] = Field(default=None, description="The ID of the extraction")
    likelihoods: Optional[dict[str, Any]] = Field(default=None, description="The likelihoods of the extraction")
    field_locations: Optional[dict[str, Any]] = Field(default=None, description="The field locations of the extraction")
    agentic_field_locations: Optional[dict[str, Any]] = Field(default=None, description="The field locations of the extraction extracted by an llm")
    consensus_details: Optional[list[dict[str, Any]]] = Field(default=None, description="The consensus details of the extraction")
    api_cost: Optional[Amount] = Field(default=None, description="The cost of the API call for this document (if any -- ground truth for example)")


class PredictionData(BaseModel):
    prediction: dict[str, Any] = Field(default={}, description="The result of the extraction or manual annotation")
    metadata: Optional[PredictionMetadata] = Field(default=None, description="The metadata of the prediction")
    updated_at: Optional[datetime.datetime] = Field(default=None, description="The creation date of the prediction")
