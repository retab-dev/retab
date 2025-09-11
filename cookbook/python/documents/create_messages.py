import os
from dotenv import load_dotenv
from retab import Retab

# ---------------------------------------------
## Quick example: Parse a document in markdown format using Retab.
# ---------------------------------------------

# Load environment variables
load_dotenv()

retab_api_key = os.getenv("RETAB_API_KEY")

assert retab_api_key, "Missing RETAB_API_KEY"

# Retab Setup
client = Retab(api_key=retab_api_key)
result = client.documents.parse(
    document="../../assets/docs/booking_confirmation.jpg",
    model="gemini-2.5-flash",
    image_resolution_dpi=192,
    browser_canvas="A4",
)

print(result.pages)