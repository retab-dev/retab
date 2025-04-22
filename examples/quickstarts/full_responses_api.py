# ---------------------------------------------
## Full example: Use UiForm with OpenAI's Responses API and JSON Schema extraction.
# ---------------------------------------------

import os
from dotenv import load_dotenv
from uiform import UiForm, Schema
from openai import OpenAI
from uiform._utils.json_schema import filter_reasoning_fields_json

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
            'X-FieldPrompt': 'Provide a descriptive and concise name for the event.',
            'description': 'The name of the calendar event.',
            'title': 'Name',
            'type': 'string'
        },
        'date': {
            'X-ReasoningPrompt': 'The user can mention it in any format, like **next week** or **tomorrow**. Infer the right date format from the user input.',
            'description': 'The date of the calendar event in ISO 8601 format.',
            'title': 'Date',
            'type': 'string'
        }
    },
    'required': ['name', 'date'],
    'title': 'CalendarEvent',
    'type': 'object'
}

# Optional image preprocessing
image_settings = {
    "correct_image_orientation": True,
    "dpi": 72,
    "image_to_text": "ocr",
    "browser_canvas": "A4"
}

# Configuration
model = "gpt-4o"
modality = "native"
temperature = 0.0

# UiForm Setup
uiclient = UiForm(api_key=uiform_api_key)
doc_msg = uiclient.documents.create_messages(
    document="../../assets/calendar_event.xlsx",
    modality=modality,
    image_settings=image_settings,
)
schema_obj = Schema(json_schema=json_schema)

# OpenAI Responses API call
client = OpenAI(api_key=api_key)
response = client.responses.create(
    model=model,
    temperature=temperature,
    input=schema_obj.openai_responses_input + doc_msg.openai_responses_input,
    text={
        "format": {
            "type": "json_schema",
            "name": schema_obj.id,
            "schema": schema_obj.inference_json_schema,
            "strict": True
        }
    }
)

# Validate response and remove reasoning fields
extraction = schema_obj.pydantic_model.model_validate(
    filter_reasoning_fields_json(response.output_text)
)

# Output
print("\n✅ Extracted Result (Responses API):")
print(extraction.model_dump_json(indent=2))
