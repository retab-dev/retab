from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.outlook.get(
    outlook_id="outlook_DDw72RgWgIfKFXGaWXcGu",
)

print(automation.model_dump_json(indent=2))
