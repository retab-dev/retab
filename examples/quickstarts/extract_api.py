# ---------------------------------------------
## Quick example: Extract structured data using Retab’s all-in-one `.parse()` method.
# ---------------------------------------------

import os

from dotenv import load_dotenv

from retab import Retab

# Load environment variables
load_dotenv()

retab_api_key = os.getenv("RETAB_API_KEY")
assert retab_api_key, "Missing RETAB_API_KEY"

# Retab Setup
reclient = Retab(api_key=retab_api_key)

# Document Extraction via Retab API
with open("../../assets/booking_confirmation.jpg", "rb") as f:
    response = reclient.documents.extractions.parse(
        document=f,
        model="gpt-4o-mini",
        json_schema={
            'X-SystemPrompt': 'You are a useful assistant.',
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
        },
        modality="text",
    )

# Output
print("\n✅ Extracted Result:")
print(response.choices[0].message.content)
