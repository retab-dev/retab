import base64
import json
from typing import Any, Literal, Optional
import datetime

from openai.types.chat import ChatCompletionMessageParam
from openai.types.chat.chat_completion import ChatCompletion
from openai.types.chat.chat_completion_chunk import ChatCompletionChunk
from openai.types.chat.chat_completion_chunk import Choice as ChoiceChunk
from openai.types.chat.chat_completion_chunk import ChoiceDelta as ChoiceDeltaChunk
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from openai.types.chat.parsed_chat_completion import ParsedChatCompletion, ParsedChoice
from openai.types.responses.response import Response
from openai.types.responses.response_input_param import ResponseInputItemParam
from openai.types.chat.parsed_chat_completion import ParsedChatCompletionMessage
from pydantic import BaseModel, ConfigDict, Field, ValidationInfo, field_validator, model_validator
from ..chat import ChatCompletionRetabMessage
from ..mime import MIMEData
from ..standards import StreamingBaseModel
from ...utils.json_schema import filter_auxiliary_fields_json, convert_basemodel_to_partial_basemodel, convert_json_schema_to_basemodel, unflatten_dict

class DocumentExtractRequest(BaseModel):
    model_config = ConfigDict(arbitrary_types_allowed=True, extra="ignore")
    document: MIMEData = Field(..., description="Document to be analyzed")
    image_resolution_dpi: int = Field(default=192, description="Resolution of the image sent to the LLM", ge=96, le=300)
    model: str = Field(..., description="Model used for chat completion")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])
    reasoning_effort: ChatCompletionReasoningEffort = Field(
        default="minimal", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )
    n_consensus: int = Field(default=1, description="Number of consensus models to use for extraction. If greater than 1 the temperature cannot be 0.")
    # Regular fields
    stream: bool = Field(default=False, description="If true, the extraction will be streamed to the user using the active WebSocket connection")
    seed: int | None = Field(default=None, description="Seed for the random number generator. If not provided, a random seed will be generated.", examples=[None])
    store: bool = Field(default=True, description="If true, the extraction will be stored in the database")
    chunking_keys: Optional[dict[str, str]] = Field(default=None, description="If set, keys to be used for the extraction of long lists of data using Parallel OCR", examples=[{"properties": "ID", "products": "identity.id"}])
    web_search: bool = Field(default=False, description="Enable web search enrichment with Parallel AI to add external context during extraction")
    metadata: dict[str, str] = Field(default_factory=dict, description="User-defined metadata to associate with this extraction")
    extraction_id: Optional[str] = Field(default=None, description="Extraction ID to use for this extraction. If not provided, a new ID will be generated.")
    additional_messages: Optional[list[ChatCompletionRetabMessage]] = Field(default=None, description="Additional chat messages to append after the document content messages. Useful for providing extra context or instructions.")

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
        default="minimal", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )


class RetabParsedChoice(ParsedChoice):
    # Adaptable ParsedChoice that allows None for the finish_reason
    finish_reason: Literal["stop", "length", "tool_calls", "content_filter", "function_call"] | None = None  # type: ignore
    key_mapping: dict[str, Optional[str]] | None = Field(default=None, description="Mapping of consensus keys to original model keys")


LikelihoodsSource = Literal["consensus", "log_probs"]

class RetabParsedChatCompletion(ParsedChatCompletion):
    model_config = ConfigDict(arbitrary_types_allowed=True, extra="ignore")
    extraction_id: str | None = None
    choices: list[RetabParsedChoice]  # type: ignore
    # Additional metadata fields
    likelihoods: Optional[dict[str, Any]] = Field(
        default=None, description="Object defining the uncertainties of the fields extracted when using consensus. Follows the same structure as the extraction object."
    )

    requires_human_review: bool = Field(default=False, description="Flag indicating if the extraction requires human review")

    request_at: datetime.datetime | None = Field(default=None, description="Timestamp of the request")
    first_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the first token of the document. If non-streaming, set to last_token_at")
    last_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the last token of the document")


