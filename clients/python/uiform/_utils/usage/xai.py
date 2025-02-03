




from ...types.usage import Amount, TokenPrice, Pricing
from openai.types.completion_usage import CompletionUsage
from typing import List

xai_pricing_list: List[Pricing] = [
    # Vision models with both text and image inputs:
    Pricing(
        model="grok-2-vision-1212",
        text=TokenPrice(prompt=2.00, completion=10.00, cached_discount=0.5),
        audio=TokenPrice(prompt=2.00, completion=0.0, cached_discount=0.5)
    ),
    Pricing(
        model="grok-2-vision",
        text=TokenPrice(prompt=2.00, completion=10.00, cached_discount=0.5),
        audio=TokenPrice(prompt=2.00, completion=0.0, cached_discount=0.5)
    ),
    Pricing(
        model="grok-2-vision-latest",
        text=TokenPrice(prompt=2.00, completion=10.00, cached_discount=0.5),
        audio=TokenPrice(prompt=2.00, completion=0.0, cached_discount=0.5)
    ),
    # Text-only models:
    Pricing(
        model="grok-2-1212",
        text=TokenPrice(prompt=2.00, completion=10.00, cached_discount=0.5),
        audio=None
    ),
    Pricing(
        model="grok-2",
        text=TokenPrice(prompt=2.00, completion=10.00, cached_discount=0.5),
        audio=None
    ),
    Pricing(
        model="grok-2-latest",
        text=TokenPrice(prompt=2.00, completion=10.00, cached_discount=0.5),
        audio=None
    ),
    # Models with higher pricing:
    Pricing(
        model="grok-vision-beta",
        text=TokenPrice(prompt=5.00, completion=15.00, cached_discount=0.5),
        audio=TokenPrice(prompt=5.00, completion=0.0, cached_discount=0.5)
    ),
    Pricing(
        model="grok-beta",
        text=TokenPrice(prompt=5.00, completion=15.00, cached_discount=0.5),
        audio=None
    ),
]
