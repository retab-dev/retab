from typing import Any, Literal, Self, TypedDict, Literal
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, model_validator

from ..mime import MIMEData
from ..modalities import Modality


class EnhanceSchemaConfig(BaseModel):
    allow_fields_insertion: bool = False  # Whether to allow the llm to insert new fields
    allow_fields_deletion: bool = False  # Whether to allow the llm to delete existing fields
    allow_field_description_update: bool = False  # Whether to allow the llm to update the description of existing fields
    allow_reasoning_field_toggle: bool = False  # Whether to allow the llm to toggle the reasoning for fields
    allow_system_prompt_update: bool = True  # Whether to allow the llm to update the system prompt

    # Model validator
    @model_validator(mode="after")
    def check_at_least_one_tool_allowed(self) -> Self:
        if not any(
            [self.allow_fields_insertion, self.allow_fields_deletion, self.allow_field_description_update, self.allow_reasoning_field_toggle, self.allow_system_prompt_update]
        ):
            raise ValueError("At least one tool must be allowed")
        # There are some not implemented tools
        if any([self.allow_fields_insertion, self.allow_fields_deletion, self.allow_reasoning_field_toggle]):
            raise ValueError("Fields insertion, Field deletion, Field description update and Reasoning field toggle are not implemented yet, sorry about that!")
        return self


# Define a typed Dict for EnhanceSchemaConfig (for now it is kind static, but we will add more flexibility in the future)
class EnhanceSchemaConfigDict(TypedDict, total=False):
    allow_fields_insertion: Literal[False]
    allow_fields_deletion: Literal[False]
    allow_reasoning_field_toggle: Literal[False]
    allow_field_description_update: bool
    allow_system_prompt_update: bool


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
    browser_canvas: Literal['A3', 'A4', 'A5'] = 'A4'
    """The image operations to apply to the document."""

    stream: bool = False
    """Whether to stream the response."""

    tools_config: EnhanceSchemaConfig = Field(default_factory=EnhanceSchemaConfig, description="The configuration for the tools to use")

    json_schema: dict[str, Any]
    instructions: str | None = None
    # flat_likelihoods: list[dict[str, float]] | dict[str, float] | None = None
