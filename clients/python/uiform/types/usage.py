from pydantic import BaseModel
from typing import Optional, Literal, Dict, List
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
    text: TokenPrice
    audio: Optional[TokenPrice] = None  # May be None if the model does not support audio tokens.
    ft_price_hike: float = 1.0  # Price hike for fine-tuned models.


# Fix the type definition - use proper Python syntax
ModelModality = Literal["text", "audio", "image"]

# Define supported endpoints
EndpointType = Literal[
    "chat_completions", 
    "responses", 
    "assistants", 
    "batch", 
    "fine_tuning",
    "embeddings",
    "speech_generation",
    "translation",
    "completions_legacy",
    "image_generation",
    "transcription",
    "moderation",
    "realtime"
]

# Define supported features
FeatureType = Literal[
    "streaming", 
    "function_calling", 
    "structured_outputs", 
    "distillation", 
    "fine_tuning",
    "predicted_outputs"
]

class ModelCapabilities(BaseModel):
    """
    Represents the capabilities of an AI model, including supported modalities,
    endpoints, and features.
    """
    modalities: List[ModelModality]
    endpoints: List[EndpointType]
    features: List[FeatureType]


class ModelCard(BaseModel):
    """
    Model card that includes pricing and capabilities.
    """
    model: str
    pricing: Pricing
    capabilities: ModelCapabilities


