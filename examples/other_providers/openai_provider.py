# ---------------------------------------------
## Example using OpenAI with Retab to extract structured data from a document.
# ---------------------------------------------

import os

from dotenv import load_dotenv
from openai import OpenAI
from pydantic import BaseModel, ConfigDict, Field

from retab import Schema, Retab

# Load environment variables
load_dotenv()
api_key = os.getenv("OPENAI_API_KEY")
retab_api_key = os.getenv("RETAB_API_KEY")

assert api_key, "Missing OPENAI_API_KEY"
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

# OpenAI client
client = OpenAI(api_key=api_key)

# Run extraction with schema
""" 
completion = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=doc_msg.openai_messages + [
        {
            "role": "user",
            "content": "Summarize the document"
        }
    ]
)

# --- OR 

completion = client.chat.completions.create(
    model="gpt-4o-mini",
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

# --- OR  
"""

completion = client.beta.chat.completions.parse(
    model="gpt-4o-mini", messages=schema_obj.openai_messages + doc_msg.openai_messages, response_format=schema_obj.inference_pydantic_model
)

# --- Does not work with beta.chat.completions.parse !

print("\n✅ Extracted Data:")
print(completion.choices[0].message.parsed.model_dump_json(indent=2))
