# ---------------------------------------------
## Full example: Extract structured data using Retab + OpenAI Chat Completion + JSON Schema
# ---------------------------------------------

import os

from dotenv import load_dotenv
from openai import OpenAI

from retab import Schema, Retab
from retab.utils.json_schema import filter_auxiliary_fields_json

# Load environment variables
load_dotenv()

api_key = os.getenv("OPENAI_API_KEY")
retab_api_key = os.getenv("RETAB_API_KEY")

assert api_key, "Missing OPENAI_API_KEY"
assert retab_api_key, "Missing RETAB_API_KEY"

# Define schema
json_schema = {
    "X-SystemPrompt": "You are a useful assistant extracting information from documents.",
    "properties": {
        "name": {
            "description": "The name of the calendar event.",
            "title": "Name",
            "type": "string",
        },
        "date": {
            "X-ReasoningPrompt": "The user can mention it in any format, like **next week** or **tomorrow**. Infer the right date format from the user input.",
            "description": "The date of the calendar event in ISO 8601 format.",
            "title": "Date",
            "type": "string",
        },
    },
    "required": ["name", "date"],
    "title": "CalendarEvent",
    "type": "object",
}

# Define model and modality
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

# OpenAI Chat Completion with schema-based prompting
client = OpenAI(api_key=api_key)
# Example 1: Use the `inference_json_schema` parameter
completion = client.chat.completions.create(
    model=model,
    temperature=temperature,
    messages=schema_obj.openai_messages + doc_msg.openai_messages,
    response_format={
        "type": "json_schema",
        "json_schema": {
            "name": schema_obj.id,
            "schema": schema_obj.inference_json_schema,
            "strict": True,
        },
    },
)
# Example 2: Use the `inference_pydantic_model` parameter
completion = client.beta.chat.completions.parse(
    model=model,
    messages=schema_obj.openai_messages + doc_msg.openai_messages,
    response_format=schema_obj.inference_pydantic_model,
)


# Validate and clean the output
assert completion.choices[0].message.content is not None
extraction = schema_obj.pydantic_model.model_validate(filter_auxiliary_fields_json(completion.choices[0].message.content))

# Output
print("\n✅ Extracted Result:")
print(extraction.model_dump_json(indent=2))
