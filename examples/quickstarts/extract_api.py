# ---------------------------------------------
## Quick example: Extract structured data using UiForm’s all-in-one `.parse()` method.
# ---------------------------------------------

import os
from dotenv import load_dotenv
from uiform import UiForm

# Load environment variables
load_dotenv()

uiform_api_key = os.getenv("UIFORM_API_KEY")
assert uiform_api_key, "Missing UIFORM_API_KEY"

# UiForm Setup
uiclient = UiForm(api_key=uiform_api_key)

# Document Extraction via UiForm API
with open("../../assets/booking_confirmation.jpg", "rb") as f:
    response = uiclient.documents.extractions.parse(
        document=f,
        model="gpt-4o-mini",
        json_schema={
            'X-SystemPrompt': 'You are a useful assistant.',
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
        },
        modality="text"
    )

# Output
print("\n✅ Extracted Result:")
print(response.choices[0].message.content)
