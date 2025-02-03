from ...types.usage import TokenPrice, Pricing
from typing import List

gemini_pricing_list: List[Pricing] = [
    Pricing(
        model="gemini-1.5-flash",
        text=TokenPrice(prompt=0.075, completion=0.30, cached_discount=0.5),
        audio=None
    ),
    Pricing(
        model="gemini-1.5-flash-8b",
        text=TokenPrice(prompt=0.0375, completion=0.15, cached_discount=0.5),
        audio=None
    ),
    Pricing(
        model="gemini-2.0-flash-exp",
        text=TokenPrice(prompt=0.075, completion=0.30, cached_discount=0.5),
        audio=None
    ),
    Pricing(
        model="gemini-1.5-pro",
        text=TokenPrice(prompt=1.25, completion=5.00, cached_discount=0.5),
        audio=None
    ),
]
