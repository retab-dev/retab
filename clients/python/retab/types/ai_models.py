import datetime
from typing import List, Literal, Optional
from pydantic import BaseModel, Field, computed_field
from retab.types.inference_settings import InferenceSettings


AIProvider = Literal["OpenAI", "Anthropic", "Gemini", "xAI", "Retab"]
OpenAICompatibleProvider = Literal["OpenAI", "xAI"]
GeminiModel = Literal[
    "gemini-2.5-pro",
    "gemini-2.5-flash",
    "gemini-2.5-pro-preview-06-05",
    "gemini-2.5-pro-preview-05-06",
    "gemini-2.5-pro-preview-03-25",
    "gemini-2.5-flash-preview-05-20",
    "gemini-2.5-flash-preview-04-17",
    "gemini-2.5-flash-lite-preview-06-17",
    "gemini-2.5-pro-exp-03-25",
    "gemini-2.0-flash-lite",
    "gemini-2.0-flash",
]

AnthropicModel = Literal[
    "claude-3-5-sonnet-latest",
    "claude-3-5-sonnet-20241022",
    # "claude-3-5-haiku-20241022",
    "claude-3-opus-20240229",
    "claude-3-sonnet-20240229",
    "claude-3-haiku-20240307",
]
OpenAIModel = Literal[
    "gpt-4o",
    "gpt-4o-mini",
    "chatgpt-4o-latest",
    "gpt-4.1",
    "gpt-4.1-mini",
    "gpt-4.1-mini-2025-04-14",
    "gpt-4.1-2025-04-14",
    "gpt-4.1-nano",
    "gpt-4.1-nano-2025-04-14",
    "gpt-4o-2024-11-20",
    "gpt-4o-2024-08-06",
    "gpt-4o-2024-05-13",
    "gpt-4o-mini-2024-07-18",
    # "o3-mini",
    # "o3-mini-2025-01-31",
    "o1",
    "o1-2024-12-17",
    # "o1-preview-2024-09-12",
    # "o1-mini",
    # "o1-mini-2024-09-12",
    "o3",
    "o3-2025-04-16",
    "o4-mini",
    "o4-mini-2025-04-16",
    "gpt-4o-audio-preview-2024-12-17",
    "gpt-4o-audio-preview-2024-10-01",
    "gpt-4o-realtime-preview-2024-12-17",
    "gpt-4o-realtime-preview-2024-10-01",
    "gpt-4o-mini-audio-preview-2024-12-17",
    "gpt-4o-mini-realtime-preview-2024-12-17",
]
xAI_Model = Literal["grok-3", "grok-3-mini"]
RetabModel = Literal["auto-large", "auto-small", "auto-micro"]
PureLLMModel = Literal[OpenAIModel, AnthropicModel, xAI_Model, GeminiModel, RetabModel]
LLMModel = Literal[PureLLMModel, "human"]


class FinetunedModel(BaseModel):
    object: Literal["finetuned_model"] = "finetuned_model"
    organization_id: str
    model: str
    schema_id: str
    schema_data_id: str
    finetuning_props: InferenceSettings
    evaluation_id: str | None = None
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))


# Monthly Usage
class MonthlyUsageResponseContent(BaseModel):
    credits_count: float  # Changed to float to support decimal credits


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
    cached_discount: float = 1.0  # Price discount for cached tokens. (1.0 means no discount or tokens are not cached)


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
FeatureType = Literal["streaming", "function_calling", "structured_outputs", "distillation", "fine_tuning", "predicted_outputs", "schema_generation"]


class ModelCapabilities(BaseModel):
    """
    Represents the capabilities of an AI model, including supported modalities,
    endpoints, and features.
    """

    modalities: List[ModelModality]
    endpoints: List[EndpointType]
    features: List[FeatureType]


class ModelCardPermissions(BaseModel):
    show_in_free_picker: bool = False
    show_in_paid_picker: bool = False


class ModelCard(BaseModel):
    """
    Model card that includes pricing and capabilities.
    """

    model: LLMModel | str  # Can be a random string for finetuned models
    pricing: Pricing
    capabilities: ModelCapabilities
    temperature_support: bool = True
    reasoning_effort_support: bool = False
    permissions: ModelCardPermissions = ModelCardPermissions()

    @computed_field
    @property
    def is_finetuned(self) -> bool:
        if "ft:" in self.model:
            return True
        return False
