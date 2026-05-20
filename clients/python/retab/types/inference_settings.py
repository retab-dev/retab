from pydantic import Field, ConfigDict
from retab.types.base import RetabBaseModel
from typing import Any


def compute_temperature(n_consensus: int) -> float:
    """Compute temperature from n_consensus. 0.0 for single extraction, 1.0 for consensus."""
    return 1.0 if n_consensus > 1 else 0.0


class InferenceSettings(RetabBaseModel):
    model: str = "retab-small"
    image_resolution_dpi: int = Field(default=192, description="Resolution of the image sent to the LLM", ge=96, le=300)
    n_consensus: int = Field(default=1, ge=1, le=8, description="Number of consensus rounds to perform")
    model_config = ConfigDict(extra="ignore")


class ExtractionSettings(InferenceSettings):
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
