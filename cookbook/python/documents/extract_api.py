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
client = Retab(api_key=retab_api_key)

# Document Extraction via Retab API
response = client.documents.extract(
    document="../../assets/docs/invoice.jpeg",
    model="gpt-5",
    json_schema="../../assets/code/invoice_schema.json",
    image_resolution_dpi=192,
    browser_canvas="A4",
    temperature=0.0,
    n_consensus=1,
)

# Output
print("\n✅ Extracted Result:")
print(response.choices[0].message.content)
