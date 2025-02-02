from pydantic import BaseModel, Field
from typing import Optional


#https://platform.openai.com/docs/guides/prompt-caching
from ...types.usage import Amount
from openai.types.completion_usage import CompletionUsage
from .openai import openai_pricing_list, compute_openai_api_call_cost

# ─── PRICING MODELS ────────────────────────────────────────────────────────────

def compute_cost_from_model(model: str, usage: CompletionUsage) -> Amount:
    pricing = next((p for p in openai_pricing_list if p.model == model), None)
    if not pricing:
        raise ValueError(f"No pricing information found for model: {model}")
    return compute_openai_api_call_cost(pricing, usage)

    


########################
# USELESS FOR NOW
########################

class CompletionsUsage(BaseModel):
    """The aggregated completions usage details of the specific time bucket."""
    
    object: str = Field(
        default="organization.usage.completions.result",
        description="Type identifier for the completions usage object"
    )
    input_tokens: int = Field(
        description="The aggregated number of text input tokens used, including cached tokens. For customers subscribe to scale tier, this includes scale tier tokens."
    )
    input_cached_tokens: int = Field(
        description="The aggregated number of text input tokens that has been cached from previous requests. For customers subscribe to scale tier, this includes scale tier tokens."
    )
    output_tokens: int = Field(
        description="The aggregated number of text output tokens used. For customers subscribe to scale tier, this includes scale tier tokens."
    )
    input_audio_tokens: int = Field(
        description="The aggregated number of audio input tokens used, including cached tokens."
    )
    output_audio_tokens: int = Field(
        description="The aggregated number of audio output tokens used."
    )
    num_model_requests: int = Field(
        description="The count of requests made to the model."
    )
    project_id: Optional[str] = Field(
        default=None,
        description="When group_by=project_id, this field provides the project ID of the grouped usage result."
    )
    user_id: Optional[str] = Field(
        default=None,
        description="When group_by=user_id, this field provides the user ID of the grouped usage result."
    )
    api_key_id: Optional[str] = Field(
        default=None,
        description="When group_by=api_key_id, this field provides the API key ID of the grouped usage result."
    )
    model: Optional[str] = Field(
        default=None,
        description="When group_by=model, this field provides the model name of the grouped usage result."
    )
    batch: Optional[bool] = Field(
        default=None,
        description="When group_by=batch, this field tells whether the grouped usage result is batch or not."
    )
