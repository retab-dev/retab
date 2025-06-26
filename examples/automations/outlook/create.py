from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.outlook.create(
    name="Outlook Automation",
    processor_id="proc_-AnnMLvJs6JwNc5scIWPq",
    webhook_url="https://your-server.com/webhook",
)

print(automation.model_dump_json(indent=2))
