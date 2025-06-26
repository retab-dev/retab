from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.local")

reclient = Retab()

automation = reclient.processors.automations.endpoints.get(
    endpoint_id="endp_jKH0WW6X1dD3wtWbnuuXQ",
)

print(automation.model_dump_json(indent=2))
