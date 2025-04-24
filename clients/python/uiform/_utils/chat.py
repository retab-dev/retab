import base64
import io
import logging
from typing import List, Literal, Optional, Union, cast

import requests
from anthropic.types.content_block import ContentBlock
from anthropic.types.image_block_param import ImageBlockParam, Source
from anthropic.types.message_param import MessageParam
from anthropic.types.text_block_param import TextBlockParam
from anthropic.types.tool_result_block_param import ToolResultBlockParam
from anthropic.types.tool_use_block_param import ToolUseBlockParam
from google.genai.types import BlobDict, ContentDict, ContentUnionDict, PartDict  # type: ignore
from openai.types.chat.chat_completion_content_part_image_param import ChatCompletionContentPartImageParam
from openai.types.chat.chat_completion_content_part_input_audio_param import ChatCompletionContentPartInputAudioParam
from openai.types.chat.chat_completion_content_part_param import ChatCompletionContentPartParam
from openai.types.chat.chat_completion_content_part_text_param import ChatCompletionContentPartTextParam
from openai.types.chat.chat_completion_message_param import ChatCompletionMessageParam
from PIL import Image

from ..types.chat import ChatCompletionUiformMessage

MediaType = Literal["image/jpeg", "image/png", "image/gif", "image/webp"]


