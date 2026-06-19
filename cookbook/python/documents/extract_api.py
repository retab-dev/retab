# ---------------------------------------------
## Quick example: Extract structured data using `client.extractions.create()`.
# ---------------------------------------------

import json
import os

from dotenv import load_dotenv

from retab import Retab

# Load environment variables
load_dotenv()

retab_api_key = os.getenv("RETAB_API_KEY")
assert retab_api_key, "Missing RETAB_API_KEY"

# Retab Setup
client = Retab(api_key=retab_api_key)

# Document Extraction via Retab API.
# create() takes keyword args directly (no ExtractionRequest wrapper); the
# document may be a path/bytes/URL, but json_schema must be a dict.
response = client.extractions.create(
    document="../../../assets/docs/invoice.jpeg",
    json_schema=json.load(open("../../../assets/code/invoice_schema.json")),
    model="retab-small",
    n_consensus=1,
)

# Output
print("\n✅ Extracted Result:")
print(response.output)
