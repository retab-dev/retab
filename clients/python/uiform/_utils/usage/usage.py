from typing import Optional, Dict

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

########################
# DETAILED COST BREAKDOWN
########################

class TokenCounts(BaseModel):
    """Detailed breakdown of token counts by type and category."""
    
    # Prompt token counts
    prompt_regular_text: int
    prompt_cached_text: int
    prompt_audio: int
    
    # Completion token counts
    completion_regular_text: int
    completion_audio: int
    
    # Total tokens (should match sum of all components)
    total_tokens: int


class CostBreakdown(BaseModel):
    """Detailed breakdown of API call costs by token type and usage category."""
    
    # Total cost amount
    total: Amount
    
    # Text token costs broken down by category
    text_prompt_cost: Amount
    text_cached_cost: Amount
    text_completion_cost: Amount
    text_total_cost: Amount
    
    # Audio token costs broken down by category (if applicable)
    audio_prompt_cost: Optional[Amount] = None
    audio_completion_cost: Optional[Amount] = None
    audio_total_cost: Optional[Amount] = None
    
    # Token counts for reference
    token_counts: TokenCounts
    
    # Model and fine-tuning information
    model: str
    is_fine_tuned: bool = False


def compute_api_call_cost_with_breakdown(pricing: Pricing, usage: CompletionUsage, model: str, is_ft: bool = False) -> CostBreakdown:
    """
    Computes a detailed price breakdown for the given token usage, based on the pricing.
    
    Returns a CostBreakdown object containing costs broken down by token type and category.
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
    cost_audio_prompt = 0.0
    cost_audio_completion = 0.0
    total_audio_cost = 0.0
    
    if pricing.audio and (prompt_audio > 0 or completion_audio > 0):
        cost_audio_prompt = prompt_audio * pricing.audio.prompt
        cost_audio_completion = completion_audio * pricing.audio.completion
        total_audio_cost = cost_audio_prompt + cost_audio_completion

    # Convert to dollars (divide by 1M) and create Amount objects
    ft_multiplier = pricing.ft_price_hike if is_ft else 1.0
    
    # Create Amount objects for each cost category
    text_prompt_amount = Amount(value=(cost_text_prompt / 1e6) * ft_multiplier, currency="USD")
    text_cached_amount = Amount(value=(cost_text_cached / 1e6) * ft_multiplier, currency="USD")
    text_completion_amount = Amount(value=(cost_text_completion / 1e6) * ft_multiplier, currency="USD")
    text_total_amount = Amount(value=(total_text_cost / 1e6) * ft_multiplier, currency="USD")
    
    # Audio amounts (if applicable)
    audio_prompt_amount = None
    audio_completion_amount = None
    audio_total_amount = None
    
    if pricing.audio and (prompt_audio > 0 or completion_audio > 0):
        audio_prompt_amount = Amount(value=(cost_audio_prompt / 1e6) * ft_multiplier, currency="USD")
        audio_completion_amount = Amount(value=(cost_audio_completion / 1e6) * ft_multiplier, currency="USD")
        audio_total_amount = Amount(value=(total_audio_cost / 1e6) * ft_multiplier, currency="USD")
    
    # Total cost
    total_cost = (total_text_cost + total_audio_cost) / 1e6 * ft_multiplier
    total_amount = Amount(value=total_cost, currency="USD")
    
    # Create TokenCounts object with token usage breakdown
    token_counts = TokenCounts(
        prompt_regular_text=prompt_regular_text,
        prompt_cached_text=prompt_cached_text,
        prompt_audio=prompt_audio,
        completion_regular_text=completion_regular_text,
        completion_audio=completion_audio,
        total_tokens=usage.total_tokens
    )
    
    return CostBreakdown(
        total=total_amount,
        text_prompt_cost=text_prompt_amount,
        text_cached_cost=text_cached_amount,
        text_completion_cost=text_completion_amount,
        text_total_cost=text_total_amount,
        audio_prompt_cost=audio_prompt_amount,
        audio_completion_cost=audio_completion_amount,
        audio_total_cost=audio_total_amount,
        token_counts=token_counts,
        model=model,
        is_fine_tuned=is_ft
    )


def compute_cost_from_model_with_breakdown(model: str, usage: CompletionUsage) -> CostBreakdown:
    """
    Computes a detailed cost breakdown for an API call using the specified model and usage.
    
    Args:
        model: The model name (can be a fine-tuned model like "ft:gpt-4o:uiform:4389573")
        usage: Token usage statistics for the API call
        
    Returns:
        CostBreakdown object with detailed cost information
        
    Raises:
        ValueError: If no pricing information is found for the model
    """
    # Extract base model name for fine-tuned models like "ft:gpt-4o:uiform:4389573"
    original_model = model
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
        raise ValueError(f"No pricing information found for model: {original_model}")

    return compute_api_call_cost_with_breakdown(pricing, usage, original_model, is_ft)
