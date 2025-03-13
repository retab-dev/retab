
from ...types.usage import Amount, TokenPrice, Pricing
from openai.types.completion_usage import CompletionUsage

openai_pricing_list = [

    # o3-mini family
    Pricing(
        model="o3-mini-2025-01-31",
        text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40),
        audio=None,
        ft_price_hike=1.5
    ),
    Pricing(
        model="o3-mini",
        text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40),
        audio=None,
        ft_price_hike=1.5
    ),

    # gpt-4.5-preview family
    Pricing(
        model="gpt-4.5-preview",
        text=TokenPrice(prompt=75, cached_discount=0.5, completion=150.00),
        audio=None,
        ft_price_hike=1.5
    ),
    Pricing(
        model="gpt-4.5-preview-2025-02-27",
        text=TokenPrice(prompt=75, cached_discount=0.5, completion=150.00),
        audio=None,
        ft_price_hike=1.5
    ),

    # gpt-4o family (text-only)
    Pricing(
        model="chatgpt-4o-latest",
        text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
        audio=None,
        ft_price_hike=1.5
    ),

    Pricing(
        model="gpt-4o",
        text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
        audio=None,
        ft_price_hike=1.5
    ),

    Pricing(
        model="gpt-4o-2024-08-06",
        text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
        audio=None,
        ft_price_hike=1.5
    ),
    Pricing(
        model="gpt-4o-2024-11-20",
        text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
        audio=None,
        ft_price_hike=1.5
    ),
    Pricing(
        model="gpt-4o-2024-05-13",
        text=TokenPrice(prompt=5.00, cached_discount=0.5, completion=15.00),
        audio=None,
        ft_price_hike=1.5
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
        model="gpt-4o-mini",
        text=TokenPrice(prompt=0.15, cached_discount=0.5, completion=0.60),
        audio=None,
        ft_price_hike=2.0
    ),

    Pricing(
        model="gpt-4o-mini-2024-07-18",
        text=TokenPrice(prompt=0.15, cached_discount=0.5, completion=0.60),
        audio=None,
        ft_price_hike=2.0
    ),
    Pricing(
        model="gpt-4o-mini-audio-preview-2024-12-17",
        text=TokenPrice(prompt=0.15, cached_discount=0.5, completion=0.60),
        audio=TokenPrice(prompt=10.00, cached_discount=0.2, completion=20.00),
        ft_price_hike=2.0
    ),
    Pricing(
        model="gpt-4o-mini-realtime-preview-2024-12-17",
        text=TokenPrice(prompt=0.60, cached_discount=0.5, completion=2.40),
        audio=TokenPrice(prompt=10.00, cached_discount=0.2, completion=20.00),
        ft_price_hike=2.0
    ),

    # o1 family
    Pricing(
        model="o1",
        text=TokenPrice(prompt=15.00, cached_discount=0.5, completion=60.00),
        audio=None
    ),
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

