import base64
import logging
from typing import List, Literal, Optional, Union, cast
import datetime
import json

from jiter import from_json
import requests

from anthropic.types.image_block_param import ImageBlockParam
from anthropic.types.message_param import MessageParam
from anthropic.types.text_block_param import TextBlockParam
from google.genai.types import BlobDict, ContentDict, ContentUnionDict, PartDict  # type: ignore
from openai.types.chat.chat_completion_content_part_input_audio_param import ChatCompletionContentPartInputAudioParam
from openai.types.chat.chat_completion_message_param import ChatCompletionMessageParam
from openai.types.chat.chat_completion_content_part_image_param import ChatCompletionContentPartImageParam, ImageURL
from openai.types.chat.chat_completion_content_part_param import ChatCompletionContentPartParam
from openai.types.chat.chat_completion_content_part_text_param import ChatCompletionContentPartTextParam
from openai.types.chat.parsed_chat_completion import ParsedChatCompletionMessage
from openai.types.completion_usage import CompletionTokensDetails, CompletionUsage, PromptTokensDetails
from openai.types.responses.easy_input_message_param import EasyInputMessageParam
from openai.types.responses.response import Response
from openai.types.responses.response_input_image_param import ResponseInputImageParam
from openai.types.responses.response_input_message_content_list_param import ResponseInputMessageContentListParam
from openai.types.responses.response_input_param import ResponseInputItemParam
from openai.types.responses.response_input_text_param import ResponseInputTextParam

from ...types.chat import ChatCompletionRetabMessage
from ...types.documents.extract import RetabParsedChatCompletion, RetabParsedChoice


MediaType = Literal["image/jpeg", "image/png", "image/gif", "image/webp"]


