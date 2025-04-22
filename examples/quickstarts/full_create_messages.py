# ---------------------------------------------
## Quick example: Create UiForm messages with OCR and custom image settings.
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

# Message creation with advanced image settings
doc_msg = uiclient.documents.create_messages(
    document="../../assets/calendar_event.xlsx",
    modality="native",
    image_settings={
        "correct_image_orientation": True,
        "dpi": 72,
        "image_to_text": "ocr",
        "browser_canvas": "A4"
    }
)

# Output
print("\n📤 Created Messages:")
print(doc_msg.openai_messages)
