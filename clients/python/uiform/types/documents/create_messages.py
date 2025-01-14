from pydantic import BaseModel, Field, model_validator
from typing import cast, Iterable, Optional, Self, TypedDict, Literal, Union, List, Any

import base64
import PIL.Image
import requests
import logging
from io import BytesIO

from ..modalities import Modality
from ..._utils.ai_model import find_provider_from_model
from ..mime import BaseMIMEData, MIMEData
from ..ai_model import AIProvider
from .text_operations import TextOperations

from openai.types.chat.chat_completion_message_param import ChatCompletionMessageParam
from openai.types.chat.chat_completion_content_part_param import ChatCompletionContentPartParam
from openai.types.chat.chat_completion_content_part_text_param import ChatCompletionContentPartTextParam
from openai.types.chat.chat_completion_content_part_image_param import ChatCompletionContentPartImageParam
from openai.types.chat.chat_completion_content_part_input_audio_param import ChatCompletionContentPartInputAudioParam

from google.generativeai.types import content_types  # type: ignore
from google.generativeai.types import generation_types

from anthropic.types.message_param import MessageParam
from anthropic.types.content_block import ContentBlock
from anthropic.types.text_block_param import TextBlockParam
from anthropic.types.image_block_param import ImageBlockParam, Source
from anthropic.types.tool_use_block_param import ToolUseBlockParam
from anthropic.types.tool_result_block_param import ToolResultBlockParam
from anthropic._types import NotGiven, NOT_GIVEN

BlockUnion = Union[TextBlockParam, ImageBlockParam, ToolUseBlockParam, ToolResultBlockParam, ContentBlock]
ContentBlockAnthropicUnion = Union[str, Iterable[BlockUnion]]
MediaType = Literal["image/jpeg", "image/png", "image/gif", "image/webp"]

class ChatCompletionUiformMessage(TypedDict):  # homemade replacement for ChatCompletionMessageParam because iterable messes the serialization with pydantic
    role: Literal['user', 'system', 'assistant']
    content : Union[str, list[ChatCompletionContentPartParam]]

def convert_to_google_genai_format(
    messages: List[ChatCompletionUiformMessage]
) -> content_types.ContentsType:
    """
    Converts a list of ChatCompletionUiFormMessage to a format compatible with the google.genai SDK.


    Example:
        ```python
        import google.generativeai as genai
        
        # Configure the Gemini client
        genai.configure(api_key=os.environ["GEMINI_API_KEY"])
        
        # Initialize the model
        model = genai.GenerativeModel("gemini-2.0-flash")
        
        # Get messages in Gemini format
        gemini_messages = document_message.gemini_messages
        
        # Generate a response
        ```

    Args:
        messages (List[ChatCompletionUiformMessage]): List of chat messages.

    Returns:
        List[Union[Dict[str, str], str]]: A list of formatted inputs for the google.genai SDK.
    """

    formatted_inputs: content_types.ContentsType = []
    
    for message in messages:
        if isinstance(message['content'], str):
            # Direct string content is treated as the prompt for the SDK
            formatted_inputs.append(message['content'])
        elif isinstance(message['content'], list):
            # Handle structured content
            for part in message['content']:
                if 'text' in part : 
                    if isinstance(part.get('text', None), str):
                        # If the part has a text key, treat it as plain text
                        formatted_inputs.append(part['text'])  # type: ignore
                if 'data' in part: 
                    if isinstance(part.get('data', None), bytes):
                        # Handle binary data
                        formatted_inputs.append({
                            "mime_type": part.get("mime_type", "application/octet-stream"), # type: ignore
                            "data": base64.b64encode(part["data"]).decode('utf-8') # type: ignore
                        })
                elif 'data' in part: 
                    if isinstance(part.get('data', None), str):
                        # Handle string data with a mime_type
                        formatted_inputs.append({
                            "mime_type": part.get("mime_type", "text/plain"), # type: ignore
                            "data": part["data"] # type: ignore
                        })
                if part.get('type') == 'image_url':
                    if 'image_url' in part:
                        # Handle image URLs containing base64-encoded data
                        url = part['image_url'].get('url', '')  # type: ignore
                        if url.startswith('data:image/jpeg;base64,'):
                            # Extract base64 data and add it to the formatted inputs
                            base64_data = url.replace('data:image/jpeg;base64,', '')
                            formatted_inputs.append({
                                "mime_type": "image/jpeg",
                                "data": base64_data # type: ignore
                            })
    
    return formatted_inputs
    
