from typing import Literal

AIProvider = Literal["OpenAI", "Gemini", "xAI", "UiForm"]  # , "Anthropic", "xAI"]
OpenAICompatibleProvider = Literal["OpenAI", "xAI"]  # , "xAI"]
GeminiModel = Literal[
    "gemini-2.5-pro-preview-05-06",
    "gemini-2.5-flash-preview-05-20",
    "gemini-2.5-flash-preview-04-17",
    "gemini-2.5-pro-exp-03-25",
    "gemini-2.0-flash-lite",
    "gemini-2.0-flash",
]

AnthropicModel = Literal[
    "claude-3-5-sonnet-latest", "claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022", "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"
]
OpenAIModel = Literal[
    "gpt-4o",
    "gpt-4o-mini",
    "chatgpt-4o-latest",
    "gpt-4.1",
    "gpt-4.1-mini",
    "gpt-4.1-nano",
    "gpt-4o-2024-11-20",
    "gpt-4o-2024-08-06",
    "gpt-4o-2024-05-13",
    "gpt-4o-mini-2024-07-18",
    "o3-mini",
    "o3-mini-2025-01-31",
    "o1",
    "o1-2024-12-17",
    "o1-preview-2024-09-12",
    "o1-mini",
    "o1-mini-2024-09-12",
    "o3",
    "o3-2025-04-16",
    "o4-mini",
    "o4-mini-2025-04-16",
    "gpt-4.5-preview",
    "gpt-4.5-preview-2025-02-27",
    "gpt-4o-audio-preview-2024-12-17",
    "gpt-4o-audio-preview-2024-10-01",
    "gpt-4o-realtime-preview-2024-12-17",
    "gpt-4o-realtime-preview-2024-10-01",
    "gpt-4o-mini-audio-preview-2024-12-17",
    "gpt-4o-mini-realtime-preview-2024-12-17",
]
xAI_Model = Literal["grok-3-beta", "grok-3-mini-beta"]
UiFormModel = Literal["auto", "auto-small"]
LLMModel = Literal[OpenAIModel, "human", AnthropicModel, xAI_Model, GeminiModel, UiFormModel]

import datetime

from pydantic import BaseModel, Field

from uiform.types.jobs.base import InferenceSettings


class FinetunedModel(BaseModel):
    object: Literal["finetuned_model"] = "finetuned_model"
    organization_id: str
    model: str
    schema_id: str
    schema_data_id: str
    finetuning_props: InferenceSettings
    eval_id: str | None = None
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))


from typing import Dict, List, Literal, Optional

from pydantic import BaseModel


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

    prompt: float  # Price per 1M prompt tokens.
    completion: float  # Price per 1M completion tokens.
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
    "realtime",
]

# Define supported features
FeatureType = Literal["streaming", "function_calling", "structured_outputs", "distillation", "fine_tuning", "predicted_outputs"]


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

    model: LLMModel | str  # Can be a random string for finetuned models
    pricing: Pricing
    capabilities: ModelCapabilities
    temperature_support: bool = True
    reasoning_effort_support: bool = False


