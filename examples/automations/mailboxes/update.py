from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.links.update(
    link_id="lnk_Xf15nXFpo7mwfGT1aSYo4",
    name="Link Automation Updated",
)

print(automation.model_dump_json(indent=2))
