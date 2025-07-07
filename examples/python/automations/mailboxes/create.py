from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.mailboxes.create(
    name="Mailbox Automation",
    processor_id="proc_2BTqDkKOTO_Ddz1N7JwVZ",
    webhook_url="https://api.retab.com/test-webhook",
    email="invoices-new@mailbox.retab.com",
)

print(automation.model_dump_json(indent=2))