# List of model cards with pricing and capabilities
openai_model_cards = [
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
        pricing=Pricing(text=TokenPrice(prompt=15.00, cached_discount=0.5, completion=60.00), audio=None),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=False,
        reasoning_effort_support=True,
    ),
    ModelCard(
        model="o1-2024-12-17",
        pricing=Pricing(text=TokenPrice(prompt=15.00, cached_discount=0.5, completion=60.00), audio=None),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=False,
        reasoning_effort_support=True,
    ),
    ModelCard(
        model="o1-preview-2024-09-12",
        pricing=Pricing(text=TokenPrice(prompt=15.00, cached_discount=0.5, completion=60.00), audio=None),
        capabilities=ModelCapabilities(modalities=["text"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling"]),
        temperature_support=False,
        reasoning_effort_support=True,
    ),
    # ----------------------
    # o1-mini family
    # ----------------------
    ModelCard(
        model="o1-mini",
        pricing=Pricing(text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40), audio=None),
        capabilities=ModelCapabilities(modalities=["text"], endpoints=["chat_completions", "responses", "assistants"], features=["streaming"]),
        temperature_support=False,
        reasoning_effort_support=True,
    ),
    ModelCard(
        model="o1-mini-2024-09-12",
        pricing=Pricing(text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40), audio=None),
        capabilities=ModelCapabilities(modalities=["text"], endpoints=["chat_completions", "responses", "assistants"], features=["streaming"]),
        temperature_support=False,
        reasoning_effort_support=True,
    ),
    # ----------------------
    # o3 family
    # ----------------------
    ModelCard(
        model="o3",
        pricing=Pricing(text=TokenPrice(prompt=10.0, cached_discount=2.5 / 10.0, completion=40.0), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=False,
        reasoning_effort_support=True,
    ),
    ModelCard(
        model="o3-2025-04-16",
        pricing=Pricing(text=TokenPrice(prompt=10.0, cached_discount=2.5 / 10.0, completion=40.0), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=False,
        reasoning_effort_support=True,
    ),
    # ----------------------
    # o3-mini family
    # ----------------------
    ModelCard(
        model="o3-mini-2025-01-31",
        pricing=Pricing(text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=False,
        reasoning_effort_support=True,
    ),
    ModelCard(
        model="o3-mini",
        pricing=Pricing(text=TokenPrice(prompt=1.10, cached_discount=0.5, completion=4.40), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=False,
        reasoning_effort_support=True,
    ),
    # ----------------------
    # o4-mini family
    # ----------------------
    ModelCard(
        model="o4-mini",
        pricing=Pricing(text=TokenPrice(prompt=1.10, cached_discount=0.275 / 1.1, completion=4.40), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=False,
        reasoning_effort_support=True,
    ),
    ModelCard(
        model="o4-mini-2025-04-16",
        pricing=Pricing(text=TokenPrice(prompt=1.10, cached_discount=0.275 / 1.1, completion=4.40), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=False,
        reasoning_effort_support=True,
    ),
    ########################
    ########################
    # ----------------------
    # Chat models
    # ----------------------
    ########################
    ########################
    # ----------------------
    # gpt-4.1 family
    # ----------------------
    ModelCard(
        model="gpt-4.1",
        pricing=Pricing(text=TokenPrice(prompt=2.00, cached_discount=0.25, completion=8.00), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=True,
        reasoning_effort_support=False,
    ),
    ModelCard(
        model="gpt-4.1-2025-04-14",
        pricing=Pricing(text=TokenPrice(prompt=2.00, cached_discount=0.25, completion=8.00), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=True,
        reasoning_effort_support=False,
    ),
    ModelCard(
        model="gpt-4.1-mini",
        pricing=Pricing(text=TokenPrice(prompt=0.40, cached_discount=0.25, completion=1.60), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=True,
        reasoning_effort_support=False,
    ),
    ModelCard(
        model="gpt-4.1-mini-2025-04-14",
        pricing=Pricing(text=TokenPrice(prompt=0.40, cached_discount=0.25, completion=1.60), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=True,
        reasoning_effort_support=False,
    ),
    ModelCard(
        model="gpt-4.1-nano",
        pricing=Pricing(text=TokenPrice(prompt=0.10, cached_discount=0.25, completion=0.40), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=True,
        reasoning_effort_support=False,
    ),
    ModelCard(
        model="gpt-4.1-nano-2025-04-14",
        pricing=Pricing(text=TokenPrice(prompt=0.10, cached_discount=0.25, completion=0.40), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"], endpoints=["chat_completions", "responses", "assistants", "batch"], features=["streaming", "function_calling", "structured_outputs"]
        ),
        temperature_support=True,
        reasoning_effort_support=False,
    ),
    # ----------------------
    # gpt-4.5 family
    # ----------------------
    # ModelCard(
    #     model="gpt-4.5-preview",
    #     pricing=Pricing(
    #         text=TokenPrice(prompt=75, cached_discount=0.5, completion=150.00),
    #         audio=None,
    #         ft_price_hike=1.5
    #     ),
    #     capabilities=ModelCapabilities(
    #         modalities=["text", "image"],
    #         endpoints=["chat_completions", "responses", "assistants", "batch"],
    #         features=["streaming", "function_calling", "structured_outputs"]
    #     ),
    #     logprobs_support=False,
    # ),
    # ModelCard(
    #     model="gpt-4.5-preview-2025-02-27",
    #     pricing=Pricing(
    #         text=TokenPrice(prompt=75, cached_discount=0.5, completion=150.00),
    #         audio=None,
    #         ft_price_hike=1.5
    #     ),
    #     capabilities=ModelCapabilities(
    #         modalities=["text", "image"],
    #         endpoints=["chat_completions", "responses", "assistants", "batch"],
    #         features=["streaming", "function_calling", "structured_outputs"]
    #     ),
    #     logprobs_support=False,
    # ),
    # ----------------------
    # gpt-4o family
    # ----------------------
    ModelCard(
        model="chatgpt-4o-latest",
        pricing=Pricing(text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning", "distillation", "predicted_outputs"],
        ),
    ),
    ModelCard(
        model="gpt-4o",
        pricing=Pricing(text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning", "distillation", "predicted_outputs"],
        ),
    ),
    ModelCard(
        model="gpt-4o-2024-08-06",
        pricing=Pricing(text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning", "distillation", "predicted_outputs"],
        ),
    ),
    ModelCard(
        model="gpt-4o-2024-11-20",
        pricing=Pricing(text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning", "distillation", "predicted_outputs"],
        ),
    ),
    ModelCard(
        model="gpt-4o-2024-05-13",
        pricing=Pricing(text=TokenPrice(prompt=5.00, cached_discount=0.5, completion=15.00), audio=None, ft_price_hike=1.5),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning", "distillation", "predicted_outputs"],
        ),
    ),
    # ----------------------
    # gpt-4o-audio family
    # ----------------------
    ModelCard(
        model="gpt-4o-audio-preview-2024-12-17",
        pricing=Pricing(text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00), audio=TokenPrice(prompt=40.00, cached_discount=0.2, completion=80.00)),
        capabilities=ModelCapabilities(modalities=["text", "audio"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    ModelCard(
        model="gpt-4o-audio-preview-2024-10-01",
        pricing=Pricing(text=TokenPrice(prompt=2.50, cached_discount=0.5, completion=10.00), audio=TokenPrice(prompt=100.00, cached_discount=0.2, completion=200.00)),
        capabilities=ModelCapabilities(modalities=["text", "audio"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    ModelCard(
        model="gpt-4o-realtime-preview-2024-12-17",
        pricing=Pricing(text=TokenPrice(prompt=5.00, cached_discount=0.5, completion=20.00), audio=TokenPrice(prompt=40.00, cached_discount=0.2, completion=80.00)),
        capabilities=ModelCapabilities(modalities=["text", "audio"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    ModelCard(
        model="gpt-4o-realtime-preview-2024-10-01",
        pricing=Pricing(text=TokenPrice(prompt=5.00, cached_discount=0.5, completion=20.00), audio=TokenPrice(prompt=100.00, cached_discount=0.2, completion=200.00)),
        capabilities=ModelCapabilities(modalities=["text", "audio"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    # ----------------------
    # gpt-4o-mini family
    # ----------------------
    ModelCard(
        model="gpt-4o-mini",
        pricing=Pricing(text=TokenPrice(prompt=0.15, cached_discount=0.5, completion=0.60), audio=None, ft_price_hike=2.0),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning"],
        ),
    ),
    ModelCard(
        model="gpt-4o-mini-2024-07-18",
        pricing=Pricing(text=TokenPrice(prompt=0.15, cached_discount=0.5, completion=0.60), audio=None, ft_price_hike=2.0),
        capabilities=ModelCapabilities(
            modalities=["text", "image"],
            endpoints=["chat_completions", "responses", "assistants", "batch", "fine_tuning"],
            features=["streaming", "function_calling", "structured_outputs", "fine_tuning"],
        ),
    ),
    # ----------------------
    # gpt-4o-mini-audio family
    # ----------------------
    ModelCard(
        model="gpt-4o-mini-audio-preview-2024-12-17",
        pricing=Pricing(
            text=TokenPrice(prompt=0.15, cached_discount=0.5, completion=0.60), audio=TokenPrice(prompt=10.00, cached_discount=0.2, completion=20.00), ft_price_hike=2.0
        ),
        capabilities=ModelCapabilities(modalities=["text", "audio"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    ModelCard(
        model="gpt-4o-mini-realtime-preview-2024-12-17",
        pricing=Pricing(
            text=TokenPrice(prompt=0.60, cached_discount=0.5, completion=2.40), audio=TokenPrice(prompt=10.00, cached_discount=0.2, completion=20.00), ft_price_hike=2.0
        ),
        capabilities=ModelCapabilities(modalities=["text", "audio"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
]

# Anthropic Model Cards
anthropic_model_cards = [
    ModelCard(
        model="claude-3-5-sonnet-20241022",
        pricing=Pricing(text=TokenPrice(prompt=3.00, cached_discount=0.5, completion=15.00), audio=None),
        capabilities=ModelCapabilities(modalities=["text", "image"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    ModelCard(
        model="claude-3-5-haiku-20241022",
        pricing=Pricing(text=TokenPrice(prompt=0.80, cached_discount=0.5, completion=4.00), audio=None),
        capabilities=ModelCapabilities(modalities=["text"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    ModelCard(
        model="claude-3-opus-20240229",
        pricing=Pricing(text=TokenPrice(prompt=15.00, cached_discount=0.5, completion=75.00), audio=None),
        capabilities=ModelCapabilities(modalities=["text", "image"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    ModelCard(
        model="claude-3-sonnet-20240229",
        pricing=Pricing(text=TokenPrice(prompt=3.00, cached_discount=0.5, completion=15.00), audio=None),
        capabilities=ModelCapabilities(modalities=["text", "image"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    ModelCard(
        model="claude-3-haiku-20240307",
        pricing=Pricing(text=TokenPrice(prompt=0.25, cached_discount=0.5, completion=1.25), audio=None),
        capabilities=ModelCapabilities(modalities=["text", "image"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
]

# xAI model cards
xai_model_cards = [
    ModelCard(
        model="grok-2-vision-1212",
        pricing=Pricing(text=TokenPrice(prompt=2.00, cached_discount=0.5, completion=10.00), audio=None),
        capabilities=ModelCapabilities(modalities=["text", "image"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    ModelCard(
        model="grok-2-1212",
        pricing=Pricing(text=TokenPrice(prompt=2.00, cached_discount=0.5, completion=10.00), audio=None),
        capabilities=ModelCapabilities(modalities=["text"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    ModelCard(
        model="grok-3-beta",
        pricing=Pricing(text=TokenPrice(prompt=3.00, cached_discount=0.5, completion=15.00), audio=None),
        capabilities=ModelCapabilities(modalities=["text"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
    ModelCard(
        model="grok-3-mini-beta",
        pricing=Pricing(text=TokenPrice(prompt=0.30, cached_discount=0.5, completion=0.50), audio=None),
        capabilities=ModelCapabilities(modalities=["text"], endpoints=["chat_completions"], features=["streaming", "function_calling"]),
    ),
]

# Add Gemini model cards
gemini_model_cards = [
    # ----------------------
    # gemini-2.5-pro-exp-03-25 family
    # ----------------------
    ModelCard(
        model="gemini-2.5-pro-exp-03-25",
        pricing=Pricing(text=TokenPrice(prompt=1.25, cached_discount=0.25, completion=10.00), audio=None),
        capabilities=ModelCapabilities(modalities=["text", "image"], endpoints=["chat_completions"], features=["streaming", "function_calling", "structured_outputs"]),
        temperature_support=True,
    ),
    ModelCard(
        model="gemini-2.5-pro-preview-05-06",
        pricing=Pricing(text=TokenPrice(prompt=1.25, cached_discount=0.25, completion=10.00), audio=None),
        capabilities=ModelCapabilities(modalities=["text", "image"], endpoints=["chat_completions"], features=["streaming", "function_calling", "structured_outputs"]),
    ),
    ModelCard(
        model="gemini-2.5-pro-preview-03-25",
        pricing=Pricing(text=TokenPrice(prompt=1.25, cached_discount=0.25, completion=10.00), audio=None),
        capabilities=ModelCapabilities(modalities=["text", "image"], endpoints=["chat_completions"], features=["streaming", "function_calling", "structured_outputs"]),
        temperature_support=True,
    ),
    ModelCard(
        model="gemini-2.5-flash-preview-04-17",
        pricing=Pricing(text=TokenPrice(prompt=0.15, cached_discount=0.25, completion=0.60), audio=None),
        capabilities=ModelCapabilities(modalities=["text", "image"], endpoints=["chat_completions"], features=["streaming", "function_calling", "structured_outputs"]), 
        temperature_support=True,
    ),
    ModelCard(
        model="gemini-2.5-flash-preview-05-20",
        pricing=Pricing(text=TokenPrice(prompt=0.15, cached_discount=0.25, completion=0.60), audio=None),
        capabilities=ModelCapabilities(modalities=["text", "image"], endpoints=["chat_completions"], features=["streaming", "function_calling", "structured_outputs"]),
        temperature_support=True,
    ),
    # ----------------------
    # gemini-2.0-flash family
    # ----------------------
    ModelCard(
        model="gemini-2.0-flash",
        pricing=Pricing(text=TokenPrice(prompt=0.1, cached_discount=0.025 / 0.1, completion=0.40), audio=TokenPrice(prompt=0.7, cached_discount=0.175 / 0.7, completion=1000)),
        capabilities=ModelCapabilities(modalities=["text", "image"], endpoints=["chat_completions"], features=["streaming", "function_calling", "structured_outputs"]),
        temperature_support=True,
    ),
    ModelCard(
        model="gemini-2.0-flash-lite",
        pricing=Pricing(text=TokenPrice(prompt=0.075, cached_discount=1.0, completion=0.30), audio=TokenPrice(prompt=0.075, cached_discount=1.0, completion=1000)),
        capabilities=ModelCapabilities(
            modalities=["text", "image", "audio"],
            endpoints=["chat_completions"],
            features=["streaming", "structured_outputs"],
        ),
        temperature_support=True,
    ),
]


model_cards = openai_model_cards + gemini_model_cards + xai_model_cards
xAI_model_cards = [
    # ----------------------
    # grok3-family
    # ----------------------
    ModelCard(
        model="grok-3-beta",
        pricing=Pricing(text=TokenPrice(prompt=3, cached_discount=1.0, completion=15), audio=None),
        capabilities=ModelCapabilities(
            modalities=["text"],
            endpoints=["chat_completions"],
            features=["streaming", "structured_outputs"],
        ),
        temperature_support=True,
    ),
    ModelCard(
        model="grok-3-mini-beta",
        pricing=Pricing(text=TokenPrice(prompt=0.3, cached_discount=1.0, completion=0.5), audio=None),
        capabilities=ModelCapabilities(
            modalities=["text"],
            endpoints=["chat_completions"],
            features=["streaming", "structured_outputs"],
        ),
        temperature_support=True,
    ),
]


model_cards = openai_model_cards + gemini_model_cards + xAI_model_cards


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
