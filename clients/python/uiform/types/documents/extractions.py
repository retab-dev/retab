import base64
import datetime
from typing import Any, Literal, Optional

from anthropic.types.message import Message
from anthropic.types.message_param import MessageParam
from openai.types.chat import ChatCompletionMessageParam
from openai.types.chat.chat_completion import ChatCompletion
from openai.types.chat.chat_completion_chunk import ChatCompletionChunk
from openai.types.chat.chat_completion_chunk import Choice as ChoiceChunk
from openai.types.chat.chat_completion_chunk import ChoiceDelta as ChoiceDeltaChunk
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from openai.types.chat.parsed_chat_completion import ParsedChatCompletion, ParsedChoice
from openai.types.responses.response import Response
from openai.types.responses.response_input_param import ResponseInputItemParam
from pydantic import BaseModel, ConfigDict, Field, ValidationInfo, computed_field, field_validator, model_validator

from ..._utils.usage.usage import compute_cost_from_model, compute_cost_from_model_with_breakdown, CostBreakdown

from ..._utils.ai_models import find_provider_from_model
from ..ai_models import AIProvider, Amount, get_model_card
from ..chat import ChatCompletionUiformMessage
from ..mime import MIMEData
from ..modalities import Modality
from ..standards import ErrorDetail, StreamingBaseModel


class DocumentExtractRequest(BaseModel):
    model_config = ConfigDict(arbitrary_types_allowed=True)

    document: MIMEData = Field(..., description="Document to be analyzed")
    modality: Modality
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: Literal['A3', 'A4', 'A5'] = Field(default='A4', description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type.")
    model: str = Field(..., description="Model used for chat completion")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])
    reasoning_effort: ChatCompletionReasoningEffort = Field(
        default="medium", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )
    n_consensus: int = Field(default=1, description="Number of consensus models to use for extraction. If greater than 1 the temperature cannot be 0.")
    # Regular fields
    stream: bool = Field(default=False, description="If true, the extraction will be streamed to the user using the active WebSocket connection")
    seed: int | None = Field(default=None, description="Seed for the random number generator. If not provided, a random seed will be generated.", examples=[None])
    store: bool = Field(default=True, description="If true, the extraction will be stored in the database")
    need_validation: bool = Field(default=False, description="If true, the extraction will be validated against the schema")

    # Add a model validator that rejects n_consensus > 1 if temperature is 0
    @field_validator("n_consensus")
    def check_n_consensus(cls, v: int, info: ValidationInfo) -> int:
        if v > 1 and info.data.get("temperature") == 0:
            raise ValueError("n_consensus greater than 1 but temperature is 0")
        return v


class ConsensusModel(BaseModel):
    model: str = Field(description="Model name")
    temperature: float = Field(default=0.0, description="Temperature for consensus")
    reasoning_effort: ChatCompletionReasoningEffort = Field(
        default="medium", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )



# For location of fields in the document (OCR)
class FieldLocation(BaseModel):
    label: str = Field(..., description="The label of the field")
    value: str = Field(..., description="The extracted value of the field")
    quote: str = Field(..., description="The quote of the field (verbatim from the document)")
    file_id: str | None = Field(default=None, description="The ID of the file")
    page: int | None = Field(default=None, description="The page number of the field (1-indexed)")
    bboxes_normalized: list[tuple[float, float, float, float]] | None = Field(default=None, description="The normalized bounding boxes of the field")
    score: float | None = Field(default=None, description="The score of the field")
    match_level: Literal["token", "line", "block"] | None = Field(default=None, description="The level of the match (token, line, block)")


class UiParsedChoice(ParsedChoice):
    # Adaptable ParsedChoice that allows None for the finish_reason
    finish_reason: Literal["stop", "length", "tool_calls", "content_filter", "function_call"] | None = None  # type: ignore
    field_locations: dict[str, list[FieldLocation]] | None = Field(default=None, description="The locations of the fields in the document, if available")
    key_mapping: dict[str, Optional[str]] | None = Field(default=None, description="Mapping of consensus keys to original model keys")


LikelihoodsSource = Literal["consensus", "log_probs"]


class UiParsedChatCompletion(ParsedChatCompletion):
    extraction_id: str | None = None
    choices: list[UiParsedChoice]
    # Additional metadata fields (UIForm)
    likelihoods: Optional[dict[str, Any]] = Field(
        default=None, description="Object defining the uncertainties of the fields extracted when using consensus. Follows the same structure as the extraction object."
    )
    schema_validation_error: ErrorDetail | None = None
    # Timestamps
    request_at: datetime.datetime | None = Field(default=None, description="Timestamp of the request")
    first_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the first token of the document. If non-streaming, set to last_token_at")
    last_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the last token of the document")

    @computed_field
    @property
    def api_cost(self) -> Optional[Amount]:
        if self.usage:
            try:
                cost = compute_cost_from_model(self.model, self.usage)
                return cost
            except Exception as e:
                print(f"Error computing cost: {e}")
                return None
        return None


