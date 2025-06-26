from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.endpoints.update(
    endpoint_id="endp_rh-D1lcOuPZj1qqhL3YPc",
    name="Endpoint Automation Updated",
)

print(automation.model_dump_json(indent=2))
