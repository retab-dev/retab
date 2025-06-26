from dotenv import load_dotenv
from retab import Retab
import os

assert load_dotenv("../../../.env.local")

client = Retab()

link = client.processors.automations.links.create(
    name="Link Automation",
    processor_id="proc_o4dtLxizT0kDAjeKuyVLA",
    webhook_url="https://api.retab.dev/webhook",
    webhook_headers={"API-Key": os.getenv("RETAB_API_KEY")},
)


# If you want to test the file processing logic:
log = client.processors.automations.tests.upload(automation_id=link.id, document="your-invoice-email.eml")
