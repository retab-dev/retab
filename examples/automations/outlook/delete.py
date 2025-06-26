from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.outlook.get(
    outlook_id="aut_01G34H8J2K",
)

print(automation.model_dump_json(indent=2))
