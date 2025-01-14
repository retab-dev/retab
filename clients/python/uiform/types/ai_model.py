from typing import Literal

AIProvider = Literal["OpenAI", "Anthropic", "xAI", "Gemini"]
GeminiModel = Literal[ "gemini-2.0-flash-exp",
                      "gemini-1.5-flash-8b", "gemini-1.5-flash","gemini-1.5-pro"]
AnthropicModel = Literal["claude-3-5-sonnet-latest","claude-3-5-sonnet-20241022",
                         "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"]
OpenAIModel = Literal["gpt-4o", "gpt-4o-mini","chatgpt-4o-latest",
                      "gpt-4o-2024-11-20", "gpt-4o-2024-08-06", "gpt-4o-2024-05-13",
                      "gpt-4o-mini-2024-07-18",
                      "o1-2024-12-17", "o1-mini-2024-09-12"]
xAI_Model = Literal["grok-2-vision-1212", "grok-2-1212"]
LLMModel = Literal[AnthropicModel, OpenAIModel, xAI_Model, GeminiModel]