def convert_to_anthropic_format(messages: List[ChatCompletionUiformMessage]) -> tuple[str | NotGiven, List[MessageParam]]:
    """
    Converts a list of ChatCompletionUiformMessage to a format compatible with the Anthropic SDK.

    Args:
        messages (List[ChatCompletionUiformMessage]): List of chat messages.

    Returns:
        List[dict]: A list of formatted content blocks for the Anthropic SDK.
    """
    formatted_messages: list[MessageParam] = []
    system_message: str | NotGiven = NOT_GIVEN
    
    for message in messages:
        content_blocks: list[Union[TextBlockParam, ImageBlockParam]] = []
        if message["role"] == "system":
            assert isinstance(message["content"], str), "System message content must be a string."
            if system_message is not NOT_GIVEN:
                raise ValueError("Only one system message is allowed per chat.")
            else:
                system_message = message["content"]
            continue

        
        if isinstance(message['content'], str):
            # Direct string content is treated as text block
            content_blocks.append({
                "type": "text",
                "text": message['content']
            })
        elif isinstance(message['content'], list):
            # Handle structured content
            for part in message['content']:
                if part["type"] == "text":
                    part = cast(ChatCompletionContentPartTextParam, part)
                    # Text content
                    content_blocks.append({
                        "type": "text",
                        "text": part['text'] # type: ignore
                    })
                elif part["type"] == "image_url":
                    part = cast(ChatCompletionContentPartInputAudioParam, part)
                    # TODO: Audio Input not supported yet
                    logging.warning("Audio input is not supported yet.")

                elif part["type"] == "image_url":
                    # Two cases: Either is a URL or a base64 encoded image
                    part = cast(ChatCompletionContentPartImageParam, part)
                    base64_url: str | None = None
                    if "base64" in part["image_url"]["url"]:
                        base64_url = part["image_url"]["url"]
                    else:
                        try:
                            base64_url = base64.b64encode(requests.get(part["image_url"]["url"]).content).decode('utf-8')
                        except Exception:
                            logging.warning("Failed to load image from URL: %s", part["image_url"]["url"], exc_info=True, stack_info=True)
                    if base64_url:
                        # Base64 encoded image
                        media_type, data_content = base64_url.split(":data")[0].split(";base64,")
                        content_blocks.append(
                            {
                                "type": "image",
                                "source": {
                                    "type": "base64",
                                    "media_type": cast(MediaType, media_type),
                                    "data": data_content
                                }
                            }
                        )


        
        formatted_messages.append(MessageParam(
            role=message["role"],  # type: ignore
            content=content_blocks
        ))
    
    return system_message, formatted_messages

def convert_to_openai_format(messages: List[ChatCompletionUiformMessage]) -> List[ChatCompletionMessageParam]:
    return cast(list[ChatCompletionMessageParam], messages)
    

def str_messages(messages: list[ChatCompletionUiformMessage], max_length: int = 100) -> str:
    """
    Converts a list of chat messages into a string representation with faithfully serialized structure.

    Args:
        messages (list[ChatCompletionUiformMessage]): The list of chat messages.
        max_length (int): Maximum length for content before truncation.

    Returns:
        str: A string representation of the messages with applied truncation.
    """
    def truncate(text: str, max_len: int) -> str:
        """Truncate text to max_len with ellipsis."""
        return text if len(text) <= max_len else f"{text[:max_len]}..."

    serialized: list[ChatCompletionUiformMessage] = []
    for message in messages:
        role = message["role"]
        content = message["content"]

        if isinstance(content, str):
            serialized.append({"role": role, "content": truncate(content, max_length)})
        elif isinstance(content, list):
            truncated_content: list[ChatCompletionContentPartParam] = []
            for part in content:
                if part["type"] == "text" and part["text"]:
                    truncated_content.append({"type": "text", "text": truncate(part["text"], max_length)})
                elif part["type"] == "image_url" and part["image_url"]:
                    image_url = part["image_url"].get("url", "unknown image")
                    truncated_content.append({"type": "image_url", "image_url": {"url": truncate(image_url, max_length)}})
            serialized.append({"role": role, "content": truncated_content})

    return repr(serialized)

class DocumentCreateMessageRequest(BaseModel):
    document: MIMEData
    """The document to load."""

    modality: Modality
    """The modality of the document to load."""

    model: str | None = None
    """The model to use for loading the document."""

    provider: AIProvider = "OpenAI"
    """The AI provider to use for loading the document."""

    text_operations: TextOperations | None = Field(default=None, description="Additional context to be used by the AI model", examples=[{
        "regex_instructions": [{
            "name": "VAT Number",
            "description": "All potential VAT numbers in the documents",
            "pattern": r"[Ff][Rr]\s*(\d\s*){11}"
        }],
    }])
    """The text operations to apply to the document."""

    @model_validator(mode="after")
    def validate_model_and_provider(self) -> Self:
        """Validate and finalize model and provider values."""
        # If model is provided, infer provider if not explicitly set
        if self.model:
            inferred_provider = find_provider_from_model(self.model)
            if self.provider == "OpenAI" and self.provider != inferred_provider:
                self.provider = inferred_provider
            elif self.provider != inferred_provider:
                raise ValueError(
                    f"Provided provider '{self.provider}' does not match "
                    f"the provider '{inferred_provider}' inferred from model '{self.model}'."
                )

        return self
    
    def __str__(self)->str:
        return f"DocumentCreateMessageRequest(document={self.document}, model={self.model}, modality={self.modality}, provider={self.provider})"
    
    def __repr__(self)->str:
        return f"DocumentCreateMessageRequest(document={self.document}, model={self.model}, modality={self.modality}, provider={self.provider})"


