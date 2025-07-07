from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.links.get(
    link_id="lnk_Xf15nXFpo7mwfGT1aSYo4",
)

print(automation.model_dump_json(indent=2))
