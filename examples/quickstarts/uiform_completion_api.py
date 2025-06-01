# ---------------------------------------------
## Full example: Extract structured data using UiForm + OpenAI Chat Completion + JSON Schema
# ---------------------------------------------

import os

from dotenv import load_dotenv
from openai import OpenAI

from uiform import Schema, UiForm
from uiform._utils.json_schema import filter_auxiliary_fields_json

# Load environment variables
load_dotenv()

api_key = os.getenv("OPENAI_API_KEY")
uiform_api_key = os.getenv("UIFORM_API_KEY")

assert api_key, "Missing OPENAI_API_KEY"
assert uiform_api_key, "Missing UIFORM_API_KEY"

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

# Optional image processing settings
model = "gpt-4o"
modality = "native"
temperature = 0.0
n_consensus = 1

# UiForm Setup
uiclient = UiForm(api_key=uiform_api_key)
doc_msg = uiclient.documents.create_messages(
    document="../../assets/calendar_event.xlsx",
    modality=modality,
    image_resolution_dpi=96,
    browser_canvas="A4",
)

# OpenAI Chat Completion with schema-based prompting
completion = uiclient.completions.create(
    model=model,
    temperature=temperature,
    messages=doc_msg.openai_messages,
    response_format="json_schema",  # You can either put a json_schema, a pydantic model, a file path...
    n_consensus=n_consensus,
)

# Same for parse
completion = uiclient.completions.parse(
    model=model,
    temperature=temperature,
    messages=doc_msg.openai_messages,
    response_format="json_schema",  # You can either put a json_schema, a pydantic model, a file path...
    n_consensus=n_consensus,
)


# Validate and clean the output
assert completion.choices[0].message.content is not None
extraction = schema_obj.pydantic_model.model_validate(filter_auxiliary_fields_json(completion.choices[0].message.content))

# Output
print("\n✅ Extracted Result:")
print(extraction.model_dump_json(indent=2))
