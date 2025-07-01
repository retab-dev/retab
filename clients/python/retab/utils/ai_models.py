import os
import yaml
from typing import get_args

from ..types.ai_models import AIProvider, GeminiModel, OpenAIModel, xAI_Model, RetabModel, PureLLMModel, ModelCard

MODEL_CARDS_DIR = os.path.join(os.path.dirname(__file__), "_model_cards")


def merge_model_cards(base: dict, override: dict) -> dict:
    result = base.copy()
    for key, value in override.items():
        if key == "inherits":
            continue
        if isinstance(value, dict) and key in result:
            result[key] = merge_model_cards(result[key], value)
        else:
            result[key] = value
    return result


def load_model_cards(yaml_file: str) -> list[ModelCard]:
    raw_cards = yaml.safe_load(open(yaml_file))
    name_to_card = {c["model"]: c for c in raw_cards if "inherits" not in c}

    final_cards = []
    for card in raw_cards:
        if "inherits" in card:
            parent = name_to_card[card["inherits"]]
            merged = merge_model_cards(parent, card)
            final_cards.append(ModelCard(**merged))
        else:
            final_cards.append(ModelCard(**card))
    return final_cards


# Load all model cards
model_cards = sum(
    [
        load_model_cards(os.path.join(MODEL_CARDS_DIR, "openai.yaml")),
        load_model_cards(os.path.join(MODEL_CARDS_DIR, "anthropic.yaml")),
        load_model_cards(os.path.join(MODEL_CARDS_DIR, "xai.yaml")),
        load_model_cards(os.path.join(MODEL_CARDS_DIR, "gemini.yaml")),
        load_model_cards(os.path.join(MODEL_CARDS_DIR, "auto.yaml")),
    ],
    [],
)
model_cards_dict = {card.model: card for card in model_cards}


# Validate that model cards
all_model_names = set(model_cards_dict.keys())
if all_model_names.symmetric_difference(set(get_args(PureLLMModel))):
    raise ValueError(f"Mismatch between model cards and PureLLMModel type: {all_model_names.symmetric_difference(set(get_args(PureLLMModel)))}")


def get_model_from_model_id(model_id: str) -> str:
    """
    Get the model name from the model id.
    """
    if model_id.startswith("ft:"):
        parts = model_id.split(":")
        return parts[1]
    else:
        return model_id


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
    model_name = get_model_from_model_id(model)
    if model_name in model_cards_dict:
        model_card = ModelCard(**model_cards_dict[model_name].model_dump())
        if model_name != model:
            # Fine-tuned model -> Change the name
            model_card.model = model
            # Remove the fine-tuning feature (if exists)
            try:
                model_card.capabilities.features.remove("fine_tuning")
            except ValueError:
                pass
        return model_card

    raise ValueError(f"No model card found for model: {model_name}")


def get_provider_for_model(model_id: str) -> AIProvider:
    """
    Determine the AI provider associated with the given model identifier.
    Returns one of: "Anthropic", "xAI", "OpenAI", "Gemini", "Retab" or None if unknown.
    """
    model_name = get_model_from_model_id(model_id)
    # if model_name in get_args(AnthropicModel):
    #     return "Anthropic"
    # if model_name in get_args(xAI_Model):
    #     return "xAI"
    if model_name in get_args(OpenAIModel):
        return "OpenAI"
    if model_name in get_args(GeminiModel):
        return "Gemini"
    if model_name in get_args(RetabModel):
        return "Retab"
    raise ValueError(f"Unknown model: {model_name}")


def assert_valid_model_extraction(model: str) -> None:
    try:
        get_provider_for_model(model)
    except ValueError:
        raise ValueError(
            f"Invalid model for extraction: {model}.\nValid OpenAI models: {get_args(OpenAIModel)}\n"
            f"Valid xAI models: {get_args(xAI_Model)}\n"
            f"Valid Gemini models: {get_args(GeminiModel)}"
        ) from None


