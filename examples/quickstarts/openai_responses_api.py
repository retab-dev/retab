# ---------------------------------------------
## Full example: Use Retab with OpenAI's Responses API and JSON Schema extraction.
# ---------------------------------------------

import os

from dotenv import load_dotenv
from openai import OpenAI

from retab import Schema, Retab
from retab._utils.json_schema import filter_auxiliary_fields_json

# Load environment variables
load_dotenv()

api_key = os.getenv("OPENAI_API_KEY")
retab_api_key = os.getenv("RETAB_API_KEY")

assert api_key, "Missing OPENAI_API_KEY"
assert retab_api_key, "Missing RETAB_API_KEY"

# Define schema
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
model = "gpt-4o"
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

# OpenAI Responses API call
client = OpenAI(api_key=api_key)
# Example 1: With the client.responses.create() method
response = client.responses.create(
    model=model,
    temperature=temperature,
    input=schema_obj.openai_responses_input + doc_msg.openai_responses_input,
    text={"format": {"type": "json_schema", "name": schema_obj.id, "schema": schema_obj.inference_json_schema, "strict": True}},
)

# Example 2: With the client.responses.parse() method
client = OpenAI(api_key=api_key)
response = client.responses.parse(
    model=model,
    temperature=temperature,
    input=schema_obj.openai_responses_input + doc_msg.openai_responses_input,
    text_format=schema_obj.inference_pydantic_model,
)

# Example 3: Put the system prompt in the instructions parameter
client = OpenAI(api_key=api_key)
response = client.responses.parse(
    model=model,
    temperature=temperature,
    instructions=schema_obj.developer_system_prompt,
    input=doc_msg.openai_responses_input,
    text_format=schema_obj.inference_pydantic_model,
)


# Validate response and remove reasoning fields
extraction = schema_obj.pydantic_model.model_validate(filter_auxiliary_fields_json(response.output_text))

# Output
print("\n✅ Extracted Result (Responses API):")
print(extraction.model_dump_json(indent=2))
