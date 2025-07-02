from typing import Any, Self, TypedDict

from openai.types.chat.chat_completion_reasoning_effort import (
    ChatCompletionReasoningEffort,
)
from pydantic import BaseModel, Field, model_validator

from ..mime import MIMEData
from ..modalities import Modality
from ..browser_canvas import BrowserCanvas


class EnhanceSchemaConfig(BaseModel):
    allow_reasoning_fields_added: bool = True  # Whether to allow the llm to add reasoning fields
    allow_field_description_update: bool = False  # Whether to allow the llm to update the description of existing fields
    allow_system_prompt_update: bool = True  # Whether to allow the llm to update the system prompt
    allow_field_simple_type_change: bool = False  # Whether to allow the llm to make simple type changes (optional, string to date, etc.)
    allow_field_data_structure_breakdown: bool = False  # Whether to allow the llm to make complex data-structure changes (raw diff)

    # Model validator
    @model_validator(mode="after")
    def check_at_least_one_tool_allowed(self) -> Self:
        if not any(
            [
                self.allow_reasoning_fields_added,
                self.allow_field_description_update,
                self.allow_system_prompt_update,
                self.allow_field_simple_type_change,
                self.allow_field_data_structure_breakdown,
            ]
        ):
            raise ValueError("At least one tool must be allowed")
        return self


# Define a typed Dict for EnhanceSchemaConfig (for now it is kind static, but we will add more flexibility in the future)
class EnhanceSchemaConfigDict(TypedDict, total=False):
    allow_reasoning_fields_added: bool
    allow_field_description_update: bool
    allow_system_prompt_update: bool
    allow_field_simple_type_change: bool
    allow_field_data_structure_breakdown: bool


class EnhanceSchemaRequest(BaseModel):
    """
    The request body for enhancing a JSON Schema.
    """

    documents: list[MIMEData]
    ground_truths: list[dict[str, Any]] | None = None
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    reasoning_effort: ChatCompletionReasoningEffort = "medium"
    modality: Modality
    """The modality of the document to load."""

    image_resolution_dpi: int = 96
    browser_canvas: BrowserCanvas = "A4"
    """The image operations to apply to the document."""

    stream: bool = False
    """Whether to stream the response."""

    tools_config: EnhanceSchemaConfig = Field(
        default_factory=EnhanceSchemaConfig,
        description="The configuration for the tools to use",
    )

    json_schema: dict[str, Any]
    instructions: str | None = None
    flat_likelihoods: list[dict[str, float]] | dict[str, float] | None = None
