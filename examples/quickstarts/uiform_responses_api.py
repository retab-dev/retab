# ---------------------------------------------
## Full example: Use Retab with Responses API implementation compatible with OpenAI's Responses API
# ---------------------------------------------

import os
from typing import List
from pydantic import BaseModel, Field

from dotenv import load_dotenv

from retab import Retab, Schema
from retab._utils.json_schema import filter_auxiliary_fields_json

# Load environment variables
load_dotenv()

retab_api_key = os.getenv("RETAB_API_KEY")
assert retab_api_key, "Missing RETAB_API_KEY"


# 1. Define schema using Pydantic model for parse() method
class CalendarEvent(BaseModel):
    name: str = Field(description="The name of the calendar event.")
    date: str = Field(description="The date of the calendar event in ISO 8601 format.")


# Define schema for create() method
json_schema = {
    'X-SystemPrompt': 'You are a useful assistant extracting information from documents.',
    'properties': {
        'name': {
            'description': 'The name of the calendar event.',
            'title': 'Name',
            'type': 'string',
        },
        'date': {
            'X-ReasoningPrompt': 'The user can mention it in any format, like **next week** or **tomorrow**. Infer the right date format from the user input.',
            'description': 'The date of the calendar event in ISO 8601 format.',
            'title': 'Date',
            'type': 'string',
        },
    },
    'required': ['name', 'date'],
    'title': 'CalendarEvent',
    'type': 'object',
}

# Configuration
model = "gpt-4o-2024-08-06"
modality = "native"
temperature = 0.0

# Retab Setup
reclient = Retab(api_key=retab_api_key)
doc_msg = reclient.documents.create_messages(
    document="../../assets/calendar_event.xlsx",
    modality=modality,
    image_resolution_dpi=96,
    browser_canvas="A4",
)
schema_obj = Schema(json_schema=json_schema)

# Example 1: Use the Retab's responses.create() method
input_messages = schema_obj.openai_responses_input + doc_msg.openai_responses_input
response = reclient.consensus.responses.create(
    model=model,
    temperature=temperature,
    input=input_messages,
    text={"format": {"type": "json_schema", "name": schema_obj.id, "schema": schema_obj.inference_json_schema, "strict": True}},
)

# Output
print("\n✅ Extracted Result Example 1 (responses.create()):")
print(f"Output text: {response.output_text}")
extraction = schema_obj.pydantic_model.model_validate(filter_auxiliary_fields_json(response.output_text))
print(extraction.model_dump_json(indent=2))

# Example 2: Use the Retab's responses.parse() method with Pydantic model
response = reclient.consensus.responses.parse(
    model=model, temperature=temperature, input=doc_msg.openai_responses_input, text_format=CalendarEvent, instructions="Extract the calendar event information from the document."
)

# Output
print("\n✅ Extracted Result Example 2 (responses.parse()):")
print(f"Output text: {response.output_text}")
extraction = CalendarEvent.model_validate_json(response.output_text)
print(extraction.model_dump_json(indent=2))

# Example 3: Working with plain text input
response = reclient.consensus.responses.create(
    model=model,
    temperature=temperature,
    input="Meeting with John on June 15th at 3pm",
    text={"format": {"type": "json_schema", "name": schema_obj.id, "schema": schema_obj.inference_json_schema, "strict": True}},
)

# Output
print("\n✅ Extracted Result Example 3 (text input):")
print(f"Output text: {response.output_text}")
extraction = schema_obj.pydantic_model.model_validate(filter_auxiliary_fields_json(response.output_text))
print(extraction.model_dump_json(indent=2))