class UiResponse(Response):
    extraction_id: str | None = None
    # Additional metadata fields (UIForm)
    likelihoods: Optional[dict[str, Any]] = Field(
        default=None, description="Object defining the uncertainties of the fields extracted when using consensus. Follows the same structure as the extraction object."
    )
    schema_validation_error: ErrorDetail | None = None
    # Timestamps
    request_at: datetime.datetime | None = Field(default=None, description="Timestamp of the request")
    first_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the first token of the document. If non-streaming, set to last_token_at")
    last_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the last token of the document")


class LogExtractionRequest(BaseModel):
    messages: list[ChatCompletionUiformMessage] | None = None  # TODO: compatibility with Anthropic
    openai_messages: list[ChatCompletionMessageParam] | None = None
    openai_responses_input: list[ResponseInputItemParam] | None = None
    anthropic_messages: list[MessageParam] | None = None
    anthropic_system_prompt: str | None = None
    document: MIMEData = Field(
        default=MIMEData(
            filename="dummy.txt",
            # url is a base64 encoded string with the mime type and the content. For the dummy one we will send a .txt file with the text "No document provided"
            url="data:text/plain;base64," + base64.b64encode(b"No document provided").decode("utf-8"),
        ),
        description="Document analyzed, if not provided a dummy one will be created with the text 'No document provided'",
    )
    completion: dict | UiParsedChatCompletion | Message | ParsedChatCompletion | ChatCompletion | None = None
    openai_responses_output: Response | None = None
    json_schema: dict[str, Any]
    model: str
    temperature: float

    # Validate that at least one of the messages, openai_messages, anthropic_messages is provided using model_validator
    @model_validator(mode="before")
    def validation(cls, data: Any) -> Any:
        messages_candidates = [data.get("messages"), data.get("openai_messages"), data.get("anthropic_messages"), data.get("openai_responses_input")]
        messages_candidates = [candidate for candidate in messages_candidates if candidate is not None]
        if len(messages_candidates) != 1:
            raise ValueError("Exactly one of the messages, openai_messages, anthropic_messages, openai_responses_input must be provided")

        # Validate that if anthropic_messages is provided, anthropic_system_prompt is also provided
        if data.get("anthropic_messages") is not None and data.get("anthropic_system_prompt") is None:
            raise ValueError("anthropic_system_prompt must be provided if anthropic_messages is provided")

        completion_candidates = [data.get("completion"), data.get("openai_responses_output")]
        completion_candidates = [candidate for candidate in completion_candidates if candidate is not None]
        if len(completion_candidates) != 1:
            raise ValueError("Exactly one of completion, openai_responses_output must be provided")

        return data


class LogExtractionResponse(BaseModel):
    extraction_id: str | None = None  # None only in case of error
    status: Literal["success", "error"]
    error_message: str | None = None


# DocumentExtractResponse = UiParsedChatCompletion


###### I'll place here for now -- New Streaming API


# We build from the openai.types.chat.chat_completion_chunk.ChatCompletionChunk adding just two three additional fields:
# - missing_content: list[str]              #  The missing content (to be concatenated to the cumulated content to create a potentially valid JSON), it should be aligned with the choices index
# - is_valid_json: list[bool]               #  Whether the total accumulated content is a valid JSON
# - likelihoods: dict[str, float]     #  The delta of the flattened likelihoods (to be merged with the cumulated likelihoods)
# - schema_validation_error: ErrorDetail | None = None #  The error in the schema validation of the total accumulated content


class UiParsedChoiceDeltaChunk(ChoiceDeltaChunk):
    flat_likelihoods: dict[str, float] = {}
    flat_parsed: dict[str, Any] = {}
    flat_deleted_keys: list[str] = []
    field_locations: dict[str, list[FieldLocation]] | None = Field(default=None, description="The locations of the fields in the document, if available")
    missing_content: str = ""
    is_valid_json: bool = False
    key_mapping: dict[str, Optional[str]] | None = Field(default=None, description="Mapping of consensus keys to original model keys")


class UiParsedChoiceChunk(ChoiceChunk):
    delta: UiParsedChoiceDeltaChunk


