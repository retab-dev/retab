from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.local")

reclient = Retab()

print("making a query to create an endpoint")
automation = reclient.processors.automations.endpoints.create(
    name="Endpoint Automation 2",
    processor_id="proc_o4dtLxizT0kDAjeKuyVLA",
    webhook_url="https://your-server.com/webhook",
)

print(automation.model_dump_json(indent=2))
