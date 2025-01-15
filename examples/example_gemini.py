import google.generativeai as genai # type: ignore
from uiform import UiForm

uiclient = UiForm()
doc_msg = uiclient.documents.create_messages(
    document = "freight/booking_confirmation.jpg"
)

genai.configure()
model = genai.GenerativeModel("gemini-1.5-flash")
response = model.generate_content(
    doc_msg.gemini_messages + ["Summarize the document"] # type: ignore



)

# ------------------------------------------------------------------------------------------------
# Structured extraction with OpenAI API

from uiform import UiForm, Schema
from openai import OpenAI

uiclient = UiForm()
doc_msg = uiclient.documents.create_messages(
    document = "document_1.xlsx"
)
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

# Now you can use your favorite model to analyze your document
import os
client = OpenAI(
    api_key = os.getenv("GEMINI_API_KEY"),
    base_url = "https://generativelanguage.googleapis.com/v1beta/openai/"
)
completion = client.chat.completions.create(
    model="gemini-1.5-flash",
    messages=schema_obj.openai_messages + doc_msg.openai_messages,
    response_format={
        "type": "json_schema",
        "json_schema": {
            "name": schema_obj.schema_version,
            "schema": schema_obj.inference_gemini_json_schema, # Very important ! 
            # We need to put the inference_gemini_json_schema here because gemini does not allow nullable fields,
            # every field is required and the anyOf is not supported.
            "strict": True
        }
    }
)