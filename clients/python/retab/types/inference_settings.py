from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, ConfigDict, field_validator
from typing import Any


def compute_temperature(n_consensus: int) -> float:
    """Compute temperature from n_consensus. 0.0 for single extraction, 1.0 for consensus."""
    return 1.0 if n_consensus > 1 else 0.0


class InferenceSettings(BaseModel):
    model: str = "retab-small"
    reasoning_effort: ChatCompletionReasoningEffort = "minimal"
    image_resolution_dpi: int = Field(default=192, description="Resolution of the image sent to the LLM", ge=96, le=300)
    n_consensus: int = Field(default=1, ge=1, le=8, description="Number of consensus rounds to perform")
    chunking_keys: dict[str, str] | None = Field(default=None, description="If set, keys to be used for the extraction of long lists of data using Parallel OCR", examples=[{"properties": "ID", "products": "identity.id"}])
    model_config = ConfigDict(extra="ignore")

    @field_validator("reasoning_effort", mode="before")
    @classmethod
    def force_minimal_reasoning_effort(cls, _: Any) -> ChatCompletionReasoningEffort:
        return "minimal"


class ExtractionSettings(InferenceSettings):
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    