def assert_valid_model_schema_generation(model: str) -> None:
    """Assert that the model is either a standard OpenAI model or a valid fine-tuned model.

    Valid formats:
    - Standard model: Must be in OpenAIModel
    - Fine-tuned model: Must be {base_model}:{id} where base_model is in OpenAIModel

    Raises:
        ValueError: If the model format is invalid
    """
    if get_model_from_model_id(model) in get_args(OpenAIModel):
        return
    else:
        raise ValueError(
            f"Invalid model format: {model}. Must be either:\n"
            f"1. A standard model: {get_args(OpenAIModel)}\n"
            f"2. A fine-tuned model in format 'base_model:id' where base_model is one of the standard openai models"
        ) from None


def get_model_credits(model: str) -> float:
    """
    Get the credit cost for a given model based on its capabilities and size.

    Credit tiers:
    - 0.1 credits: Micro/nano models (fastest, cheapest)
    - 0.5 credits: Small/mini models (balanced performance)
    - 2.0 credits: Large/advanced models (highest capability)

    Args:
        model: The model name to look up

    Returns:
        The credit cost for the model

    Raises:
        ValueError: If no model card is found for the specified model
    """
    try:
        model_card = get_model_card(model)
        model_name = get_model_from_model_id(model)
    except ValueError:
        # Unknown model, return 0 credits (no billing)
        return 0.0

    # Define credit mapping based on model capabilities and naming patterns
    model_credits = {
        # 0.1 credit models - Micro/Nano tier (fastest, most efficient)
        "auto-micro": 0.1,
        "gemini-flash-lite": 0.1,
        "gpt-4o-mini": 0.1,
        "gpt-3.5-turbo": 0.1,
        "gpt-4.1-nano": 0.1,  # Future model
        # 0.5 credit models - Small/Mini tier (balanced performance)
        "auto-small": 0.5,
        "gemini-flash": 0.5,
        "gpt-4o": 0.5,
        "gpt-4-turbo": 0.5,
        "gpt-4.1-mini": 0.5,  # Future model
        "claude-3-haiku": 0.5,
        "claude-3.5-haiku": 0.5,
        # 2.0 credit models - Large/Advanced tier (highest capability)
        "auto-large": 2.0,
        "gemini-pro": 2.0,
        "gpt-4": 2.0,
        "gpt-4.1": 2.0,  # Future model
        "o1-mini": 2.0,
        "o1-preview": 2.0,
        "o3": 5.0,  # Future model
        "claude-3-sonnet": 2.0,
        "claude-3-opus": 2.0,
        "claude-3.5-sonnet": 2.0,
        "grok-beta": 2.0,
        "grok-2": 2.0,
        # Special reasoning models - Higher tier
        "o1": 3.0,
        "o3-max": 3.0,  # Future model, highest tier
    }

    # Return the credits for the specific model
    if model_name in model_credits:
        return model_credits[model_name]

    # Fallback logic based on model patterns and capabilities
    model_lower = model_name.lower()

    # Auto-model fallback logic
    if model_lower.startswith("auto-"):
        if "micro" in model_lower or "nano" in model_lower:
            return 0.1
        elif "small" in model_lower or "mini" in model_lower:
            return 0.5
        elif "large" in model_lower or "pro" in model_lower:
            return 2.0

    # Gemini model fallback logic
    if "gemini" in model_lower:
        if "lite" in model_lower or "nano" in model_lower:
            return 0.1
        elif "flash" in model_lower:
            return 0.5
        elif "pro" in model_lower or "ultra" in model_lower:
            return 2.0

    # GPT model fallback logic
    if "gpt" in model_lower:
        if "mini" in model_lower or "3.5" in model_lower:
            return 0.1
        elif "4o" in model_lower and "mini" not in model_lower:
            return 0.5
        elif "4" in model_lower or "o1" in model_lower:
            return 2.0

    # Claude model fallback logic
    if "claude" in model_lower:
        if "haiku" in model_lower:
            return 0.5
        elif "sonnet" in model_lower or "opus" in model_lower:
            return 2.0

    # Default for unknown models - use model card info if available
    try:
        # Try to determine based on model card properties
        # This could be enhanced based on the actual ModelCard structure
        return 1.0  # Default middle tier
    except:
        return 0.0  # No billing for completely unknown models
