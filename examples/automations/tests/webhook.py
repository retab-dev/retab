from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.local")

client = Retab()

link = client.processors.automations.links.create(
    name="Link Automation 4",
    processor_id="proc_2BTqDkKOTO_Ddz1N7JwVZ",
    webhook_url="http://localhost:4000/test-webhook",
)

# If you just want to send a test request to your webhook
log = client.processors.automations.tests.webhook(
    automation_id=link.id,
)

print(log.model_dump_json(indent=2))
