# ---------------------------------------------
## Quick example: Summarize a document using UiForm + OpenAI, without a schema.
# ---------------------------------------------

import os

from dotenv import load_dotenv
from uiform import UiForm

# Load environment variables
load_dotenv()

api_key = os.getenv("OPENAI_API_KEY")
uiform_api_key = os.getenv("UIFORM_API_KEY")

assert api_key, "Missing OPENAI_API_KEY"
assert uiform_api_key, "Missing UIFORM_API_KEY"

# UiForm Setup
uiclient = UiForm(api_key=uiform_api_key)
doc_msg = uiclient.documents.create_messages(
    document="../../assets/booking_confirmation.jpg",
    modality="native", 
    image_settings={
        "correct_image_orientation": True, 
        "dpi": 72, 
        "image_to_text": "ocr", 
        "browser_canvas": "A4"
    }
)


# OpenAI Summarization (no structured output)
from openai import OpenAI
client = OpenAI(api_key=api_key)
completion = client.chat.completions.create(
    model="gpt-4o-mini", 
    messages=doc_msg.openai_messages + [{"role": "user", "content": "Summarize the document"}]
)

# Output
print("\n📝 OpenAI Completion Output:")
print(completion.model_dump_json(indent=2))
