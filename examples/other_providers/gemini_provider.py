# ---------------------------------------------
## Example using Gemini API with Retab to extract structured data from a document.
# ---------------------------------------------

import os

from dotenv import load_dotenv
from openai import OpenAI  # Still using OpenAI interface structure
from pydantic import BaseModel, ConfigDict, Field

from retab import Schema, Retab

# Load environment variables
load_dotenv()
gemini_api_key = os.getenv("GEMINI_API_KEY")
retab_api_key = os.getenv("RETAB_API_KEY")

assert gemini_api_key, "Missing GEMINI_API_KEY"
assert retab_api_key, "Missing RETAB_API_KEY"


# Define schema
class CalendarEvent(BaseModel):
    model_config = ConfigDict(json_schema_extra={"X-SystemPrompt": "You are a useful assistant."})
    name: str = Field(..., description="Event name", json_schema_extra={"X-FieldPrompt": "Provide a descriptive and concise name for the event."})
    date: str = Field(..., description="Date in ISO 8601", json_schema_extra={"X-ReasoningPrompt": "Infer the date from vague formats like 'next week' or 'tomorrow'."})


# Retab setup
reclient = Retab(api_key=retab_api_key)
doc_msg = reclient.documents.create_messages(document="../../assets/booking_confirmation.jpg")
schema_obj = Schema(pydantic_model=CalendarEvent)

# Gemini-compatible OpenAI-like client
client = OpenAI(api_key=gemini_api_key, base_url="https://generativelanguage.googleapis.com/v1beta/openai/")

# Run extraction with schema
completion = client.chat.completions.create(
    model="gemini-1.5-flash",
    messages=schema_obj.openai_messages + doc_msg.openai_messages + [{"role": "user", "content": "Extract information from this document."}],
    response_format={"type": "json_schema", "json_schema": {"name": schema_obj.id, "schema": schema_obj.inference_gemini_json_schema, "strict": True}},
)

print("\n✅ Extracted Data (Gemini):")
print(completion.model_dump_json(indent=2))
