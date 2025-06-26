from dotenv import load_dotenv
from retab import Retab
import os

assert load_dotenv("../../../.env.production")

client = Retab()

mailbox = client.processors.automations.mailboxes.create(
    name="Mailbox Automation",
    processor_id="proc_2BTqDkKOTO_Ddz1N7JwVZ",
    webhook_url="http://api.retab.dev/webhook",
    webhook_headers={"API-Key": os.getenv("RETAB_API_KEY")},
    email="invoices-4@mailbox.retab.dev",
)

print(mailbox.model_dump_json(indent=2))

# If you want to test a full email forwarding
log = client.processors.automations.mailboxes.tests.forward(email="invoices-3@mailbox.retab.dev", document="your-invoice-email.eml")

print(log.model_dump_json(indent=2))