class LogExtractionRequest(BaseModel):
    messages: list[ChatCompletionRetabMessage] | None = None  # TODO: compatibility with Anthropic
    openai_messages: list[ChatCompletionMessageParam] | None = None
    openai_responses_input: list[ResponseInputItemParam] | None = None

    document: MIMEData = Field(
        default=MIMEData(
            filename="dummy.txt",
            # url is a base64 encoded string with the mime type and the content. For the dummy one we will send a .txt file with the text "No document provided"
            url="data:text/plain;base64," + base64.b64encode(b"No document provided").decode("utf-8"),
        ),
        description="Document analyzed, if not provided a dummy one will be created with the text 'No document provided'",
    )
    completion: dict | RetabParsedChatCompletion | ParsedChatCompletion | ChatCompletion | None = None
    openai_responses_output: Response | None = None
    json_schema: dict[str, Any]
    model: str
    temperature: float

    # Validate that at least one of the messages, openai_messages is provided using model_validator
    @model_validator(mode="before")
    def validation(cls, data: Any) -> Any:
        # Handle both dict and model instance cases
        if isinstance(data, dict):
            messages_candidates = [data.get("messages"), data.get("openai_messages"), data.get("openai_responses_input")]
            messages_candidates = [candidate for candidate in messages_candidates if candidate is not None]
            if len(messages_candidates) != 1:
                raise ValueError("Exactly one of the messages, openai_messages, openai_responses_input must be provided")

            completion_candidates = [data.get("completion"), data.get("openai_responses_output")]
            completion_candidates = [candidate for candidate in completion_candidates if candidate is not None]
            if len(completion_candidates) != 1:
                raise ValueError("Exactly one of completion, openai_responses_output must be provided")
        else:
            # Handle model instance case
            messages_candidates = [
                getattr(data, "messages", None),
                getattr(data, "openai_messages", None),
                getattr(data, "openai_responses_input", None),
            ]
            messages_candidates = [candidate for candidate in messages_candidates if candidate is not None]
            if len(messages_candidates) != 1:
                raise ValueError("Exactly one of the messages, openai_messages, openai_responses_input must be provided")

            completion_candidates = [getattr(data, "completion", None), getattr(data, "openai_responses_output", None)]
            completion_candidates = [candidate for candidate in completion_candidates if candidate is not None]
            if len(completion_candidates) != 1:
                raise ValueError("Exactly one of completion, openai_responses_output must be provided")

        return data


class LogExtractionResponse(BaseModel):
    status: Literal["success", "error"]
    error_message: str | None = None


# DocumentExtractResponse = RetabParsedChatCompletion


###### I'll place here for now -- New Streaming API


# We build from the openai.types.chat.chat_completion_chunk.ChatCompletionChunk adding just two three additional fields:
# - is_valid_json: list[bool]               #  Whether the total accumulated content is a valid JSON
# - likelihoods: dict[str, float]     #  The delta of the flattened likelihoods (to be merged with the cumulated likelihoods)
# - schema_validation_error: ErrorDetail | None = None #  The error in the schema validation of the total accumulated content


class RetabParsedChoiceDeltaChunk(ChoiceDeltaChunk):
    flat_likelihoods: dict[str, float] = {}
    flat_parsed: dict[str, Any] = {}
    flat_deleted_keys: list[str] = []
    is_valid_json: bool = False
    key_mapping: dict[str, Optional[str]] | None = Field(default=None, description="Mapping of consensus keys to original model keys")


class RetabParsedChoiceChunk(ChoiceChunk):
    delta: RetabParsedChoiceDeltaChunk  # type: ignore


