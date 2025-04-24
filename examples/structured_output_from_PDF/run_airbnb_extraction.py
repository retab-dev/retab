# ---------------------------------------------
## Example using OpenAI with UiForm to extract structured data from a document.
# ---------------------------------------------

import json
import os

from airbnb_schema import PitchDeck
from dotenv import load_dotenv
from openai import OpenAI

from uiform import Schema, UiForm

# Load environment variables
load_dotenv()
api_key = os.getenv("OPENAI_API_KEY")
uiform_api_key = os.getenv("UIFORM_API_KEY")

assert api_key, "Missing OPENAI_API_KEY"
assert uiform_api_key, "Missing UIFORM_API_KEY"

uiclient = UiForm(api_key=uiform_api_key)
doc_msg = uiclient.documents.create_messages(document="../../assets/airbnb_pitch_deck.pdf")

# Define schema
schema_obj = Schema(pydantic_model=PitchDeck)

print("\n\nJSON Schema:")
print(json.dumps(PitchDeck.model_json_schema(), indent=2))

# Now you can use your favorite model to analyze your document
client = OpenAI(api_key=api_key)
completion = client.beta.chat.completions.parse(
    model="gpt-4o-mini", messages=schema_obj.openai_messages + doc_msg.openai_messages, response_format=schema_obj.inference_pydantic_model, store=True
)

assert completion.choices[0].message.parsed is not None
print(json.dumps(completion.choices[0].message.parsed.model_dump(), indent=2))


# ---------------------------------------------
## Eventually: validate the response against the original schema if you want to remove the reasoning fields
# ---------------------------------------------

from uiform._utils.json_schema import filter_reasoning_fields_json

assert completion.choices[0].message.content is not None
extraction = schema_obj.pydantic_model.model_validate(filter_reasoning_fields_json(completion.choices[0].message.content))

print(json.dumps(extraction.model_dump(), indent=2))