class DocumentMessage(BaseModel):
    id: str
    """A unique identifier for the document loading."""

    object: Literal["document.message"] = Field(default="document.message")
    """The type of object being loaded."""

    messages: List[ChatCompletionUiformMessage] 
    """A list of messages containing the document content and metadata."""

    created: int
    """The Unix timestamp (in seconds) of when the document was loaded."""

    modality: Modality
    """The modality of the document to load."""
    
    document: BaseMIMEData
    """The document being loaded."""

    @property
    def items(self) -> list[str | PIL.Image.Image]:
        """Returns the document contents as a list of strings and images.

        This property processes the message content and converts it into a list of either
        text strings or PIL Image objects. It handles various content types including:
        - Plain text
        - Base64 encoded images
        - Remote image URLs
        - Audio data (represented as truncated string)

        Returns:
            list[str | PIL.Image.Image]: A list containing either strings for text content
                or PIL.Image.Image objects for image content. Failed image loads will
                return their URLs as strings instead.
        """
        results: list[str | PIL.Image.Image] = []
        
        for msg in self.messages:
            for content_item in msg["content"]:
                if isinstance(content_item, str):
                    results.append(content_item)
                else:
                    item_type = content_item.get("type")
                    # If item is an image
                    if item_type == "image_url":
                        assert "image_url" in content_item, "image_url is required in ChatCompletionContentPartImageParam"
                        image_data_url = content_item["image_url"]["url"] # type: ignore

                        # 1) Base64 inline data
                        if image_data_url.startswith("data:image/"):
                            try:
                                prefix, base64_part = image_data_url.split(",", 1)
                                img_bytes = base64.b64decode(base64_part)
                                img = PIL.Image.open(BytesIO(img_bytes))
                                results.append(img)
                            except Exception as e:
                                print(f"Error decoding base64 data:\n  {e}")
                                results.append(image_data_url)

                        # 2) Otherwise, assume it's a remote URL
                        else:
                            try:
                                response = requests.get(image_data_url)
                                response.raise_for_status()  # raises HTTPError if not 200
                                img = PIL.Image.open(BytesIO(response.content))
                                results.append(img)
                            except Exception as e:
                                # Here, log or print the actual error
                                print(f"Could not download image from {image_data_url}:\n  {e}")
                                results.append(image_data_url)

                    # If item is text (or other types)
                    elif item_type == "text":
                        text_value = content_item.get("text", "")
                        assert isinstance(text_value, str), "text is required in ChatCompletionContentPartTextParam"
                        results.append(text_value)

                    elif item_type == "input_audio":
                        # Handle audio input content
                        if "input_audio" in content_item:
                            audio_data = content_item["input_audio"]["data"] # type: ignore
                            results.append(f"Audio data: {audio_data[:100]}...")  # Truncate long audio data

                    else:
                        # Fallback for unrecognized item types
                        results.append(f"Unrecognized type: {item_type}")

        return results

    @property
    def openai_messages(self) -> list[ChatCompletionMessageParam]:
        """Returns the messages formatted for OpenAI's API.

        Converts the internal message format to OpenAI's expected format for
        chat completions.

        Returns:
            list[ChatCompletionMessageParam]: Messages formatted for OpenAI's chat completion API.
        """
        return convert_to_openai_format(self.messages)

    @property
    def anthropic_system_prompt(self) -> str | NotGiven:
        """Returns the system prompt formatted for Anthropic's Claude API.

        Extracts and formats the system message for use with Claude. If no system
        message exists, returns NOT_GIVEN.

        Returns:
            str | NotGiven: The system prompt if present, otherwise NOT_GIVEN.
        """
        return convert_to_anthropic_format(self.messages)[0]

    @property
    def anthropic_messages(self) -> list[MessageParam]:
        """Returns the messages formatted for Anthropic's Claude API.

        Converts the internal message format to Claude's expected format,
        handling text, images, and other content types appropriately.

        Returns:
            list[MessageParam]: Messages formatted for Claude's API.
        """
        return convert_to_anthropic_format(self.messages)[1]
    
    @property
    def gemini_messages(self) -> content_types.ContentsType:
        """Returns the messages formatted for Google's Gemini API.

        Converts the internal message format to Gemini's expected format,
        handling various content types including text and images.

        Returns:
            content_types.ContentsType: Messages formatted for Gemini's API.
        """
        return convert_to_google_genai_format(self.messages)

    
    def __str__(self)->str:
        return f"DocumentMessage(id={self.id}, object={self.object}, created={self.created}, messages={str_messages(self.messages)}, modality={self.modality}, document={self.document})"
    
    def __repr__(self)->str:
        return f"DocumentMessage(id={self.id}, object={self.object}, created={self.created}, messages={str_messages(self.messages)}, modality={self.modality}, document={self.document})"
    
