from pydantic import BaseModel, Field
from typing import Optional, Literal


#https://platform.openai.com/docs/guides/prompt-caching
from openai.types.completion_usage import CompletionUsage, PromptTokensDetails, CompletionTokensDetails

# ─── PRICING MODELS ────────────────────────────────────────────────────────────

class TokenPrice(BaseModel):
    """
    Holds pricing information (price per 1M tokens) for one token category.
    (For example, for text tokens used in the prompt.)
    """
    prompt: float        # Price per 1M prompt tokens.
    completion: float    # Price per 1M completion tokens.
    cached_discount: float = 0.5  # Price discount for cached tokens.


class Pricing(BaseModel):
    """
    Contains all pricing information that is useful for a given model.
    (For example, the grid below shows that for gpt-4o-2024-08-06 text prompt tokens
     cost $2.50 per 1M tokens while completion tokens cost $10.00 per 1M tokens.)
    """
    model: str
    text: TokenPrice
    audio: Optional[TokenPrice] = None  # May be None if the model does not support audio tokens.


pricing_list = [
    # gpt-4o family (text-only)
    Pricing(
        model="gpt-4o-2024-08-06",
        text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
        audio=None
    ),
    Pricing(
        model="gpt-4o-2024-11-20",
        text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
        audio=None
    ),
    Pricing(
        model="gpt-4o-2024-05-13",
        text=TokenPrice(prompt=5.00, cached_discount=0.5, completion=15.00),
        audio=None
    ),

    # gpt-4o-audio-preview family (both text and audio pricing)
    Pricing(
        model="gpt-4o-audio-preview-2024-12-17",
        text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
        audio=TokenPrice(prompt=40.00, cached_discount=0.2, completion=80.00)
    ),
    Pricing(
        model="gpt-4o-audio-preview-2024-10-01",
        text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
        audio=TokenPrice(prompt=100.00, cached_discount=0.2, completion=200.00)
    ),

    # gpt-4o-realtime-preview family (both text and audio pricing)
    Pricing(
        model="gpt-4o-realtime-preview-2024-12-17",
        text=TokenPrice(prompt=5.00, cached_discount=0.5, completion=20.00),
        audio=TokenPrice(prompt=40.00, cached_discount=0.2, completion=80.00)
    ),
    Pricing(
        model="gpt-4o-realtime-preview-2024-10-01",
        text=TokenPrice(prompt=5.00, cached_discount=0.5, completion=20.00),
        audio=TokenPrice(prompt=100.00, cached_discount=0.2, completion=200.00)
    ),

    # gpt-4o-mini family
    Pricing(
        model="gpt-4o-mini-2024-07-18",
        text=TokenPrice(prompt=0.15, cached_discount=0.5, completion=0.60),
        audio=None
    ),
    Pricing(
        model="gpt-4o-mini-audio-preview-2024-12-17",
        text=TokenPrice(prompt=0.15, cached_discount=0.5, completion=0.60),
        audio=TokenPrice(prompt=10.00, cached_discount=0.2, completion=20.00)
    ),
    Pricing(
        model="gpt-4o-mini-realtime-preview-2024-12-17",
        text=TokenPrice(prompt=0.60, cached_discount=0.5, completion=2.40),
        audio=TokenPrice(prompt=10.00, cached_discount=0.2, completion=20.00)
    ),

    # o1 family
    Pricing(
        model="o1-2024-12-17",
        text=TokenPrice(prompt=15.00, cached_discount=0.5, completion=60.00),
        audio=None
    ),
    Pricing(
        model="o1-preview-2024-09-12",
        text=TokenPrice(prompt=15.00, cached_discount=0.5, completion=60.00),
        audio=None
    ),

    # o3/o1-mini families
    Pricing(
        model="o3-mini-2025-01-31",
        text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40),
        audio=None
    ),
    Pricing(
        model="o1-mini-2024-09-12",
        text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40),
        audio=None
    ),
]

# ─── GENERIC FUNCTION TO COMPUTE COST ─────────────────────────────────────────


class Amount(BaseModel):
    value: float
    currency: str

class Cost(BaseModel):
    object: Literal['cost'] = 'cost'
    usage: CompletionUsage
    model: str

    @property
    def amount(self) -> Amount:
        return compute_cost_from_model(self.model, self.usage)




def compute_cost(pricing: Pricing, usage: CompletionUsage) -> Amount:
    """
    Computes the price (as an Amount) for the given token usage, based on the pricing.
    
    Assumptions / rules:
      - The pricing rates are per 1,000,000 tokens.
      - The usage is broken out into prompt and completion tokens.
      - If details are provided, the prompt tokens are split into:
          • cached text tokens (if any),
          • audio tokens (if any), and
          • the remaining are treated as “regular” text tokens.
      - Similarly, completion tokens are split into audio tokens (if any) and the remaining text.
      - For text tokens, if a cached price is not explicitly provided in pricing.text.cached,
        we assume a 50% discount off the prompt price.
      - (A similar rule could be applied to audio tokens – i.e. 20% of the normal price –
       if cached audio tokens were provided. In our usage model, however, only text tokens
       are marked as cached.)
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

    return Amount(value=total_cost, currency="USD")

def compute_cost_from_model(model: str, usage: CompletionUsage) -> Amount:
    pricing = next((p for p in pricing_list if p.model == model), None)
    if not pricing:
        raise ValueError(f"No pricing information found for model: {model}")
    return compute_cost(pricing, usage)

    


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
