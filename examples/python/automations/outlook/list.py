from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automations = reclient.processors.automations.outlook.list(
    processor_id="proc_o4dtLxizT0kDAjeKuyVLA",
)

print(automations.model_dump_json(indent=2))