def convert_to_google_genai_format(messages: List[ChatCompletionUiformMessage]) -> tuple[str, list[ContentUnionDict]]:
    """
    Converts a list of ChatCompletionUiFormMessage to a format compatible with the google.genai SDK.


    Example:
        ```python
        import google.genai as genai

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
    system_message: str = ""
    formatted_content: list[ContentUnionDict] = []
    for message in messages:
        # -----------------------
        # Handle system message
        # -----------------------
        if message["role"] in ("system", "developer"):
            assert isinstance(message["content"], str), "System message content must be a string."
            if system_message != "":
                raise ValueError("Only one system message is allowed per chat.")
            system_message += message["content"]
            continue
        parts: list[PartDict] = []

        message_content = message['content']
        if isinstance(message_content, str):
            # Direct string content is treated as the prompt for the SDK
            parts.append(PartDict(text=message_content))
        elif isinstance(message_content, list):
            # Handle structured content
            for part in message_content:
                if part["type"] == "text":
                    parts.append(PartDict(text=part["text"]))
                elif part["type"] == "image_url":
                    url = part['image_url'].get('url', '')  # type: ignore
                    if url.startswith('data:image'):
                        # Extract base64 data and add it to the formatted inputs
                        media_type, data_content = url.split(";base64,")
                        media_type = media_type.split("data:")[-1]  # => "image/jpeg"
                        base64_data = data_content

                        # Try to convert to PIL.Image and append it to the formatted inputs
                        try:
                            image_bytes = base64.b64decode(base64_data)
                            parts.append(PartDict(inline_data=BlobDict(data=image_bytes, mime_type=media_type)))
                        except Exception:
                            pass
                elif part["type"] == "input_audio":
                    pass
                elif part["type"] == "file":
                    pass
                else:
                    pass

        formatted_content.append(ContentDict(parts=parts, role=("user" if message["role"] == "user" else "model")))

    return system_message, formatted_content


def convert_to_anthropic_format(messages: List[ChatCompletionUiformMessage]) -> tuple[str, List[MessageParam]]:
    """
    Converts a list of ChatCompletionUiformMessage to a format compatible with the Anthropic SDK.

    Args:
        messages (List[ChatCompletionUiformMessage]): List of chat messages.

    Returns:
        (system_message, formatted_messages):
            system_message (str | NotGiven):
                The system message if one was found, otherwise NOT_GIVEN.
            formatted_messages (List[MessageParam]):
                A list of formatted messages ready for Anthropic.
    """

    formatted_messages: list[MessageParam] = []
    system_message: str = ""

    for message in messages:
        content_blocks: list[Union[TextBlockParam, ImageBlockParam]] = []

        # -----------------------
        # Handle system message
        # -----------------------
        if message["role"] in ("system", "developer"):
            assert isinstance(message["content"], str), "System message content must be a string."
            if system_message != "":
                raise ValueError("Only one system message is allowed per chat.")
            system_message += message["content"]
            continue

        # -----------------------
        # Handle non-system roles
        # -----------------------
        if isinstance(message['content'], str):
            # Direct string content is treated as a single text block
            content_blocks.append(
                {
                    "type": "text",
                    "text": message['content'],
                }
            )

        elif isinstance(message['content'], list):
            # Handle structured content
            for part in message['content']:
                if part["type"] == "text":
                    part = cast(ChatCompletionContentPartTextParam, part)
                    content_blocks.append(
                        {
                            "type": "text",
                            "text": part['text'],  # type: ignore
                        }
                    )

                elif part["type"] == "input_audio":
                    part = cast(ChatCompletionContentPartInputAudioParam, part)
                    logging.warning("Audio input is not supported yet.")
                    # No blocks appended since not supported

                elif part["type"] == "image_url":
                    # Handle images that may be either base64 data-URLs or standard remote URLs
                    part = cast(ChatCompletionContentPartImageParam, part)
                    image_url = part["image_url"]["url"]

                    if "base64," in image_url:
                        # The string is already something like: data:image/jpeg;base64,xxxxxxxx...
                        media_type, data_content = image_url.split(";base64,")
                        # media_type might look like: "data:image/jpeg"
                        media_type = media_type.split("data:")[-1]  # => "image/jpeg"
                        base64_data = data_content
                    else:
                        # It's a remote URL, so fetch, encode, and derive media type from headers
                        try:
                            r = requests.get(image_url)
                            r.raise_for_status()
                            content_type = r.headers.get("Content-Type", "image/jpeg")
                            # fallback "image/jpeg" if no Content-Type given

                            # Only keep recognized image/* for anthropic
                            if content_type not in ("image/jpeg", "image/png", "image/gif", "image/webp"):
                                logging.warning(
                                    "Unrecognized Content-Type '%s' - defaulting to image/jpeg",
                                    content_type,
                                )
                                content_type = "image/jpeg"

                            media_type = content_type
                            base64_data = base64.b64encode(r.content).decode("utf-8")

                        except Exception:
                            logging.warning(
                                "Failed to load image from URL: %s",
                                image_url,
                                exc_info=True,
                                stack_info=True,
                            )
                            # Skip adding this block if error
                            continue

                    # Finally, append to content blocks
                    content_blocks.append(
                        {
                            "type": "image",
                            "source": {
                                "type": "base64",
                                "media_type": cast(MediaType, media_type),
                                "data": base64_data,
                            },
                        }
                    )

        formatted_messages.append(
            MessageParam(
                role=message["role"],  # type: ignore
                content=content_blocks,
            )
        )

    return system_message, formatted_messages


def convert_from_anthropic_format(messages: list[MessageParam], system_prompt: str) -> list[ChatCompletionUiformMessage]:
    """
    Converts a list of Anthropic MessageParam to a list of ChatCompletionUiformMessage.
    """
    formatted_messages: list[ChatCompletionUiformMessage] = [ChatCompletionUiformMessage(role="developer", content=system_prompt)]

    for message in messages:
        role = message["role"]
        content_blocks = message["content"]

        # Handle different content structures
        if isinstance(content_blocks, list) and len(content_blocks) == 1 and isinstance(content_blocks[0], dict) and content_blocks[0].get("type") == "text":
            # Simple text message
            formatted_messages.append(cast(ChatCompletionUiformMessage, {"role": role, "content": content_blocks[0].get("text", "")}))
        elif isinstance(content_blocks, list):
            # Message with multiple content parts or non-text content
            formatted_content: list[ChatCompletionContentPartParam] = []

            for block in content_blocks:
                if isinstance(block, dict):
                    if block.get("type") == "text":
                        formatted_content.append(cast(ChatCompletionContentPartParam, {"type": "text", "text": block.get("text", "")}))
                    elif block.get("type") == "image":
                        source = block.get("source", {})
                        if isinstance(source, dict) and source.get("type") == "base64":
                            # Convert base64 image to data URL format
                            media_type = source.get("media_type", "image/jpeg")
                            data = source.get("data", "")
                            image_url = f"data:{media_type};base64,{data}"

                            formatted_content.append(cast(ChatCompletionContentPartParam, {"type": "image_url", "image_url": {"url": image_url}}))

            formatted_messages.append(cast(ChatCompletionUiformMessage, {"role": role, "content": formatted_content}))

    return formatted_messages


def convert_to_openai_format(messages: List[ChatCompletionUiformMessage]) -> List[ChatCompletionMessageParam]:
    return cast(list[ChatCompletionMessageParam], messages)


def convert_from_openai_format(messages: list[ChatCompletionMessageParam]) -> list[ChatCompletionUiformMessage]:
    return cast(list[ChatCompletionUiformMessage], messages)


def separate_messages(
    messages: list[ChatCompletionUiformMessage],
) -> tuple[Optional[ChatCompletionUiformMessage], list[ChatCompletionUiformMessage], list[ChatCompletionUiformMessage]]:
    """
    Separates messages into system, user and assistant messages.

    Args:
        messages: List of chat messages containing system, user and assistant messages

    Returns:
        Tuple containing:
        - The system message if present, otherwise None
        - List of user messages
        - List of assistant messages
    """
    system_message = None
    user_messages = []
    assistant_messages = []

    for message in messages:
        if message["role"] in ("system", "developer"):
            system_message = message
        elif message["role"] == "user":
            user_messages.append(message)
        elif message["role"] == "assistant":
            assistant_messages.append(message)

    return system_message, user_messages, assistant_messages


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
