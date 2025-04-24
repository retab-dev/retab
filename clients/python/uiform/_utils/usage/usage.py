from typing import Optional

from openai.types.completion_usage import CompletionUsage
from pydantic import BaseModel, Field

# https://platform.openai.com/docs/guides/prompt-caching
from ...types.ai_models import Amount, Pricing, get_model_card

# ─── PRICING MODELS ────────────────────────────────────────────────────────────


def compute_api_call_cost(pricing: Pricing, usage: CompletionUsage, is_ft: bool = False) -> Amount:
    """
    Computes the price (as an Amount) for the given token usage, based on the pricing.

    Assumptions / rules:
      - The pricing rates are per 1,000,000 tokens.
      - The usage is broken out into prompt and completion tokens.
      - If details are provided, the prompt tokens are split into:
          • cached text tokens (if any),
          • audio tokens (if any), and
          • the remaining are treated as "regular" text tokens.
      - Similarly, completion tokens are split into audio tokens (if any) and the remaining text.
      - For text tokens, if a cached price is not explicitly provided in pricing.text.cached,
        we assume a 50% discount off the prompt price.
      - (A similar rule could be applied to audio tokens – i.e. 20% of the normal price –
       if cached audio tokens were provided. In our usage model, however, only text tokens
       are marked as cached.)
      - If is_ft is True, the price is adjusted using the ft_price_hike multiplier.
    """

    # ----- Process prompt tokens -----
    prompt_cached_text = 0
    prompt_audio = 0
    if usage.prompt_tokens_details:
        prompt_cached_text = usage.prompt_tokens_details.cached_tokens or 0
        prompt_audio = usage.prompt_tokens_details.audio_tokens or 0

    # The rest of the prompt tokens are assumed to be regular (non-cached) text tokens.
    prompt_regular_text = usage.prompt_tokens - prompt_cached_text - prompt_audio

    # ----- Process completion tokens -----
    completion_audio = 0
    if usage.completion_tokens_details:
        completion_audio = usage.completion_tokens_details.audio_tokens or 0

    # The rest are text tokens.
    completion_regular_text = usage.completion_tokens - completion_audio

    # ----- Calculate text token costs -----
    # Regular prompt text tokens cost at the standard prompt rate.
    cost_text_prompt = prompt_regular_text * pricing.text.prompt

    # Cached text tokens: if the pricing model provides a cached rate, use it.
    # Otherwise, apply a 50% discount (i.e. 0.5 * prompt price).
    text_cached_price = pricing.text.prompt * pricing.text.cached_discount
    cost_text_cached = prompt_cached_text * text_cached_price

    # Completion text tokens cost at the standard completion rate.
    cost_text_completion = completion_regular_text * pricing.text.completion

    total_text_cost = cost_text_prompt + cost_text_cached + cost_text_completion

    # ----- Calculate audio token costs (if any) -----
    total_audio_cost = 0.0
    if pricing.audio:
        cost_audio_prompt = prompt_audio * pricing.audio.prompt
        cost_audio_completion = completion_audio * pricing.audio.completion
        total_audio_cost = cost_audio_prompt + cost_audio_completion

    total_cost = (total_text_cost + total_audio_cost) / 1e6

    # Apply fine-tuning price hike if applicable
    if is_ft and hasattr(pricing, 'ft_price_hike'):
        total_cost *= pricing.ft_price_hike

    return Amount(value=total_cost, currency="USD")


def compute_cost_from_model(model: str, usage: CompletionUsage) -> Amount:
    # Extract base model name for fine-tuned models like "ft:gpt-4o:uiform:4389573"
    is_ft = False
    if model.startswith("ft:"):
        # Split by colon and take the second part (index 1) which contains the base model
        parts = model.split(":")
        if len(parts) > 1:
            model = parts[1]
            is_ft = True

    # Use the get_model_card function from types.ai_models
    try:
        model_card = get_model_card(model)
        pricing = model_card.pricing
    except ValueError as e:
        raise ValueError(f"No pricing information found for model: {model}")

    return compute_api_call_cost(pricing, usage, is_ft)


########################
# USELESS FOR NOW
########################


class CompletionsUsage(BaseModel):
    """The aggregated completions usage details of the specific time bucket."""

    object: str = Field(default="organization.usage.completions.result", description="Type identifier for the completions usage object")
    input_tokens: int = Field(
        description="The aggregated number of text input tokens used, including cached tokens. For customers subscribe to scale tier, this includes scale tier tokens."
    )
    input_cached_tokens: int = Field(
        description="The aggregated number of text input tokens that has been cached from previous requests. For customers subscribe to scale tier, this includes scale tier tokens."
    )
    output_tokens: int = Field(description="The aggregated number of text output tokens used. For customers subscribe to scale tier, this includes scale tier tokens.")
    input_audio_tokens: int = Field(description="The aggregated number of audio input tokens used, including cached tokens.")
    output_audio_tokens: int = Field(description="The aggregated number of audio output tokens used.")
    num_model_requests: int = Field(description="The count of requests made to the model.")
    project_id: Optional[str] = Field(default=None, description="When group_by=project_id, this field provides the project ID of the grouped usage result.")
    user_id: Optional[str] = Field(default=None, description="When group_by=user_id, this field provides the user ID of the grouped usage result.")
    api_key_id: Optional[str] = Field(default=None, description="When group_by=api_key_id, this field provides the API key ID of the grouped usage result.")
    model: Optional[str] = Field(default=None, description="When group_by=model, this field provides the model name of the grouped usage result.")
    batch: Optional[bool] = Field(default=None, description="When group_by=batch, this field tells whether the grouped usage result is batch or not.")