class RetabParsedChatCompletionChunk(StreamingBaseModel, ChatCompletionChunk):
    choices: list[RetabParsedChoiceChunk]  # type: ignore

    extraction_id: str | None = None
    request_at: datetime.datetime | None = Field(default=None, description="Timestamp of the request")
    first_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the first token of the document. If non-streaming, set to last_token_at")
    last_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the last token of the document")

    def chunk_accumulator(self, previous_cumulated_chunk: "RetabParsedChatCompletionChunk | None" = None) -> "RetabParsedChatCompletionChunk":
        """
        Accumulate the chunk into the state, returning a new RetabParsedChatCompletionChunk with the accumulated content that could be yielded alone to generate the same state.
        """

        def safe_get_delta(chnk: "RetabParsedChatCompletionChunk | None", index: int) -> RetabParsedChoiceDeltaChunk:
            if chnk is not None and index < len(chnk.choices):
                return chnk.choices[index].delta
            else:
                return RetabParsedChoiceDeltaChunk(
                    content="",
                    flat_parsed={},
                    flat_likelihoods={},
                    is_valid_json=False,
                )

        max_choices = max(len(self.choices), len(previous_cumulated_chunk.choices)) if previous_cumulated_chunk is not None else len(self.choices)

        # Get the current chunk missing content, flat_deleted_keys and is_valid_json
        acc_flat_deleted_keys = [safe_get_delta(self, i).flat_deleted_keys for i in range(max_choices)]
        acc_is_valid_json = [safe_get_delta(self, i).is_valid_json for i in range(max_choices)]
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

        return RetabParsedChatCompletionChunk(
            id=self.id,
            created=self.created,
            model=self.model,
            object=self.object,
            usage=None,
            choices=[
                RetabParsedChoiceChunk(
                    delta=RetabParsedChoiceDeltaChunk(
                        content=acc_content[i],
                        flat_parsed=acc_flat_parsed[i],
                        flat_likelihoods=acc_flat_likelihoods[i],
                        flat_deleted_keys=acc_flat_deleted_keys[i],
                        is_valid_json=acc_is_valid_json[i],
                        key_mapping=acc_key_mapping[i],
                    ),
                    index=i,
                )
                for i in range(max_choices)
            ],
            extraction_id=self.extraction_id,
            request_at=self.request_at,
            first_token_at=self.first_token_at,
            last_token_at=self.last_token_at,
        )

    def to_completion(
        self,
        override_final_flat_likelihoods: dict[str, Any] | None = None,
        override_final_flat_parseds: list[dict[str, Any]] | None = None,
    ) -> RetabParsedChatCompletion:
        if override_final_flat_parseds is None:
            override_final_flat_parseds = [self.choices[idx].delta.flat_parsed for idx in range(len(self.choices))]

        final_parsed_list = [unflatten_dict(override_final_flat_parseds[idx]) for idx in range(len(self.choices))]
        final_content_list = [json.dumps(final_parsed_list[idx]) for idx in range(len(self.choices))]

        # The final likelihoods are only on the first choice.
        if override_final_flat_likelihoods is None:
            override_final_flat_likelihoods = self.choices[0].delta.flat_likelihoods

        final_likelihoods = unflatten_dict(override_final_flat_likelihoods)

        return RetabParsedChatCompletion(
            id=self.id,
            extraction_id=self.extraction_id,
            request_at=self.request_at,
            first_token_at=self.first_token_at,
            last_token_at=self.last_token_at,
            created=self.created,
            model=self.model,
            object="chat.completion",
            choices=[
                RetabParsedChoice(
                    index=idx,
                    message=ParsedChatCompletionMessage(
                        content=final_content_list[idx],
                        role="assistant",
                        parsed=final_parsed_list[idx],
                    ),
                    key_mapping=self.choices[idx].delta.key_mapping,
                    finish_reason="stop",
                    logprobs=None,
                )
                for idx in range(len(self.choices))
            ],
            likelihoods=final_likelihoods,
            usage=self.usage,
        )


def maybe_parse_to_pydantic(schema: dict[str, Any], response: RetabParsedChatCompletion, allow_partial: bool = False) -> RetabParsedChatCompletion:
    if response.choices[0].message.content:
        try:
            full_pydantic_model = convert_json_schema_to_basemodel(schema)
            if allow_partial:
                partial_pydantic_model = convert_basemodel_to_partial_basemodel(full_pydantic_model)
                response.choices[0].message.parsed = partial_pydantic_model.model_validate(filter_auxiliary_fields_json(response.choices[0].message.content))
            else:
                response.choices[0].message.parsed = full_pydantic_model.model_validate(filter_auxiliary_fields_json(response.choices[0].message.content))
        except Exception:
            # If parsing fails (e.g., due to invalid schema), set parsed to None
            # instead of leaving it as a raw dictionary
            response.choices[0].message.parsed = None
    return response
