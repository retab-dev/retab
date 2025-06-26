from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.mailboxes.create(
    name="Mailbox Automation",
    processor_id="proc_o4dtLxizT0kDAjeKuyVLA",
    webhook_url="https://your-server.com/webhook",
    email="test@mailbox.retab.dev",
)

print(automation.model_dump_json(indent=2))
