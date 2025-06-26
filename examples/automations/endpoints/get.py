from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.endpoints.get(
    endpoint_id="endp_fJVB60EdGO7XXT5k02cvn",
)

print(automation.model_dump_json(indent=2))
