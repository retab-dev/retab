from pydantic import BaseModel
from typing import TypedDict

# The API request
class RegexInstruction(TypedDict):
    name: str           # 'A key or label for the data being extracted (e.g., "VATNumber")'
    pattern: str        # 'The regex pattern to search for'
    description: str    # 'A human-readable explanation of what is being extracted'

class TextOperations(TypedDict, total=False):
    regex_instructions: list[RegexInstruction] #  "Regex-based instructions to identify potential data candidates in the text."

class RegexInstructionResult(BaseModel):
    instruction: RegexInstruction
    hits: list[str]