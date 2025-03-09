from uiform import UiForm, Schema
from openai import OpenAI


# ---------------------------------------------
## Variables DEFINITION
# ---------------------------------------------
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

image_settings = {
    "correct_image_orientation": True,
    "dpi": 72,
    "image_to_text": "ocr",
    "browser_canvas": "A4"
}

model = "gpt-4o"
modality = "native"
temperature = 0.0


# ---------------------------------------------
# ---------------------------------------------

uiclient = UiForm()
doc_msg = uiclient.documents.create_messages(
    document = "document_1.xlsx",
    modality = modality,
    image_settings = image_settings,
)
schema_obj = Schema(
    json_schema = json_schema,
)

# Now you can use your favorite model to analyze your document
client = OpenAI()
completion = client.chat.completions.create(
    model = model,
    temperature = temperature,
    messages=schema_obj.openai_messages + doc_msg.openai_messages,
    response_format={
        "type": "json_schema",
        "json_schema": {
            "name": schema_obj.id,
            "schema": schema_obj.inference_json_schema,
            "strict": True
        }
    }
)

# Validate the response against the original schema if you want to remove the reasoning fields
from uiform._utils.json_schema import filter_reasoning_fields_json
assert completion.choices[0].message.content is not None
extraction = schema_obj.pydantic_model.model_validate(
    filter_reasoning_fields_json(completion.choices[0].message.content)
)

print("Result:",extraction)
