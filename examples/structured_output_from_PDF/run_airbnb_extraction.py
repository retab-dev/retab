# ---------------------------------------------
## Example using OpenAI with Retab to extract structured data from a document.
# ---------------------------------------------

import json
import os

from airbnb_schema import PitchDeck
from dotenv import load_dotenv
from openai import OpenAI

from retab import Schema, Retab

# Load environment variables
load_dotenv()
api_key = os.getenv("OPENAI_API_KEY")
retab_api_key = os.getenv("RETAB_API_KEY")

assert api_key, "Missing OPENAI_API_KEY"
assert retab_api_key, "Missing RETAB_API_KEY"

reclient = Retab(api_key=retab_api_key)
doc_msg = reclient.documents.create_messages(document="../../assets/airbnb_pitch_deck.pdf")

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

from retab._utils.json_schema import filter_auxiliary_fields_json

assert completion.choices[0].message.content is not None
extraction = schema_obj.pydantic_model.model_validate(filter_auxiliary_fields_json(completion.choices[0].message.content))

print(json.dumps(extraction.model_dump(), indent=2))
