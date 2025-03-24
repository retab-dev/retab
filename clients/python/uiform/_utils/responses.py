import datetime
import json

from jiter import from_json
from openai.types.chat.chat_completion_content_part_image_param import ChatCompletionContentPartImageParam, ImageURL
from openai.types.chat.chat_completion_content_part_param import ChatCompletionContentPartParam
from openai.types.chat.chat_completion_content_part_text_param import ChatCompletionContentPartTextParam
from openai.types.chat.parsed_chat_completion import ParsedChatCompletionMessage
from openai.types.completion_usage import CompletionTokensDetails, CompletionUsage, PromptTokensDetails
from openai.types.responses.easy_input_message_param import EasyInputMessageParam

# from openai.types.responses.response_input_param import ResponseInputParam
from openai.types.responses.response import Response
from openai.types.responses.response_input_image_param import ResponseInputImageParam
from openai.types.responses.response_input_message_content_list_param import ResponseInputMessageContentListParam
from openai.types.responses.response_input_param import ResponseInputItemParam
from openai.types.responses.response_input_text_param import ResponseInputTextParam

from ..types.chat import ChatCompletionUiformMessage
from ..types.documents.extractions import UiParsedChatCompletion, UiParsedChoice


def convert_to_openai_format(messages: list[ChatCompletionUiformMessage]) -> list[ResponseInputItemParam]:
    """
    Converts a list of ChatCompletionUiformMessage to the OpenAI ResponseInputParam format.

    Args:
        messages: List of chat messages in UIForm format

    Returns:
        Messages in OpenAI ResponseInputParam format for the Responses API
    """
    formatted_messages: list[ResponseInputItemParam] = []

    for message in messages:
        role = message["role"]
        content = message["content"]

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
        formatted_message = EasyInputMessageParam(role=role, content=formatted_content, type="message")

        formatted_messages.append(formatted_message)

    return formatted_messages


def convert_from_openai_format(messages: list[ResponseInputItemParam]) -> list[ChatCompletionUiformMessage]:
    """
    Converts messages from OpenAI ResponseInputParam format to ChatCompletionUiformMessage format.

    Args:
        messages: Messages in OpenAI ResponseInputParam format

    Returns:
        List of chat messages in UIForm format
    """
    formatted_messages: list[ChatCompletionUiformMessage] = []

    for message in messages:
        if "type" not in message:
            # The type is required by all other sub-types of ResponseInputItemParam except for EasyInputMessageParam and Message, which are messages.
            message["type"] = "message"

        if message["type"] != "message":
            print(f"Not supported message type: {message['type']}... Skipping...")
            continue
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
        formatted_message = ChatCompletionUiformMessage(role=role, content=formatted_content)
        formatted_messages.append(formatted_message)

    return formatted_messages


def parse_openai_responses_response(response: Response) -> UiParsedChatCompletion:
    """
    Convert an OpenAI Response (Responses API) to UiParsedChatCompletion type.

    Args:
        response: Response from OpenAI Responses API

    Returns:
        Parsed response in UiParsedChatCompletion format
    """
    # Create the UiParsedChatCompletion object
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
        UiParsedChoice(
            index=0,
            message=ParsedChatCompletionMessage(
                role="assistant",
                content=json.dumps(result_object),
            ),
            finish_reason="stop",
        )
    )

    return UiParsedChatCompletion(
        id=response.id,
        choices=choices,
        created=int(datetime.datetime.now().timestamp()),
        model=response.model,
        object="chat.completion",
        likelihoods={},
        usage=usage,
    )
