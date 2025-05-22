# ---------------------------------------------
## Example: Define and use a CalendarEvent schema using JSON Schema (manual)
# ---------------------------------------------

import json
import os

from dotenv import load_dotenv
from openai import OpenAI

from uiform import Schema, UiForm

# Load environment variables
load_dotenv()
api_key = os.getenv("OPENAI_API_KEY")
uiform_api_key = os.getenv("UIFORM_API_KEY")

assert api_key, "Missing OPENAI_API_KEY"
assert uiform_api_key, "Missing UIFORM_API_KEY"

uiclient = UiForm(api_key=uiform_api_key)
doc_msg = uiclient.documents.create_messages(document="../../assets/calendar_event.xlsx")

# Define schema
schema_obj = Schema(
    json_schema={
        'X-SystemPrompt': 'You are a useful assistant extracting information from documents.',
        'properties': {
            'name': {'description': 'The name of the calendar event.', 
                     'title': 'Name', 'type': 'string'},
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
)

# Now you can use your favorite model to analyze your document
client = OpenAI(api_key=api_key)
completion = client.chat.completions.create(
    model="gpt-4o",
    messages=schema_obj.openai_messages + doc_msg.openai_messages,
    response_format={"type": "json_schema", "json_schema": {"name": schema_obj.id, "schema": schema_obj.inference_json_schema, "strict": True}},
)

# Validate the response against the original schema if you want to remove the reasoning fields
from uiform._utils.json_schema import filter_reasoning_fields_json

assert completion.choices[0].message.content is not None
extraction = schema_obj.pydantic_model.model_validate(filter_reasoning_fields_json(completion.choices[0].message.content))

print("\n✅ Extracted Calendar Event:")
print(json.dumps(extraction.model_dump(), indent=2))