class UiParsedChatCompletionChunk(StreamingBaseModel, ChatCompletionChunk):
    extraction_id: str | None = None
    choices: list[UiParsedChoiceChunk]
    schema_validation_error: ErrorDetail | None = None
    # Timestamps
    request_at: datetime.datetime | None = Field(default=None, description="Timestamp of the request")
    first_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the first token of the document. If non-streaming, set to last_token_at")
    last_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the last token of the document")

    @computed_field
    @property
    def api_cost(self) -> Optional[Amount]:
        if self.usage:
            try:
                cost = compute_cost_from_model(self.model, self.usage)
                return cost
            except Exception as e:
                print(f"Error computing cost: {e}")
                return None
        return None

    @computed_field  # type: ignore
    @property
    def cost_breakdown(self) -> Optional[CostBreakdown]:
        if self.usage:
            try:
                cost = compute_cost_from_model_with_breakdown(self.model, self.usage)
                return cost
            except Exception as e:
                print(f"Error computing cost: {e}")
                return None
        return None

    def chunk_accumulator(self, previous_cumulated_chunk: "UiParsedChatCompletionChunk | None" = None) -> "UiParsedChatCompletionChunk":
        """
        Accumulate the chunk into the state, returning a new UiParsedChatCompletionChunk with the accumulated content that could be yielded alone to generate the same state.
        """

        def safe_get_delta(chnk: "UiParsedChatCompletionChunk | None", index: int) -> UiParsedChoiceDeltaChunk:
            if chnk is not None and index < len(chnk.choices):
                return chnk.choices[index].delta
            else:
                return UiParsedChoiceDeltaChunk(
                    content="",
                    flat_parsed={},
                    flat_likelihoods={},
                    missing_content="",
                    is_valid_json=False,
                )

        max_choices = max(len(self.choices), len(previous_cumulated_chunk.choices)) if previous_cumulated_chunk is not None else len(self.choices)

        # Get the current chunk missing content, flat_deleted_keys and is_valid_json
        acc_missing_content = [safe_get_delta(self, i).missing_content or "" for i in range(max_choices)]
        acc_flat_deleted_keys = [safe_get_delta(self, i).flat_deleted_keys for i in range(max_choices)]
        acc_is_valid_json = [safe_get_delta(self, i).is_valid_json for i in range(max_choices)]
        acc_field_locations = [safe_get_delta(self, i).field_locations for i in range(max_choices)]  # This is only present in the last chunk.
        # Delete from previous_cumulated_chunk.choices[i].delta.flat_parsed the keys that are in safe_get_delta(self, i).flat_deleted_keys
        for i in range(max_choices):
            previous_delta = safe_get_delta(previous_cumulated_chunk, i)
            current_delta = safe_get_delta(self, i)
            for deleted_key in current_delta.flat_deleted_keys:
                previous_delta.flat_parsed.pop(deleted_key, None)
                previous_delta.flat_likelihoods.pop(deleted_key, None)
        # Accumulate the flat_parsed and flat_likelihoods
        acc_flat_parsed = [safe_get_delta(previous_cumulated_chunk, i).flat_parsed | safe_get_delta(self, i).flat_parsed for i in range(max_choices)]
        acc_flat_likelihoods = [safe_get_delta(previous_cumulated_chunk, i).flat_likelihoods | safe_get_delta(self, i).flat_likelihoods for i in range(max_choices)]
        acc_key_mapping = [safe_get_delta(previous_cumulated_chunk, i).key_mapping or safe_get_delta(self, i).key_mapping for i in range(max_choices)]

        acc_content = [(safe_get_delta(previous_cumulated_chunk, i).content or "") + (safe_get_delta(self, i).content or "") for i in range(max_choices)]
        usage = self.usage
        first_token_at = self.first_token_at
        last_token_at = self.last_token_at
        request_at = self.request_at

        return UiParsedChatCompletionChunk(
            extraction_id=self.extraction_id,
            id=self.id,
            created=self.created,
            model=self.model,
            object=self.object,
            usage=usage,
            choices=[
                UiParsedChoiceChunk(
                    delta=UiParsedChoiceDeltaChunk(
                        content=acc_content[i],
                        flat_parsed=acc_flat_parsed[i],
                        flat_likelihoods=acc_flat_likelihoods[i],
                        flat_deleted_keys=acc_flat_deleted_keys[i],
                        field_locations=acc_field_locations[i],
                        missing_content=acc_missing_content[i],
                        is_valid_json=acc_is_valid_json[i],
                        key_mapping=acc_key_mapping[i],
                    ),
                    index=i,
                )
                for i in range(max_choices)
            ],
            schema_validation_error=self.schema_validation_error,
            request_at=request_at,
            first_token_at=first_token_at,
            last_token_at=last_token_at,
        )
