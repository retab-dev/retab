


from ...types.usage import Amount, TokenPrice, Pricing
from openai.types.completion_usage import CompletionUsage


anthropic_pricing_list = [
    Pricing(
        model="claude-3-5-sonnet-20241022",
        text=TokenPrice(prompt=3.00, cached_discount=0.5, completion=15.00),
        audio=None  # Anthropic models do not support audio.
    ),
    Pricing(
        model="claude-3-5-haiku-20241022",
        text=TokenPrice(prompt=0.80, cached_discount=0.5, completion=4.00),
        audio=None
    ),
    Pricing(
        model="claude-3-opus-20240229",
        text=TokenPrice(prompt=15.00, cached_discount=0.5, completion=75.00),
        audio=None
    ),
    Pricing(
        model="claude-3-sonnet-20240229",
        text=TokenPrice(prompt=3.00, cached_discount=0.5, completion=15.00),
        audio=None
    ),
    Pricing(
        model="claude-3-haiku-20240307",
        text=TokenPrice(prompt=0.25, cached_discount=0.5, completion=1.25),
        audio=None
    ),
]