# List of model cards with pricing and capabilities
model_cards = [

    ########################
    ########################
    # ----------------------
    # Reasoning models
    # ----------------------
    ########################
    ########################

    # ----------------------
    # o1 family
    # ----------------------
    ModelCard(
        model="o1",
        pricing=Pricing(
            text=TokenPrice(prompt=15.00, cached_discount=0.5, completion=60.00),
            audio=None
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch"],
            features=["streaming", "function_calling", "structured_outputs"]
        )
    ),
    ModelCard(
        model="o1-2024-12-17",
        pricing=Pricing(
            text=TokenPrice(prompt=15.00, cached_discount=0.5, completion=60.00),
            audio=None
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch"],
            features=["streaming", "function_calling", "structured_outputs"]
        )
    ),

    ModelCard(
        model="o1-preview-2024-09-12",
        pricing=Pricing(
            text=TokenPrice(prompt=15.00, cached_discount=0.5, completion=60.00),
            audio=None
        ),
        capabilities=ModelCapabilities(
            modalities=["text"],
            endpoints=["chat_completions", "responses", "assistants", "batch"],
            features=["streaming", "function_calling"]
        )
    ),


    # ----------------------
    # o1-mini family
    # ----------------------
    ModelCard(
        model="o1-mini",
        pricing=Pricing(
            text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40),
            audio=None
        ),
        capabilities=ModelCapabilities(
            modalities=["text"],
            endpoints=["chat_completions", "responses", "assistants"],
            features=["streaming"]
        )
    ),
    ModelCard(
        model="o1-mini-2024-09-12",
        pricing=Pricing(
            text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40),
            audio=None
        ),
        capabilities=ModelCapabilities(
            modalities=["text"],
            endpoints=["chat_completions", "responses", "assistants"],
            features=["streaming"]
        )
    ),


    # ----------------------
    # o3-mini family
    # ----------------------
    ModelCard(
        model="o3-mini-2025-01-31",
        pricing=Pricing(
            text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40),
            audio=None,
            ft_price_hike=1.5
        ),
        capabilities=ModelCapabilities(
            modalities=["text"],
            endpoints=["chat_completions", "responses", "assistants", "batch"],
            features=["streaming", "function_calling", "structured_outputs"]
        )
    ),
    ModelCard(
        model="o3-mini",
        pricing=Pricing(
            text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40),
            audio=None,
            ft_price_hike=1.5
        ),
        capabilities=ModelCapabilities(
            modalities=["text"],
            endpoints=["chat_completions", "responses", "assistants", "batch"],
            features=["streaming", "function_calling", "structured_outputs"]
        )
    ),
    

    ########################
    ########################
    # ----------------------
    # Chat models
    # ----------------------
    ########################
    ########################

    # ----------------------
    # gpt-4.5 family
    # ----------------------
    ModelCard(
        model="gpt-4.5-preview",
        pricing=Pricing(
            text=TokenPrice(prompt=75, cached_discount=0.5, completion=150.00),
            audio=None,
            ft_price_hike=1.5
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch"],
            features=["streaming", "function_calling", "structured_outputs"]
        )
    ),
    ModelCard(
        model="gpt-4.5-preview-2025-02-27",
        pricing=Pricing(
            text=TokenPrice(prompt=75, cached_discount=0.5, completion=150.00),
            audio=None,
            ft_price_hike=1.5
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch"],
            features=["streaming", "function_calling", "structured_outputs"]
        )
    ),

    # ----------------------
    # gpt-4o family
    # ----------------------
    ModelCard(
        model="chatgpt-4o-latest",
        pricing=Pricing(
            text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
            audio=None,
            ft_price_hike=1.5
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning", "distillation", "predicted_outputs"]
        )
    ),
    ModelCard(
        model="gpt-4o",
        pricing=Pricing(
            text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
            audio=None,
            ft_price_hike=1.5
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning", "distillation", "predicted_outputs"]
        )
    ),
    ModelCard(
        model="gpt-4o-2024-08-06",
        pricing=Pricing(
            text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
            audio=None,
            ft_price_hike=1.5
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning", "distillation", "predicted_outputs"]
        )
    ),
    ModelCard(
        model="gpt-4o-2024-11-20",
        pricing=Pricing(
            text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
            audio=None,
            ft_price_hike=1.5
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning", "distillation", "predicted_outputs"]
        )
    ),
    ModelCard(
        model="gpt-4o-2024-05-13",
        pricing=Pricing(
            text=TokenPrice(prompt=5.00, cached_discount=0.5, completion=15.00),
            audio=None,
            ft_price_hike=1.5
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning", "distillation", "predicted_outputs"]
        )
    ),

    # ----------------------
    # gpt-4o-audio family
    # ----------------------
    ModelCard(
        model="gpt-4o-audio-preview-2024-12-17",
        pricing=Pricing(
            text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
            audio=TokenPrice(prompt=40.00, cached_discount=0.2, completion=80.00)
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "audio"],
            endpoints=["chat_completions"],
            features=["streaming", "function_calling"]
        )
    ),
    ModelCard(
        model="gpt-4o-audio-preview-2024-10-01",
        pricing=Pricing(
            text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00),
            audio=TokenPrice(prompt=100.00, cached_discount=0.2, completion=200.00)
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "audio"],
            endpoints=["chat_completions"],
            features=["streaming", "function_calling"]
        )
    ),

        ModelCard(
        model="gpt-4o-realtime-preview-2024-12-17",
        pricing=Pricing(
            text=TokenPrice(prompt=5.00, cached_discount=0.5, completion=20.00),
            audio=TokenPrice(prompt=40.00, cached_discount=0.2, completion=80.00)
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "audio"],
            endpoints=["chat_completions"],
            features=["streaming", "function_calling"]
        )
    ),
    ModelCard(
        model="gpt-4o-realtime-preview-2024-10-01",
        pricing=Pricing(
            text=TokenPrice(prompt=5.00, cached_discount=0.5, completion=20.00),
            audio=TokenPrice(prompt=100.00, cached_discount=0.2, completion=200.00)
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "audio"],
            endpoints=["chat_completions"],
            features=["streaming", "function_calling"]
        )
    ),

    # ----------------------
    # gpt-4o-mini family
    # ----------------------
    ModelCard(
        model="gpt-4o-mini",
        pricing=Pricing(
            text=TokenPrice(prompt=0.15, cached_discount=0.5, completion=0.60),
            audio=None,
            ft_price_hike=2.0
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning"]
        )
    ),
    ModelCard(
        model="gpt-4o-mini-2024-07-18",
        pricing=Pricing(
            text=TokenPrice(prompt=0.15, cached_discount=0.5, completion=0.60),
            audio=None,
            ft_price_hike=2.0
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning"]
        )
    ),


    # ----------------------
    # gpt-4o-mini-audio family
    # ----------------------
    ModelCard(
        model="gpt-4o-mini-audio-preview-2024-12-17",
        pricing=Pricing(
            text=TokenPrice(prompt=0.15, cached_discount=0.5, completion=0.60),
            audio=TokenPrice(prompt=10.00, cached_discount=0.2, completion=20.00),
            ft_price_hike=2.0
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "audio"],
            endpoints=["chat_completions"],
            features=["streaming", "function_calling"]
        )
    ),
    ModelCard(
        model="gpt-4o-mini-realtime-preview-2024-12-17",
        pricing=Pricing(
            text=TokenPrice(prompt=0.60, cached_discount=0.5, completion=2.40),
            audio=TokenPrice(prompt=10.00, cached_discount=0.2, completion=20.00),
            ft_price_hike=2.0
        ),
        capabilities=ModelCapabilities(
            modalities=["text", "audio"],
            endpoints=["chat_completions"],
            features=["streaming", "function_calling"]
        )
    ),

    

    
]

def get_model_card(model: str) -> ModelCard:
    """
    Get the model card for a specific model.
    
    Args:
        model: The model name to look up
        
    Returns:
        The ModelCard for the specified model
        
    Raises:
        ValueError: If no model card is found for the specified model
    """
    # Extract base model name for fine-tuned models like "ft:gpt-4o:uiform:4389573"
    if model.startswith("ft:"):
        # Split by colon and take the second part (index 1) which contains the base model
        parts = model.split(":")
        if len(parts) > 1:
            model = parts[1]
    
    for card in model_cards:
        if card.model == model:
            return card
    
    raise ValueError(f"No model card found for model: {model}")