def convert_to_google_genai_format(messages: List[ChatCompletionRetabMessage]) -> tuple[str, list[ContentUnionDict]]:
    """
    Converts a list of ChatCompletionRetabMessage to a format compatible with the google.genai SDK.


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
        messages (List[ChatCompletionRetabMessage]): List of chat messages.

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
            assert isinstance(message.get("content"), str), "System message content must be a string."
            if system_message != "":
                raise ValueError("Only one system message is allowed per chat.")
            system_message += cast(str, message.get("content", ""))
            continue
        parts: list[PartDict] = []

        message_content = message.get("content")
        if isinstance(message_content, str):
            # Direct string content is treated as the prompt for the SDK
            parts.append(PartDict(text=message_content))
        elif isinstance(message_content, list):
            # Handle structured content
            for part in message_content:
                if part["type"] == "text":
                    parts.append(PartDict(text=part["text"]))
                elif part["type"] == "image_url":
                    url = part["image_url"].get("url", "")  # type: ignore
                    if url.startswith("data:image"):
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


def convert_to_anthropic_format(messages: List[ChatCompletionRetabMessage]) -> tuple[str, List[MessageParam]]:
    """
    Converts a list of ChatCompletionRetabMessage to a format compatible with the Anthropic SDK.

    Args:
        messages (List[ChatCompletionRetabMessage]): List of chat messages.

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
            assert isinstance(message.get("content"), str), "System message content must be a string."
            if system_message != "":
                raise ValueError("Only one system message is allowed per chat.")
            system_message += cast(str, message.get("content", ""))
            continue

        # -----------------------
        # Handle non-system roles
        # -----------------------
        if isinstance(message.get("content"), str):
            # Direct string content is treated as a single text block
            content_blocks.append(
                {
                    "type": "text",
                    "text": cast(str, message.get("content", "")),
                }
            )

        elif isinstance(message.get("content"), list):
            # Handle structured content
            for part in cast(list, message.get("content", [])):
                if part["type"] == "text":
                    part = cast(ChatCompletionContentPartTextParam, part)
                    content_blocks.append(
                        {
                            "type": "text",
                            "text": part["text"],  # type: ignore
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


def convert_from_anthropic_format(messages: list[MessageParam], system_prompt: str) -> list[ChatCompletionRetabMessage]:
    """
    Converts a list of Anthropic MessageParam to a list of ChatCompletionRetabMessage.
    """
    formatted_messages: list[ChatCompletionRetabMessage] = [ChatCompletionRetabMessage(role="developer", content=system_prompt)]

    for message in messages:
        role = message["role"]
        content_blocks = message["content"]

        # Handle different content structures
        if isinstance(content_blocks, list) and len(content_blocks) == 1 and isinstance(content_blocks[0], dict) and content_blocks[0].get("type") == "text":
            # Simple text message
            formatted_messages.append(cast(ChatCompletionRetabMessage, {"role": role, "content": content_blocks[0].get("text", "")}))
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

            formatted_messages.append(cast(ChatCompletionRetabMessage, {"role": role, "content": formatted_content}))

    return formatted_messages


def convert_to_openai_completions_api_format(messages: List[ChatCompletionRetabMessage]) -> List[ChatCompletionMessageParam]:
    return cast(list[ChatCompletionMessageParam], messages)


def convert_from_openai_completions_api_format(messages: list[ChatCompletionMessageParam]) -> list[ChatCompletionRetabMessage]:
    return cast(list[ChatCompletionRetabMessage], messages)


def separate_messages(
    messages: list[ChatCompletionRetabMessage],
) -> tuple[Optional[ChatCompletionRetabMessage], list[ChatCompletionRetabMessage], list[ChatCompletionRetabMessage]]:
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


def str_messages(messages: list[ChatCompletionRetabMessage], max_length: int = 100) -> str:
    """
    Converts a list of chat messages into a string representation with faithfully serialized structure.

    Args:
        messages (list[ChatCompletionRetabMessage]): The list of chat messages.
        max_length (int): Maximum length for content before truncation.

    Returns:
        str: A string representation of the messages with applied truncation.
    """

    def truncate(text: str, max_len: int) -> str:
        """Truncate text to max_len with ellipsis."""
        return text if len(text) <= max_len else f"{text[:max_len]}..."

    serialized: list[ChatCompletionRetabMessage] = []
    for message in messages:
        role = message["role"]
        content = message.get("content")

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


def convert_to_openai_responses_api_format(messages: list[ChatCompletionRetabMessage]) -> list[ResponseInputItemParam]:
    """
    Converts a list of ChatCompletionRetabMessage to the OpenAI ResponseInputParam format.

    Args:
        messages: List of chat messages in UIForm format

    Returns:
        Messages in OpenAI ResponseInputParam format for the Responses API
    """
    formatted_messages: list[ResponseInputItemParam] = []

    for message in messages:
        role = message["role"]
        content = message.get("content")

        # Handle different content formats
        formatted_content: ResponseInputMessageContentListParam = []

        if isinstance(content, str):
            # Simple text content - provide direct string value
            formatted_content.append(ResponseInputTextParam(text=content, type="input_text"))
        elif isinstance(content, list):
            # Content is a list of parts
            for part in content:
                if part["type"] == "text":
                    formatted_content.append(ResponseInputTextParam(text=part["text"], type="input_text"))
                elif part["type"] == "image_url":
                    if "detail" in part["image_url"]:
                        detail = part["image_url"]["detail"]
                    else:
                        detail = "high"
                    formatted_content.append(ResponseInputImageParam(image_url=part["image_url"]["url"], type="input_image", detail=detail))
                else:
                    print(f"Not supported content type: {part['type']}... Skipping...")

        # Create Message structure which is one of the types in ResponseInputItemParam
        role_for_response = role if role in ("user", "assistant", "system", "developer") else "assistant"
        formatted_message = EasyInputMessageParam(role=role_for_response, content=formatted_content, type="message")

        formatted_messages.append(formatted_message)

    return formatted_messages


def convert_from_openai_responses_api_format(messages: list[ResponseInputItemParam]) -> list[ChatCompletionRetabMessage]:
    """
    Converts messages from OpenAI ResponseInputParam format to ChatCompletionRetabMessage format.

    Args:
        messages: Messages in OpenAI ResponseInputParam format

    Returns:
        List of chat messages in UIForm format
    """
    formatted_messages: list[ChatCompletionRetabMessage] = []

    for message in messages:
        if "role" not in message or "content" not in message:
            # Mandatory fields for a message
            if message.get("type") != "message":
                print(f"Not supported message type: {message.get('type')}... Skipping...")
            continue

        role = message["role"]
        content = message["content"]

        if "type" not in message:
            # The type is required by all other sub-types of ResponseInputItemParam except for EasyInputMessageParam and Message, which are messages.
            message["type"] = "message"

        role = message["role"]
        content = message["content"]
        formatted_content: str | list[ChatCompletionContentPartParam]

        if isinstance(content, str):
            formatted_content = content
        else:
            # Handle different content formats
            formatted_content = []
            for part in content:
                if part["type"] == "input_text":
                    formatted_content.append(ChatCompletionContentPartTextParam(text=part["text"], type="text"))
                elif part["type"] == "input_image":
                    image_url = part.get("image_url") or ""
                    image_detail = part.get("detail") or "high"
                    formatted_content.append(ChatCompletionContentPartImageParam(image_url=ImageURL(url=image_url, detail=image_detail), type="image_url"))
                else:
                    print(f"Not supported content type: {part['type']}... Skipping...")

        # Create message in UIForm format
        formatted_message = ChatCompletionRetabMessage(role=role, content=formatted_content)
        formatted_messages.append(formatted_message)

    return formatted_messages


def parse_openai_responses_response(response: Response) -> RetabParsedChatCompletion:
    """
    Convert an OpenAI Response (Responses API) to RetabParsedChatCompletion type.

    Args:
        response: Response from OpenAI Responses API

    Returns:
        Parsed response in RetabParsedChatCompletion format
    """
    # Create the RetabParsedChatCompletion object
    if response.usage:
        usage = CompletionUsage(
            prompt_tokens=response.usage.input_tokens,
            completion_tokens=response.usage.output_tokens,
            total_tokens=response.usage.total_tokens,
            prompt_tokens_details=PromptTokensDetails(
                cached_tokens=response.usage.input_tokens_details.cached_tokens,
            ),
            completion_tokens_details=CompletionTokensDetails(
                reasoning_tokens=response.usage.output_tokens_details.reasoning_tokens,
            ),
        )
    else:
        usage = None

    # Parse the ParsedChoice
    choices = []
    output_text = response.output_text
    result_object = from_json(bytes(output_text, "utf-8"), partial_mode=True)  # Attempt to parse the result even if EOF is reached

    choices.append(
        RetabParsedChoice(
            index=0,
            message=ParsedChatCompletionMessage(
                role="assistant",
                content=json.dumps(result_object),
            ),
            finish_reason="stop",
        )
    )

    return RetabParsedChatCompletion(
        id=response.id,
        choices=choices,
        created=int(datetime.datetime.now().timestamp()),
        model=response.model,
        object="chat.completion",
        likelihoods={},
        usage=usage,
    )
