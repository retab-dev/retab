# ---------------------------------------------
## Example: Define and use a CalendarEvent schema using Pydantic (recommended for Python devs)
# ---------------------------------------------

import os

from dotenv import load_dotenv
from openai import OpenAI
from pydantic import BaseModel, ConfigDict, Field

from retab import Schema, Retab

# Load environment variables
load_dotenv()
api_key = os.getenv("OPENAI_API_KEY")
retab_api_key = os.getenv("RETAB_API_KEY")

assert api_key, "Missing OPENAI_API_KEY"
assert retab_api_key, "Missing RETAB_API_KEY"

reclient = Retab(api_key=retab_api_key)
doc_msg = reclient.documents.create_messages(document="../../assets/calendar_event.xlsx")


# Define schema
class CalendarEvent(BaseModel):
    model_config = ConfigDict(json_schema_extra={"X-SystemPrompt": "You are a useful assistant."})

    name: str = Field(..., description="The name of the calendar event.")
    date: str = Field(
        ...,
        description="The date of the calendar event in ISO 8601 format.",
        json_schema_extra={
            "X-ReasoningPrompt": "The user can mention it in any format, like **next week** or **tomorrow**. Infer the right date format from the user input.",
        },
    )


schema_obj = Schema(pydantic_model=CalendarEvent)

# Now you can use your favorite model to analyze your document
client = OpenAI(api_key=api_key)
completion = client.beta.chat.completions.parse(model="gpt-4o", messages=schema_obj.openai_messages + doc_msg.openai_messages, response_format=schema_obj.inference_pydantic_model)

# Validate the response against the original schema if you want to remove the reasoning fields
