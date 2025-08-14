from typing import Literal, TypedDict, Union

from openai.types.chat.chat_completion_content_part_param import ChatCompletionContentPartParam

class ChatCompletionRetabMessage(TypedDict):  # homemade replacement for ChatCompletionMessageParam because iterable messes the serialization with pydantic
    role: Literal["user", "system", "assistant", "developer"]
    content: Union[str, list[ChatCompletionContentPartParam]]
