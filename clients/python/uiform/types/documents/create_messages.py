import base64
from io import BytesIO
from typing import List, Literal, Dict, Union

import PIL.Image
import requests
from anthropic.types.message_param import MessageParam
from google.genai.types import ContentUnionDict  # type: ignore
from openai.types.chat.chat_completion_message_param import ChatCompletionMessageParam
from openai.types.responses.response_input_param import ResponseInputItemParam
from pydantic import BaseModel, Field, computed_field

from ..._utils.chat import convert_to_anthropic_format, convert_to_google_genai_format, str_messages
from ..._utils.chat import convert_to_openai_format as convert_to_openai_completions_api_format
from ..._utils.responses import convert_to_openai_format as convert_to_openai_responses_api_format
from ..._utils.display import count_text_tokens, count_image_tokens
from ..chat import ChatCompletionUiformMessage
from ..mime import MIMEData
from ..modalities import Modality

MediaType = Literal["image/jpeg", "image/png", "image/gif", "image/webp"]


class TokenCount(BaseModel):
    total_tokens: int = 0
    developer_tokens: int = 0
    user_tokens: int = 0

class DocumentCreateMessageRequest(BaseModel):
    document: MIMEData = Field(description="The document to load.")
    modality: Modality = Field(description="The modality of the document to load.")
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: Literal['A3', 'A4', 'A5'] = Field(default='A4', description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type.")

from typing import Any
class DocumentCreateInputRequest(DocumentCreateMessageRequest):
    json_schema: dict[str, Any] = Field(description="The json schema to use for the document.")




class DocumentMessage(BaseModel):
    id: str = Field(description="A unique identifier for the document loading.")
    object: Literal["document_message"] = Field(default="document_message", description="The type of object being loaded.")
    messages: List[ChatCompletionUiformMessage] = Field(description="A list of messages containing the document content and metadata.")
    created: int = Field(description="The Unix timestamp (in seconds) of when the document was loaded.")
    modality: Modality = Field(description="The modality of the document to load.")

    @computed_field
    def token_count(self) -> TokenCount:
        """Returns the token count for the document message.
        
        This property calculates token usage based on both text and image content
        in the messages using the token counting utilities.
        
        Returns:
            TokenCount: A Pydantic model with total, user, and developer token counts.
        """
        total_tokens = 0
        user_tokens = 0
        developer_tokens = 0
        
        for msg in self.messages:
            role = msg.get("role", "user")
            msg_tokens = 0
            
            if isinstance(msg["content"], str):
                msg_tokens = count_text_tokens(msg["content"])
            elif isinstance(msg["content"], list):
                for content_item in msg["content"]:
                    if isinstance(content_item, str):
                        msg_tokens += count_text_tokens(content_item)
                    elif isinstance(content_item, dict):
                        item_type = content_item.get("type")
                        
                        if item_type == "text" and "text" in content_item:
                            msg_tokens += count_text_tokens(content_item["text"])
                        
                        elif item_type == "image_url" and "image_url" in content_item:
                            image_url = content_item["image_url"]["url"]
                            detail = content_item["image_url"].get("detail", "high")
                            msg_tokens += count_image_tokens(image_url, detail)
            
            # Update total tokens
            total_tokens += msg_tokens
            
            # Update role-specific counts
            assert role in ["user", "developer"], f"Invalid role: {role}"
            if role == "user":
                user_tokens += msg_tokens
            elif role == "developer":
                developer_tokens += msg_tokens
        
        return TokenCount(
            total_tokens=total_tokens,
            user_tokens=user_tokens,
            developer_tokens=developer_tokens
        )

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
            if isinstance(msg["content"], str):
                results.append(msg["content"])
                continue
            assert isinstance(msg["content"], list), "content must be a list or a string"
            for content_item in msg["content"]:
                if isinstance(content_item, str):
                    results.append(content_item)
                else:
                    item_type = content_item.get("type")
                    # If item is an image
                    if item_type == "image_url":
                        assert "image_url" in content_item, "image_url is required in ChatCompletionContentPartImageParam"
                        image_data_url = content_item["image_url"]["url"]  # type: ignore

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
                            audio_data = content_item["input_audio"]["data"]  # type: ignore
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
        return convert_to_openai_completions_api_format(self.messages)

    @property
    def openai_responses_input(self) -> list[ResponseInputItemParam]:
        """Returns the messages formatted for OpenAI's Responses API.

        Converts the internal message format to OpenAI's expected format for
        responses.

        Returns:
            list[ResponseInputItemParam]: Messages formatted for OpenAI's responses API.
        """
        return convert_to_openai_responses_api_format(self.messages)

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
    def gemini_messages(self) -> list[ContentUnionDict]:
        """Returns the messages formatted for Google's Gemini API.

        Converts the internal message format to Gemini's expected format,
        handling various content types including text and images.

        Returns:
            list[PartDict]: Messages formatted for Gemini's API.
        """
        return convert_to_google_genai_format(self.messages)[1]

    def __str__(self) -> str:
        return f"DocumentMessage(id={self.id}, object={self.object}, created={self.created}, messages={str_messages(self.messages)}, modality={self.modality})"

    def __repr__(self) -> str:
        return f"DocumentMessage(id={self.id}, object={self.object}, created={self.created}, messages={str_messages(self.messages)}, modality={self.modality})"
