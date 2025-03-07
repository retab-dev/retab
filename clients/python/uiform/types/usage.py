from pydantic import BaseModel
from typing import Optional
# Monthly Usage
class MonthlyUsageResponseContent(BaseModel):
    request_count: int

MonthlyUsageResponse = MonthlyUsageResponseContent

class Amount(BaseModel):
    value: float
    currency: str


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
    ft_price_hike: float = 1.0  # Price hike for fine-tuned models.


