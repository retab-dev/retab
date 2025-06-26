import os
from dotenv import load_dotenv
from retab import Retab
from openai import OpenAI


# ---------------------------------------------
## Quick example: Summarize a document using Retab + OpenAI, without a schema.
# ---------------------------------------------

# Load environment variables
load_dotenv()

api_key = os.getenv("OPENAI_API_KEY")
retab_api_key = os.getenv("RETAB_API_KEY")

assert api_key, "Missing OPENAI_API_KEY"
assert retab_api_key, "Missing RETAB_API_KEY"

# Retab Setup
reclient = Retab(api_key=retab_api_key)
doc_msg = reclient.documents.create_messages(
    document="../../assets/code/booking_confirmation.jpg",
    modality="native",
    image_resolution_dpi=96,
    browser_canvas="A4",
)


client = OpenAI(api_key=api_key)
completion = client.chat.completions.create(model="gpt-4o-mini", messages=doc_msg.openai_messages + [{"role": "user", "content": "Summarize the document"}])

# Output
print("\n📝 OpenAI Completion Output:")
print(completion.model_dump_json(indent=2))
