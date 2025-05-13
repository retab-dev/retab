from typing import get_args

from ..types.ai_models import AIProvider, GeminiModel, OpenAIModel, xAI_Model


def find_provider_from_model(model: str) -> AIProvider:
    if model in get_args(OpenAIModel):
        return "OpenAI"
    elif ":" in model:
        # Handle fine-tuned models
        ft, base_model, model_id = model.split(":", 2)
        if base_model in get_args(OpenAIModel):
            return "OpenAI"
    # elif model in get_args(AnthropicModel):
    #    return "Anthropic"
    elif model in get_args(xAI_Model):
        return "xAI"
    elif model in get_args(GeminiModel):
        return "Gemini"
    raise ValueError(f"Could not determine AI provider for model: {model}")


def assert_valid_model_extraction(model: str) -> None:
    if model in get_args(OpenAIModel):
        return
    elif ":" in model:
        # Handle fine-tuned models
        ft, base_model, model_id = model.split(":", 2)
        if base_model in get_args(OpenAIModel):
            return
    # elif model in get_args(AnthropicModel):
    #    return
    elif model in get_args(xAI_Model):
        return
    elif model in get_args(GeminiModel):
        return
    raise ValueError(
        f"Invalid model for extraction: {model}.\nValid OpenAI models: {get_args(OpenAIModel)}\n"
        # f"Valid Anthropic models: {get_args(AnthropicModel)}\n"
        # f"Valid xAI models: {get_args(xAI_Model)}\n"
        # f"Valid Gemini models: {get_args(GeminiModel)}"
    )


def assert_valid_model_batch_processing(model: str) -> None:
    """Assert that the model is either a standard OpenAI model or a valid fine-tuned model.

    Valid formats:
    - Standard model: Must be in OpenAIModel
    - Fine-tuned model: Must be {base_model}:{id} where base_model is in OpenAIModel

    Raises:
        ValueError: If the model format is invalid
    """
    if model in get_args(OpenAIModel):
        return

    try:
        ft, base_model, model_id = model.split(":", 2)
        if base_model not in get_args(OpenAIModel):
            raise ValueError(f"Invalid base model in fine-tuned model '{model}'. Base model must be one of: {get_args(OpenAIModel)}")
        if not model_id or not model_id.strip():
            raise ValueError(f"Model ID cannot be empty in fine-tuned model '{model}'")
    except ValueError as e:
        if ":" not in model:
            raise ValueError(
                f"Invalid model format: {model}. Must be either:\n"
                f"1. A standard model: {get_args(OpenAIModel)}\n"
                f"2. A fine-tuned model in format 'base_model:id' where base_model is one of the standard models"
            ) from None
        raise


def assert_valid_model_schema_generation(model: str) -> None:
    """Assert that the model is either a standard OpenAI model or a valid fine-tuned model.

    Valid formats:
    - Standard model: Must be in OpenAIModel
    - Fine-tuned model: Must be {base_model}:{id} where base_model is in OpenAIModel

    Raises:
        ValueError: If the model format is invalid
    """
    if model in get_args(OpenAIModel):
        return

    try:
        ft, base_model, model_id = model.split(":", 2)
        if base_model not in get_args(OpenAIModel):
            raise ValueError(f"Invalid base model in fine-tuned model '{model}'. Base model must be one of: {get_args(OpenAIModel)}")
        if not model_id or not model_id.strip():
            raise ValueError(f"Model ID cannot be empty in fine-tuned model '{model}'")
    except ValueError as e:
        if ":" not in model:
            raise ValueError(
                f"Invalid model format: {model}. Must be either:\n"
                f"1. A standard model: {get_args(OpenAIModel)}\n"
                f"2. A fine-tuned model in format 'base_model:id' where base_model is one of the standard models"
            ) from None
        raise
