from uiform import UiForm
from openai import OpenAI
import os

uiclient = UiForm()

doc_msg = uiclient.documents.create_messages(document = "freight/booking_confirmation.jpg")

client = OpenAI(base_url="https://api.x.ai/v1", api_key=os.getenv("XAI_API_KEY"))
completion = client.chat.completions.create(
    model="grok-2-vision-1212",
    messages=doc_msg.openai_messages + [
        {
            "role": "user",
            "content": "Summarize the document"
        }
    ]
)

print(completion.model_dump())





# ------------------------------------------------------------------------------------------------
# Structured extraction with OpenAI API

from uiform import Schema

schema_obj = Schema(
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
)

completion = client.chat.completions.create(
    model="grok-2-vision-1212",
    messages=schema_obj.openai_messages + doc_msg.openai_messages,
    response_format={
        "type": "json_schema",
        "json_schema": {
            "name": schema_obj.schema_version,
            "schema": schema_obj.inference_json_schema,
            "strict": True
        }
    }
)

# --- OR 

completion = client.beta.chat.completions.parse(
    model="grok-2-vision-1212",
    messages=schema_obj.openai_messages + doc_msg.openai_messages,
    response_format=schema_obj.inference_pydantic_model
)
