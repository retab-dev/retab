from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.mailboxes.get(
    mailbox_id="mb_rAfmknvMdIjX4Yj_SxQeE",
)

print(automation.model_dump_json(indent=2))
