from pydantic import BaseModel, Field

# The API request
class RegexInstruction(BaseModel):
    name: str = Field(description="A key or label for the data being extracted (e.g., 'VATNumber')")
    pattern: str = Field(description="The regex pattern to search for")
    description: str = Field(description="A human-readable explanation of what is being extracted")

class TextOperations(BaseModel):
    regex_instructions: list[RegexInstruction] = Field(default_factory=list, description="Regex-based instructions to identify potential data candidates in the text.")

class RegexInstructionResult(BaseModel):
    instruction: RegexInstruction
    hits: list[str]


    