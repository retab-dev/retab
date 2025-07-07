from dotenv import load_dotenv
from retab import Retab
import os

assert load_dotenv("../../../.env.local")

client = Retab()

# If you want to test the file processing logic:
log = client.processors.automations.tests.upload(automation_id="lnk_DwHVXFUVLgIkxXQaxvx2s", document="your-invoice-email.eml")

print(log.model_dump_json(indent=2))
