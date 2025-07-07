from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.local")

reclient = Retab()

automation = reclient.processors.automations.endpoints.update(
    endpoint_id="endp_jKH0WW6X1dD3wtWbnuuXQ",
    name="Endpoint Automation Updated",
)

print(automation.model_dump_json(indent=2))
