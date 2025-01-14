import outlines
from uiform import Schema

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

schema_obj = Schema(
    json_schema = json_schema
)

model = outlines.models.transformers("microsoft/Phi-3-mini-4k-instruct")
generator = outlines.generate.json(model, response_format_json)
character = generator("Give me a character description")

extraction = schema_obj.pydantic_model.model_validate(
    character.model_dump()
)

print("Result:",extraction)
