from typing import Literal, TypedDict, Union, Optional, cast, Required

from openai.types.chat.chat_completion_message_param import ChatCompletionMessageParam
from openai.types.chat.chat_completion_content_part_param import ChatCompletionContentPartParam
from openai.types.chat.chat_completion_message_function_tool_call_param import ChatCompletionMessageFunctionToolCallParam


class ChatCompletionRetabMessage(TypedDict, total=False):  # homemade replacement for ChatCompletionMessageParam because iterable messes the serialization with pydantic
    # Basic stuff
    role: Required[Literal["user", "system", "assistant", "developer", "tool"]]
    content: Union[str, list[ChatCompletionContentPartParam], None]

    # Tool stuff
    tool_call_id: Optional[str]

    # Assistant stuff
    tool_calls: Optional[list[ChatCompletionMessageFunctionToolCallParam]]


def retab_to_openai(messages: list[ChatCompletionRetabMessage]) -> list[ChatCompletionMessageParam]:
    _out = []
    for _m in messages:
        _out_m = {**_m}
        if _m["role"] == "assistant":
            tool_calls = _m.get("tool_calls", None)
            if tool_calls is not None and len(tool_calls) == 0:
                _out_m.pop("tool_calls", None)
        tool_call_id = _m.get("tool_call_id", None)
        if tool_call_id is None:
            _out_m.pop("tool_call_id", None)
        _out.append(cast(ChatCompletionMessageParam, _out_m))
    return _out
