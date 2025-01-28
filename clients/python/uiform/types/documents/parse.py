from pydantic import BaseModel, Field, PrivateAttr, ConfigDict, field_validator, ValidationInfo
from typing import Any, Generic, Literal, List
from functools import cached_property
from openai.types.chat.chat_completion import Choice, ChatCompletion
from openai.types.chat.chat_completion_message import ChatCompletionMessage
from openai.types.chat.parsed_chat_completion import ParsedChatCompletion, ContentType

from .text_operations import TextOperations, RegexInstructionResult
from .create_messages import ChatCompletionUiformMessage, DocumentProcessingConfig
from .image_operations import ImageOperations
from ..modalities import Modality
from ..ai_model import AIProvider, LLMModel
from ..standards import ErrorDetail, StreamingBaseModel
from ..schemas.object import Schema
from ..._utils.ai_model import find_provider_from_model
from ..._utils.mime import generate_sha_hash_from_base64
from ..mime import MIMEData, BaseMIMEData

import datetime

class DocumentExtractionConfig(DocumentProcessingConfig):
    # Redeclaration of the DocumentProcessingConfig
    modality: Modality
    text_operations: TextOperations = Field(default_factory=TextOperations, description="Additional context to be used by the AI model")
    image_operations : ImageOperations = Field(default_factory=ImageOperations, description="Preprocessing operations applied to image before sending them to the llm")

    # New attributes
    model: LLMModel = Field(..., description="Model used for chat completion")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])
    additional_messages: list[ChatCompletionUiformMessage] = Field(default=[], description="Additional messages to be used by the AI model")


class DocumentExtractRequest(DocumentExtractionConfig):
    # Attributes
    stream: bool = Field(default=False, description="If true, the extraction will be streamed to the user using the active WebSocket connection")
    seed: int | None = Field(default=None, description="Seed for the random number generator. If not provided, a random seed will be generated.", examples=[None])
    store: bool = Field(default=False, description="If true, the extraction will be stored in the database")
    
    document: MIMEData = Field(..., description="Document to be analyzed")


    # Some internal attributes (hidden from serialization)
    _regex_instruction_results: list[RegexInstructionResult] | None = PrivateAttr(default=None)

    # Some properties (hidden from serialization)
    @property
    def provider(self) -> AIProvider:
        """
        Determines the AI provider based on the model specified.

        Returns:
            AIProvider: The AI provider corresponding to the given model.
        """
        return find_provider_from_model(self.model)
    
    # Some cached property, they are not exposed when serializing the object. They are computed once and stored in the object.
    @cached_property
    def form_schema(self) -> Schema:
        """
        Generates a Schema object from the JSON schema definition.
        This property is cached after first computation.

        Returns:
            Schema: A Schema object representing the form's structure.

        Raises:
            AssertionError: If the json_schema is empty or None.
        """
        assert self.json_schema, "The response format schema cannot be empty."
        return Schema(json_schema=self.json_schema)

   

class BaseDocumentExtractRequest(DocumentExtractRequest):
    document: BaseMIMEData = Field(..., description="Document analyzed (without content, for MongoDB storage)")     # type: ignore

    model_config = ConfigDict(
        json_schema_extra={"document": {"metadata": False}}
    )

    # Add a field validator that converts the input document to BaseMIMEData
    @field_validator("document")
    def convert_document_to_base_mimedata(cls, v: MIMEData | BaseMIMEData, validation_info: ValidationInfo) -> BaseMIMEData:
        return BaseMIMEData(**v.model_dump())
    
class UiParsedChatCompletionMessage(ChatCompletionMessage):
    parsed: BaseModel | None = None

class UiParsedChoice(Choice):
    message: UiParsedChatCompletionMessage
    finish_reason: Literal["stop", "length", "tool_calls", "content_filter", "function_call"] | None = None     # type: ignore


class UiParsedChatCompletion(ParsedChatCompletion): 
    choices: List[UiParsedChoice]   # type: ignore[assignment]

    # Additional metadata fields (UIForm)
    likelihoods: Any # Object defining the uncertainties of the fields extracted. Follows the same structure as the extraction object.
    regex_instruction_results: list[RegexInstructionResult] | None = None
    schema_validation_error: ErrorDetail | None = None
    
    # Timestamps
    request_at: datetime.datetime | None = Field(default=None, description="Timestamp of the request")
    first_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the first token of the document. If non-streaming, set to last_token_at")
    last_token_at: datetime.datetime | None = Field(default=None, description="Timestamp of the last token of the document")


class UiParsedChatCompletionStream(StreamingBaseModel, UiParsedChatCompletion): pass

DocumentExtractResponse = UiParsedChatCompletion
DocumentExtractResponseStream = UiParsedChatCompletionStream
